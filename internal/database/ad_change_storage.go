package database

import (
	"fmt"
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
)

const (
	// create ad query
	CreateAd = "INSERT INTO ad (author_id, header, text, region, district, ad_type, ls_enabled, comments_enabled, extra_enabled, category%s)" +
		" VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10%s) RETURNING ad_id"
	ExtraField                = ", extra_field"
	GeoPosition               = ", geo_position, lat, long"
	Blank                     = ""
	NoExtraFieldNoGeoPosition = ""
	NoExtraFieldGeoPosition   = ", ST_SetSRID(ST_POINT($11, $12), 4326), $11, $12"
	ExtraFieldNoGeoPosition   = ", $11"
	ExtraFieldGeoPosition     = ", $11, ST_SetSRID(ST_POINT($12, $13), 4326), $12, $13"

	// edit ad query
	EditAd                        = "UPDATE ad SET header=$1, text=$2, region=$3, district=$4, ad_type=$5, ls_enabled=$6, comments_enabled=$7, " +
		"extra_enabled=$8, category=$9%s where ad_id=$%d"
	NoExtraFieldNoGeoPositionEdit = ", extra_field=NULL"
	NoExtraFieldGeoPositionEdit   = ", geo_position=ST_SetSRID(ST_POINT($10, $11), 4326), lat=$10, long=$11"
	ExtraFieldNoGeoPositionEdit   = ", extra_field=$10"
	ExtraFieldGeoPositionEdit     = ", extra_field=$10, geo_position=ST_SetSRID(ST_POINT($11, $12), 4326), lat=$11, long=$12"

	// add photo to ad query
	checkAdExist = "SELECT author_id FROM ad WHERE ad_id = $1"
	AddPhotoToAd = "INSERT INTO ad_photos (ad_id, photo_url) VALUES ($1, $2)"

	// deleteAd query
	deleteAd = "DELETE FROM ad WHERE ad_id = $1"

	// deletePhotos from ad query
	deleteAdPhotos = "DELETE FROM ad_photos WHERE ad_photos_id IN (%s) RETURNING photo_url"

	// check user exists
	checkUserExists = "SELECT EXISTS(SELECT 1 FROM users WHERE vk_id = $1)"

	// set ad hidden
	SetHidden = "UPDATE ad SET hidden = true WHERE ad_id = $1"

	// set ad visible
	SetVisible = "UPDATE ad SET hidden = false WHERE ad_id = $1"
)

func (db *DB) CreateAd(ad models.Ad) (int, models.AdCreationResult) {
	exists := false
	err := db.db.QueryRow(checkUserExists, ad.AuthorId).Scan(&exists)
	if err == pgx.ErrNoRows || !exists {
		return EMPTY_RESULT, models.AdCreationResult{}
	}
	if err != nil {
		return DB_ERROR, models.AdCreationResult{}
	}
	query := ""
	res := models.AdCreationResult{}
	sign := 0
	if ad.ExtraEnabled {
		sign = 10
	}
	if ad.GeoPosition != nil {
		sign += 1
	}

	switch sign {
	case 0:
		query = fmt.Sprintf(CreateAd, Blank, NoExtraFieldNoGeoPosition)
		err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled, ad.Category).Scan(&res.AdId)
	case 1:
		query = fmt.Sprintf(CreateAd, GeoPosition, NoExtraFieldGeoPosition)
		err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude).Scan(&res.AdId)
	case 10:
		query = fmt.Sprintf(CreateAd, ExtraField, ExtraFieldNoGeoPosition)
		err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, ad.ExtraField).Scan(&res.AdId)
	case 11:
		query = fmt.Sprintf(CreateAd, ExtraField+GeoPosition, ExtraFieldGeoPosition)
		err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, ad.ExtraField, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude).Scan(&res.AdId)
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
	defer func() {
		_ = tx.Rollback()
	}()
	authorId := 0
	err = tx.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT
	}
	if err != nil {
		return DB_ERROR
	}
	if authorId != userId {
		return FORBIDDEN
	}
	_, err = tx.Exec(AddPhotoToAd, adId, pathToPhoto)
	if err != nil {
		return DB_ERROR
	}
	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) DeleteAd(adId int, userId int) int {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR
	}
	defer func() {
		_ = tx.Rollback()
	}()
	authorId := 0
	err = tx.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT
	}
	if err != nil {
		return DB_ERROR
	}

	if authorId != userId {
		return FORBIDDEN
	}
	err = db.GiveCarmaBackDelete(tx, adId, userId)
	if err != nil {
		return DB_ERROR
	}
	_, err = tx.Exec(deleteAd, adId)
	if err != nil {
		return DB_ERROR
	}
	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) DeletePhotosFromAd(adId int, userId int, photoIds []string) (int, []string) {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR, nil
	}
	defer func() {
		_ = tx.Rollback()
	}()
	authorId := 0
	err = tx.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT, nil
	}
	if err != nil {
		return DB_ERROR, nil
	}

	if authorId != userId {
		return FORBIDDEN, nil
	}
	photoIdsInterface := make([]interface{}, len(photoIds))
	for i := range photoIds {
		photoIdsInterface[i] = photoIds[i]
	}
	placeHolders := "$1"
	for i := 2; i <= len(photoIds); i++ {
		placeHolders += fmt.Sprintf(", $%d", i)
	}
	query := fmt.Sprintf(deleteAdPhotos, placeHolders)
	rows, err := tx.Query(query, photoIdsInterface...)
	photoUrls := make([]string, 0)
	defer func() {
		rows.Close()
	}()
	for rows.Next() {
		path := ""
		err = rows.Scan(&path)
		if err != nil {
			return DB_ERROR, nil
		}
		photoUrls = append(photoUrls, path)
	}
	err = tx.Commit()
	if err != nil {
		return DB_ERROR, nil
	}
	return OK, photoUrls
}

func (db *DB) EditAd(adId int, userId int, ad models.Ad) int {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR
	}
	defer func() {
		_ = tx.Rollback()
	}()
	authorId := 0
	err = tx.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT
	}
	if err != nil {
		return DB_ERROR
	}

	if authorId != userId {
		return FORBIDDEN
	}
	sign := 0
	query := ""
	err = nil
	if ad.FeedbackType == Other {
		sign = 10
	}
	if ad.GeoPosition != nil {
		sign += 1
	}
	switch sign {
	case 0:
		query = fmt.Sprintf(EditAd, NoExtraFieldNoGeoPositionEdit, 8)
		_, err = tx.Exec(query, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled, ad.Category, adId)
	case 1:
		query = fmt.Sprintf(EditAd, NoExtraFieldGeoPositionEdit, 10)
		_, err = tx.Exec(query, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude, adId)
	case 10:
		query = fmt.Sprintf(EditAd, ExtraFieldNoGeoPositionEdit, 9)
		_, err = tx.Exec(query, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, ad.ExtraField, adId)
	case 11:
		query = fmt.Sprintf(EditAd, ExtraFieldGeoPositionEdit, 11)
		_, err = tx.Exec(query, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, ad.ExtraField, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude, adId)
	}
	if err != nil {
		return DB_ERROR
	}
	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) SetAdHidden(adId int, userId int) int {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR
	}
	defer func() {
		_ = tx.Rollback()
	}()
	authorId := 0
	err = tx.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT
	}
	if err != nil {
		return DB_ERROR
	}

	if authorId != userId {
		return FORBIDDEN
	}

	_, err = tx.Exec(SetHidden, adId)

	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) SetAdVisible(adId int, userId int) int {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR
	}
	defer func() {
		_ = tx.Rollback()
	}()
	authorId := 0
	err = tx.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT
	}
	if err != nil {
		return DB_ERROR
	}

	if authorId != userId {
		return FORBIDDEN
	}

	_, err = tx.Exec(SetVisible, adId)

	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}
