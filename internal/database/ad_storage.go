package database

import (
	"fmt"
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"log"
)

const (
	// constants
	LS = "ls"
	Comments = "comments"
	Other = "other"

	// create ad query
	CreateAd = "INSERT INTO ad (author_id, header, text, region, district, is_auction, feedback_type, category%s)" +
		" VALUES($1, $2, $3, $4, $5, $6, $7, $8%s) RETURNING ad_id"
	ExtraField = ", extra_field"
	GeoPosition = ", geo_position"
	Blank = ""
	NoExtraFieldNoGeoPosition = ""
	NoExtraFieldGeoPosition = ", ST_POINT($9, $10)"
	ExtraFieldNoGeoPosition = ", $9"
	ExtraFieldGeoPosition = ", $9, ST_POINT($10, $11)"

	// add photo to ad query
	checkAdExist = "SELECT author_id FROM ad WHERE ad_id = $1"
	AddPhotoToAd = "INSERT INTO ad_photos (ad_id, photo_url) VALUES ($1, $2)"

	// get ad query
	GetAdById = "SELECT "
)

func (db *DB) CreateAd(ad models.Ad) (int, models.AdCreationResult) {
	query := ""
	var err error
	err = nil
	res := models.AdCreationResult{}
	sign := 0
	if ad.FeedbackType == Other {
		sign = 10
	}
	if ad.GeoPosition != nil {
		sign += 1
	}
	switch sign {
		case 0:
			query = fmt.Sprintf(CreateAd, Blank, NoExtraFieldNoGeoPosition)
			err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.IsAuction,
				ad.FeedbackType, ad.Category).Scan(&res.AdId)
		case 1:
			query = fmt.Sprintf(CreateAd, GeoPosition, NoExtraFieldGeoPosition)
			err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.IsAuction,
				ad.FeedbackType, ad.Category, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude).Scan(&res.AdId)
		case 10:
			query = fmt.Sprintf(CreateAd, ExtraField, ExtraFieldNoGeoPosition)
			err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.IsAuction,
				ad.FeedbackType, ad.Category, ad.ExtraField).Scan(&res.AdId)
		case 11:
			query = fmt.Sprintf(CreateAd, ExtraField + GeoPosition, ExtraFieldGeoPosition)
			err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.IsAuction,
				ad.FeedbackType, ad.Category, ad.ExtraField, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude).Scan(&res.AdId)
	}
	if err != nil {
		return DB_ERROR, res
	}
	return CREATED, res
}

func (db *DB) AddPhotoToAd(pathToPhoto string, adId int, userId int) int {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR
	}
	authorId :=0
	err = tx.QueryRow(checkAdExist).Scan(&authorId)
	if err != nil {
		return DB_ERROR
	}
	// TODO: uncomment when we can take userId from cookies
	/*if authorId != userId {
		return CONFLICT
	}*/
	_, err = tx.Exec(AddPhotoToAd, adId, pathToPhoto)
	if err != nil {
		return DB_ERROR
	}
	return FOUND
}

func (db *DB) GetAd(adId int) (models.Ad, int) {
	// TODO: implement
	row := db.db.QueryRow(GetAdById, adId)
	ad := models.Ad{}
	err := row.Scan()
	if err == pgx.ErrNoRows {
		return ad, EMPTY_RESULT
	}
	if err != nil {
		log.Println(err.Error())
		return ad, DB_ERROR
	}
	return ad, FOUND
}