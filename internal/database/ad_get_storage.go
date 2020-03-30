package database

import (
	"fmt"
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"log"
	"time"
)

const (
	// constants
	LS       = "ls"
	Comments = "comments"
	Other    = "other"

	// get ad query
	GetAdById = "SELECT a.ad_id, u.vk_id, u.carma, u.name, u.surname, u.photo_url, a.header, a.text, a.region," +
		" a.district, a.is_auction, a.feedback_type, a.extra_field, a.creation_datetime, a.lat, a.long, a.status," +
		" a.category, a.comments_count FROM ad a JOIN users u ON (a.author_id = u.vk_id) WHERE a.ad_id = $1"

	// get ads query
	GetAds = "SELECT a.ad_id, u.vk_id, u.carma, u.name, u.surname, u.photo_url, a.header, a.text, a.region," +
		" a.district, a.is_auction, a.feedback_type, a.extra_field, a.creation_datetime, a.lat, a.long, a.status," +
		" a.category, a.comments_count FROM ad a JOIN users u ON (a.author_id = u.vk_id) " +
		"JOIN (SELECT ad_id FROM ad%s ORDER BY ad_id LIMIT $%d OFFSET $%d) l ON (l.ad_id = a.ad_id) ORDER BY ad_id"
	And            = "AND"
	Where          = " WHERE "
	CategoryClause = " category = $%d "
	AuthorClause   = " author_id = $%d "
	RegionClause   = " region = $%d "
	DistrictClause = " district = $%d "
	GetAdPhotos    = "SELECT ad_photos_id, photo_url FROM ad_photos WHERE ad_id = $1"
)

func (db *DB) GetAd(adId int) (models.AdForUsers, int) {
	row := db.db.QueryRow(GetAdById, adId)
	ad := models.AdForUsers{}
	ad.GeoPosition = new(models.GeoPosition)
	ad.Author = new(models.User)
	extraFieldTry := pgx.NullString{}
	lat := pgx.NullFloat64{}
	long := pgx.NullFloat64{}
	timeStamp := time.Time{}
	err := row.Scan(&ad.AdId, &ad.Author.VkId, &ad.Author.Carma, &ad.Author.Name, &ad.Author.Surname,
		&ad.Author.PhotoUrl, &ad.Header, &ad.Text, &ad.Region, &ad.District, &ad.IsAuction, &ad.FeedbackType,
		&extraFieldTry, &timeStamp, &lat, &long, &ad.Status, &ad.Category,
		&ad.CommentsCount)
	if err == pgx.ErrNoRows {
		return ad, EMPTY_RESULT
	}
	ad.GeoPosition.Available = true
	ad.CreationDate = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
	if extraFieldTry.Valid {
		ad.ExtraField = extraFieldTry.String
	}
	if lat.Valid && long.Valid {
		ad.GeoPosition.Latitude = lat.Float64
		ad.GeoPosition.Longitude = long.Float64
	} else {
		ad.GeoPosition = nil
	}
	if err != nil {
		log.Println(err.Error())
		return ad, DB_ERROR
	}
	photosRows, err := db.db.Query(GetAdPhotos, ad.AdId)
	if err != nil {
		return ad, DB_ERROR
	}
	defer photosRows.Close()
	for photosRows.Next() {
		adPhoto := models.AdPhoto{}
		err = photosRows.Scan(&adPhoto.AdPhotoId, &adPhoto.PhotoUrl)
		if err != nil {
			return ad, DB_ERROR
		}
		ad.PathesToPhoto = append(ad.PathesToPhoto, adPhoto)
	}
	return ad, FOUND
}

func (db *DB) FindAds(query string, page int, rowsPerPage int, params map[string][]string) ([]models.AdForUsers, int) {
	panic("not implemented")
}

func (db *DB) GetAds(page int, rowsPerPage int, params map[string][]string) ([]models.AdForUsers, int) {
	offset := rowsPerPage * (page - 1)
	query := GetAds
	whereClause := ""
	strArr := make([]interface{}, 0)
	categoryArr, ok := params["category"]
	if ok && len(categoryArr) == 1 {
		strArr = append(strArr, categoryArr[0])
		whereClause += Where + fmt.Sprintf(CategoryClause, 1)
	}

	authorArr, ok := params["author_id"]

	if ok && len(authorArr) == 1 {
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
	query = fmt.Sprintf(GetAds, whereClause, len(strArr)+1, len(strArr)+2)
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
	ad.GeoPosition = new(models.GeoPosition)
	extraFieldTry := pgx.NullString{}
	lat := pgx.NullFloat64{}
	long := pgx.NullFloat64{}
	timeStamp := time.Time{}
	err := rows.Scan(&ad.AdId, &ad.Author.VkId, &ad.Author.Carma, &ad.Author.Name, &ad.Author.Surname,
		&ad.Author.PhotoUrl, &ad.Header, &ad.Text, &ad.Region, &ad.District, &ad.IsAuction, &ad.FeedbackType,
		&extraFieldTry, &timeStamp, &lat, &long, &ad.Status, &ad.Category,
		&ad.CommentsCount)
	if err != nil {
		return nil, err
	}
	ad.GeoPosition.Available = true
	ad.CreationDate = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
	if extraFieldTry.Valid {
		ad.ExtraField = extraFieldTry.String
	}
	if lat.Valid && long.Valid {
		ad.GeoPosition.Latitude = lat.Float64
		ad.GeoPosition.Longitude = long.Float64
	} else {
		ad.GeoPosition = nil
	}
	photosRows, err := db.db.Query(GetAdPhotos, ad.AdId)
	if err != nil {
		return nil, err
	}
	defer photosRows.Close()
	for photosRows.Next() {
		adPhoto := models.AdPhoto{}
		err = photosRows.Scan(&adPhoto.AdPhotoId, &adPhoto.PhotoUrl)
		if err != nil {
			return nil, err
		}
		ad.PathesToPhoto = append(ad.PathesToPhoto, adPhoto)
	}
	ads = append(ads, *ad)
	return ads, nil
}
