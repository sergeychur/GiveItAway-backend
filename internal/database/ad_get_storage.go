package database

import (
	"fmt"
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"log"
	"strconv"
	"time"
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
		" a.category, a.comments_count, aw.views_count, a.hidden, a.subscribers_count FROM ad a JOIN users u ON (a.author_id = u.vk_id) " +
		"JOIN ad_view aw ON (a.ad_id = aw.ad_id) WHERE a.ad_id = $1"

	// get ads query
	GetAds = "SELECT a.ad_id, u.vk_id, u.name, u.surname, u.photo_url, a.header, a.region," +
		" a.district, a.ad_type, a.ls_enabled, a.comments_enabled, a.extra_enabled, " +
		"a.extra_field, a.creation_datetime, a.status," +
		" a.category, a.comments_count, a.hidden FROM ad a JOIN users u ON (a.author_id = u.vk_id) " +
		"JOIN (SELECT ad_id FROM ad%s ORDER BY %s LIMIT $%d OFFSET $%d) l ON (l.ad_id = a.ad_id) ORDER BY %s"
	And            = "AND"
	Where          = " WHERE "
	CategoryClause = " category = $%d "
	AuthorClause   = " author_id = $%d "
	RegionClause   = " region = $%d "
	DistrictClause = " district = $%d "
	GetAdPhotos    = "SELECT ad_photos_id, photo_url FROM ad_photos WHERE ad_id = $1"

	ViewAd = "INSERT INTO ad_view (ad_id, views_count) VALUES ($1, 1)" +
		" ON CONFLICT (ad_id) DO UPDATE SET views_count = ad_view.views_count + 1"
)

func (db *DB) GetAd(adId int, userId int) (models.AdForUsersDetailed, int) {
	row := db.db.QueryRow(GetAdById, adId)
	ad := models.AdForUsersDetailed{}
	ad.GeoPosition = new(models.GeoPosition)
	ad.Author = new(models.User)
	extraFieldTry := pgx.NullString{}
	lat := pgx.NullFloat64{}
	long := pgx.NullFloat64{}
	timeStamp := time.Time{}
	err := row.Scan(&ad.AdId, &ad.Author.VkId, &ad.Author.Name, &ad.Author.Surname,
		&ad.Author.PhotoUrl, &ad.Header, &ad.Text, &ad.Region, &ad.District, &ad.AdType,
		&ad.LSEnabled, &ad.CommentsEnabled, &ad.ExtraEnabled,
		&extraFieldTry, &timeStamp, &lat, &long, &ad.Status, &ad.Category,
		&ad.CommentsCount, &ad.ViewsCount, &ad.Hidden, &ad.SubscribersNum)
	if err == pgx.ErrNoRows {
		return ad, EMPTY_RESULT
	}
	if err != nil {
		log.Println(err.Error())
		return ad, DB_ERROR
	}
	_, err = db.db.Exec(ViewAd, adId)
	if err != nil {
		return models.AdForUsersDetailed{}, DB_ERROR
	}
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

	ad.PathesToPhoto, err = db.GetAdPhotos(ad.AdId)
	if err != nil {
		return ad, DB_ERROR
	}
	return ad, FOUND
}

func (db *DB) FindAds(query string, page int, rowsPerPage int, params map[string][]string, userId int) ([]models.AdForUsers, int) {
	panic("not implemented")
}

func (db *DB) GetAds(page int, rowsPerPage int, params map[string][]string, userId int) ([]models.AdForUsers, int) {
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

	sortByArr, ok := params["sort_by"]
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
			//perform some sort by distance(ad geo, given geo)
		}
	}
	if !authorInQuery {
		// it's a minimal disjunctive normal form for the "if show" function
		showClose := fmt.Sprintf("(status != 'closed' AND author_id = $%d OR status='offer' AND hidden = false) ",
			len(strArr) + 1)
		if len(strArr) == 0 {
			whereClause += Where + showClose
		} else {
			whereClause += And + showClose
		}
		whereClause += And + fmt.Sprintf(" (hidden = false OR author_id = $%d)", len(strArr)+1)
		strArr = append(strArr, userId)
	} else {
		whereClause += And + fmt.Sprintf(" (hidden = false OR author_id = $%d)", len(strArr)+1)
	}

	query = fmt.Sprintf(GetAds, whereClause, innerSortByClause, len(strArr)+1, len(strArr)+2, outerSortByClause)
	strArr = append(strArr, rowsPerPage, offset)
	ads := make([]models.AdForUsers, 0)
	rows, err := db.db.Query(query, strArr...)
	if err == pgx.ErrNoRows {
		return nil, EMPTY_RESULT
	}
	if err != nil {
		return nil, DB_ERROR
	}
	defer rows.Close()
	for rows.Next() {
		ads, err = db.WorkWithOneAd(rows, ads)
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

func (db *DB) WorkWithOneAd(rows *pgx.Rows, ads Ads) (Ads, error) {
	ad := new(models.AdForUsers)
	ad.Author = new(models.User)
	//ad.GeoPosition = new(models.GeoPosition)
	extraFieldTry := pgx.NullString{}
	/*lat := pgx.NullFloat64{}
	long := pgx.NullFloat64{}*/
	timeStamp := time.Time{}
	err := rows.Scan(&ad.AdId, &ad.Author.VkId, &ad.Author.Name, &ad.Author.Surname,
		&ad.Author.PhotoUrl, &ad.Header /*&ad.Text,*/, &ad.Region, &ad.District, &ad.AdType,
		&ad.LSEnabled, &ad.CommentsEnabled, &ad.ExtraEnabled,
		&extraFieldTry, &timeStamp /*&lat, &long,*/, &ad.Status, &ad.Category,
		&ad.CommentsCount, &ad.Hidden)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("UTC")
	timeStamp.In(loc)
	ad.CreationDate = timeStamp.Format("02 Jan 06 15:04 UTC")
	if extraFieldTry.Valid {
		ad.ExtraField = extraFieldTry.String
	}
	ad.PathesToPhoto, err = db.GetAdPhotos(ad.AdId)
	if err != nil {
		return nil, err
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
