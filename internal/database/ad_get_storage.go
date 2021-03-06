package database

import (
	"fmt"
	"github.com/go-vk-api/vk"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
)

const (
	// constants
	LS       = "ls"
	Comments = "comments"
	Other    = "other"

	// get ad query
	GetAdById = "SELECT a.ad_id, u.vk_id, u.name, u.surname, u.photo_url, a.header, a.text, a.region," +
		" a.district, a.ad_type, a.ls_enabled, a.comments_enabled, a.extra_enabled, " +
		"a.extra_field, a.creation_datetime, a.lat, a.long, a.status," +
		" a.category, a.subcat_list, a.subcat, a.comments_count, aw.views_count, a.hidden, a.subscribers_count, a.metro," +
		" a.full_adress, u.last_change_time FROM ad a JOIN users u ON (a.author_id = u.vk_id) " +
		"JOIN ad_view aw ON (a.ad_id = aw.ad_id) WHERE a.ad_id = $1"

	// get ads query
	GetAds = "SELECT a.ad_id, u.vk_id, u.name, u.surname, u.photo_url, a.header, a.region," +
		" a.district, a.ad_type, a.ls_enabled, a.comments_enabled, a.extra_enabled, " +
		"a.extra_field, a.creation_datetime, a.status," +
		" a.category, a.subcat_list, a.subcat, a.comments_count, a.hidden, a.metro, u.last_change_time" +
		" FROM ad a JOIN users u ON (a.author_id = u.vk_id) " +
		"JOIN (SELECT ad_id FROM ad%s ORDER BY %s LIMIT $%d OFFSET $%d) l ON (l.ad_id = a.ad_id) ORDER BY %s"
	And              = "AND"
	Where            = " WHERE "
	CategoryClause   = " category = $%d "
	AuthorClause     = " author_id = $%d "
	RegionClause     = " region = $%d "
	DistrictClause   = " district = $%d "
	SubCatListClause = " subcat_list = $%d "
	SubCatClause     = " subcat = $%d "
	RadiusClause     = " ST_DWithin(geo_position, ST_SetSRID(ST_MakePoint($%d, $%d), 4326), $%d) "
	QueryClause      = " fts @@ to_tsquery('ru', $%d)"
	GetAdPhotos      = "SELECT ad_photos_id, photo_url FROM ad_photos WHERE ad_id = $1"

	ViewAd = "INSERT INTO ad_view (ad_id, views_count) VALUES ($1, 1)" +
		" ON CONFLICT (ad_id) DO UPDATE SET views_count = ad_view.views_count + 1"
)

func (db *DB) GetAd(adId int, userId int, MinutesAntiFlood int64, maxViews int, client *vk.Client,
	allowedDuration time.Duration) (models.AdForUsersDetailed, int) {
	row := db.db.QueryRow(GetAdById, adId)
	ad := models.AdForUsersDetailed{}
	ad.GeoPosition = new(models.GeoPosition)
	ad.Author = new(models.User)
	extraFieldTry := pgx.NullString{}
	lat := pgx.NullFloat64{}
	long := pgx.NullFloat64{}
	timeStamp := time.Time{}
	metro := pgx.NullString{}
	fullAdress := pgx.NullString{}
	subcatList := pgx.NullString{}
	subcat := pgx.NullString{}
	lastUpdated := time.Time{}
	err := row.Scan(&ad.AdId, &ad.Author.VkId, &ad.Author.Name, &ad.Author.Surname,
		&ad.Author.PhotoUrl, &ad.Header, &ad.Text, &ad.Region, &ad.District, &ad.AdType,
		&ad.LSEnabled, &ad.CommentsEnabled, &ad.ExtraEnabled,
		&extraFieldTry, &timeStamp, &lat, &long, &ad.Status, &ad.Category, &subcatList, &subcat,
		&ad.CommentsCount, &ad.ViewsCount, &ad.Hidden, &ad.SubscribersNum, &metro, &fullAdress, &lastUpdated)
	if err == pgx.ErrNoRows {
		return ad, EMPTY_RESULT
	}
	if err != nil {
		log.Println(err.Error())
		return ad, DB_ERROR
	}
	db.ViewAd(userId, adId, int(MinutesAntiFlood), maxViews)

	ad.GeoPosition.Available = true
	loc, _ := time.LoadLocation("UTC")
	timeStamp.In(loc)
	ad.CreationDate = timeStamp.Format("02 Jan 06 15:04 UTC")
	if extraFieldTry.Valid {
		ad.ExtraField = extraFieldTry.String
	}
	if lat.Valid && long.Valid {
		ad.GeoPosition.Latitude = lat.Float64
		ad.GeoPosition.Longitude = long.Float64
	} else {
		ad.GeoPosition = nil
	}
	if metro.Valid {
		ad.Metro = metro.String
	}
	if fullAdress.Valid {
		ad.FullAdress = fullAdress.String
	}

	if subcatList.Valid {
		ad.SubCatList = subcatList.String
	}

	if subcat.Valid {
		ad.SubCat = subcat.String
	}

	ad.PathesToPhoto, err = db.GetAdPhotos(ad.AdId)
	if err != nil {
		return ad, DB_ERROR
	}
	err = db.db.QueryRow(CheckIfSubscriber, adId, userId).Scan(&ad.IsSubscriber)
	if err != nil {
		return ad, DB_ERROR
	}
	userCache, err := db.CheckVKUserCacheConsist(client, ad.Author.VkId, lastUpdated, allowedDuration)
	if err == nil {
		ad.Author.Name = userCache.Name
		ad.Author.Surname = userCache.Surname
		ad.Author.PhotoUrl = userCache.PhotoURL
	}
	return ad, FOUND
}

func (db *DB) GetAds(page int, rowsPerPage int, params map[string][]string, userId int,
	client *vk.Client, allowedDuration time.Duration) ([]models.AdForUsers, int) {
	offset := rowsPerPage * (page - 1)
	query := GetAds
	whereClause := ""
	innerSortByClause := "ad_id DESC"
	outerSortByClause := "ad_id DESC"
	strArr := make([]interface{}, 0)
	categoryArr, ok := params["category"]
	if ok && len(categoryArr) == 1 {
		strArr = append(strArr, categoryArr[0])
		whereClause += Where + fmt.Sprintf(CategoryClause, 1)
	}
	isQueryInReq := false
	queryPos := -1
	queryArr, ok := params["query"]
	if ok && len(queryArr) == 1 {
		f := func(c rune) bool {
			return !unicode.IsLetter(c) && !unicode.IsNumber(c)
		}
		lexems := strings.FieldsFunc(queryArr[0], f)
		if len(lexems) != 0 {
			lexems[len(lexems)-1] += ":*"
			query = strings.Join(lexems, "&")
			if len(strArr) == 0 {
				whereClause += Where + fmt.Sprintf(QueryClause, 1)
			} else {
				whereClause += And + fmt.Sprintf(QueryClause, len(strArr)+1)
			}
			strArr = append(strArr, query)
			queryPos = len(strArr)
			isQueryInReq = true
		}

	}
	authorArr, ok := params["author_id"]
	authorInQuery := false
	if ok && len(authorArr) == 1 {
		authorInQuery = true
		if len(strArr) == 0 {
			whereClause += Where + fmt.Sprintf(AuthorClause, 1)
		} else {
			whereClause += And + fmt.Sprintf(AuthorClause, len(strArr)+1)
		}
		strArr = append(strArr, authorArr[0])
	}

	regionArr, ok := params["region"]
	if ok && len(regionArr) == 1 {
		if len(strArr) == 0 {
			whereClause += Where + fmt.Sprintf(RegionClause, 1)
		} else {
			whereClause += And + fmt.Sprintf(RegionClause, len(strArr)+1)
		}
		strArr = append(strArr, regionArr[0])
	}

	districtArr, ok := params["district"]
	if ok && len(districtArr) == 1 {
		if len(strArr) == 0 {
			whereClause += Where + fmt.Sprintf(DistrictClause, 1)
		} else {
			whereClause += And + fmt.Sprintf(DistrictClause, len(strArr)+1)
		}
		strArr = append(strArr, districtArr[0])
	}

	subCatListArr, ok := params["subcat_list"]
	if ok && len(subCatListArr) == 1 {
		if len(strArr) == 0 {
			whereClause += Where + fmt.Sprintf(SubCatListClause, 1)
		} else {
			whereClause += And + fmt.Sprintf(SubCatListClause, len(strArr)+1)
		}
		strArr = append(strArr, subCatListArr[0])
	}

	subCatArr, ok := params["subcat"]
	if ok && len(subCatArr) == 1 {
		if len(strArr) == 0 {
			whereClause += Where + fmt.Sprintf(SubCatClause, 1)
		} else {
			whereClause += And + fmt.Sprintf(SubCatClause, len(strArr)+1)
		}
		strArr = append(strArr, subCatArr[0])
	}

	radiusArr, ok := params["radius"]
	if ok && len(radiusArr) == 1 {
		radius, err := strconv.ParseFloat(radiusArr[0], 64)
		if err != nil {
			return nil, DB_ERROR
		}
		latArr, ok := params["lat"]
		if !ok || len(latArr) != 1 {
			return []models.AdForUsers{}, WRONG_INPUT
		}
		lat, err := strconv.ParseFloat(latArr[0], 64)
		if err != nil {
			return nil, DB_ERROR
		}
		longArr, ok := params["long"]
		if !ok || len(longArr) != 1 {
			return []models.AdForUsers{}, WRONG_INPUT
		}
		long, err := strconv.ParseFloat(longArr[0], 64)
		if err != nil {
			return nil, DB_ERROR
		}

		if len(strArr) == 0 {
			whereClause += Where + fmt.Sprintf(RadiusClause, 1, 2, 3)
		} else {
			whereClause += And + fmt.Sprintf(RadiusClause, len(strArr)+1, len(strArr)+2, len(strArr)+3)
		}
		strArr = append(strArr, lat, long, radius*1000)
	}

	sortByArr, ok := params["sort_by"]
	sortArgsLen := 0
	if ok && len(sortByArr) == 1 {
		if sortByArr[0] == "time" {
			log.Println("got order by time")
			innerSortByClause = "creation_datetime DESC"
			outerSortByClause = innerSortByClause
		} else if sortByArr[0] == "geo" {
			latArr, ok := params["lat"]
			if !ok || len(latArr) != 1 {
				return nil, WRONG_INPUT
			}

			longArr, ok := params["long"]
			if !ok || len(longArr) != 1 {
				return nil, WRONG_INPUT
			}
			lat, err := strconv.ParseFloat(latArr[0], 64)
			if err != nil {
				return nil, WRONG_INPUT
			}
			long, err := strconv.ParseFloat(longArr[0], 64)
			if err != nil {
				return nil, WRONG_INPUT
			}
			innerSortByClause = fmt.Sprintf("geo_position <-> ST_SetSRID(ST_POINT($%d, $%d), 4326)",
				len(strArr)+1, len(strArr)+2)
			outerSortByClause = fmt.Sprintf("a.geo_position <-> ST_SetSRID(ST_POINT($%d, $%d), 4326)",
				len(strArr)+1, len(strArr)+2)
			strArr = append(strArr, lat, long)
			sortArgsLen = 2
			//perform some sort by distance(ad geo, given geo)
		}
		//queryArr, ok := params["query"]
		//if ok && len(queryArr) == 1 {
		//	query := queryArr[0]
		//	query = strings.Replace(query, " ", "&", -1)
	}
	var admin = false
	for _, id := range WHITE_LIST {
		if userId == id {
			admin = true
		}
	}

	if isQueryInReq {
		innerSortByClause += fmt.Sprintf(", ts_rank(fts, to_tsquery('ru', $%d)) ", queryPos)
	}
	if !authorInQuery {
		// it's a minimal disjunctive normal form for the "if show" function
		showClose := fmt.Sprintf("(status != 'closed' AND status != 'aborted' AND author_id = $%d OR status='offer' AND hidden = false) ",
			len(strArr)+1)

		if len(strArr)-sortArgsLen == 0 {
			whereClause += Where + showClose
		} else {
			whereClause += And + showClose
		}
		if admin {
			whereClause = ""
			strArr = nil
			innerSortByClause = "hidden DESC, " + innerSortByClause
			outerSortByClause = innerSortByClause
		} else {
			//whereClause += And + fmt.Sprintf(" (hidden = false OR author_id = $%d)", len(strArr)+1)
			strArr = append(strArr, userId)
		}
	} else {
		if !admin {
			whereClause += And + fmt.Sprintf(" (hidden = false OR author_id = $%d)", len(strArr)+1)
			strArr = append(strArr, authorArr[0])
		}
	}

	query = fmt.Sprintf(GetAds, whereClause, innerSortByClause, len(strArr)+1, len(strArr)+2, outerSortByClause)
	strArr = append(strArr, rowsPerPage, offset)
	ads := make([]models.AdForUsers, 0)
	rows, err := db.db.Query(query, strArr...)
	if err != nil {
		log.Println("err is", err)
	}
	if err == pgx.ErrNoRows {
		return nil, EMPTY_RESULT
	}
	if err != nil {
		return nil, DB_ERROR
	}
	defer rows.Close()
	for rows.Next() {
		ads, err = db.WorkWithOneAd(rows, ads, client, allowedDuration)
		if err != nil {
			return nil, DB_ERROR
		}
	}
	if len(ads) == 0 {
		return nil, EMPTY_RESULT
	}
	return ads, FOUND
}

type Ads []models.AdForUsers

func (db *DB) WorkWithOneAd(rows *pgx.Rows, ads Ads, client *vk.Client, allowedDuration time.Duration) (Ads, error) {
	ad := new(models.AdForUsers)
	ad.Author = new(models.User)
	//ad.GeoPosition = new(models.GeoPosition)
	extraFieldTry := pgx.NullString{}
	/*lat := pgx.NullFloat64{}
	long := pgx.NullFloat64{}*/
	metro := pgx.NullString{}
	subcatList := pgx.NullString{}
	subcat := pgx.NullString{}
	timeStamp := time.Time{}

	lastUpdated := time.Time{}
	err := rows.Scan(&ad.AdId, &ad.Author.VkId, &ad.Author.Name, &ad.Author.Surname,
		&ad.Author.PhotoUrl, &ad.Header /*&ad.Text,*/, &ad.Region, &ad.District, &ad.AdType,
		&ad.LSEnabled, &ad.CommentsEnabled, &ad.ExtraEnabled,
		&extraFieldTry, &timeStamp /*&lat, &long,*/, &ad.Status, &ad.Category, &subcatList, &subcat,
		&ad.CommentsCount, &ad.Hidden, &metro, &lastUpdated)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("UTC")
	timeStamp.In(loc)
	ad.CreationDate = timeStamp.Format("02 Jan 06 15:04 UTC")
	if extraFieldTry.Valid {
		ad.ExtraField = extraFieldTry.String
	}
	if metro.Valid {
		ad.Metro = metro.String
	}

	if subcatList.Valid {
		ad.SubCatList = subcatList.String
	}

	if subcat.Valid {
		ad.SubCat = subcat.String
	}

	ad.PathesToPhoto, err = db.GetAdPhotos(ad.AdId)
	if err != nil {
		return nil, err
	}
	if client != nil {
		userCache, err := db.CheckVKUserCacheConsist(client, ad.Author.VkId, lastUpdated, allowedDuration)
		if err == nil {
			ad.Author.Name = userCache.Name
			ad.Author.Surname = userCache.Surname
			ad.Author.PhotoUrl = userCache.PhotoURL
		}
	}

	ads = append(ads, *ad)
	return ads, nil
}

func (db *DB) GetAdPhotos(adId int64) ([]models.AdPhoto, error) {
	photosRows, err := db.db.Query(GetAdPhotos, adId)
	if err != nil {
		return nil, err
	}
	photoArr := make([]models.AdPhoto, 0)
	defer photosRows.Close()
	for photosRows.Next() {
		adPhoto := models.AdPhoto{}
		err = photosRows.Scan(&adPhoto.AdPhotoId, &adPhoto.PhotoUrl)
		if err != nil {
			return nil, err
		}
		photoArr = append(photoArr, adPhoto)
	}
	return photoArr, nil
}

func (db *DB) ViewAd(AuthorId, AdId, MinutesAntiFlood, maxViews int) {
	now := time.Now()
	_, ok := db.AntiFloodAdMap[AdId]
	if !ok {
		db.AntiFloodAdMap[AdId] = make(map[int][]time.Time, 1)
	}
	_, ok = db.AntiFloodAdMap[AdId][AuthorId]
	if !ok {
		db.AntiFloodAdMap[AdId][AuthorId] = make([]time.Time, 1)
		db.AntiFloodAdMap[AdId][AuthorId][0] = time.Now()
	} else {
		n := 0
		// filter slice in place
		for _, x := range db.AntiFloodAdMap[AdId][AuthorId] {
			if now.Sub(x) <= time.Duration(MinutesAntiFlood) * time.Minute {
				db.AntiFloodAdMap[AdId][AuthorId][n] = x
				n++
			}
		}
		db.AntiFloodAdMap[AdId][AuthorId] = db.AntiFloodAdMap[AdId][AuthorId][:n]

		// add new request time
		db.AntiFloodAdMap[AdId][AuthorId] = append(db.AntiFloodAdMap[AdId][AuthorId], now)
		if len(db.AntiFloodAdMap[AdId][AuthorId]) > maxViews {
			return
		}
	}

	userId := 0
	err := db.db.QueryRow(checkAdExist, AdId).Scan(&userId)
	if err != nil {
		log.Println("view ad error: ", err)
		return
	}
	for _, id := range WHITE_LIST {
		if AuthorId == id {
			return
		}
	}
	if userId == AuthorId {
		return
	}
	_, err = db.db.Exec(ViewAd, AdId)
	if err != nil {
		log.Println(err)
	}
}
