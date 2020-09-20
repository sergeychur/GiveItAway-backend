package database

import (
	"fmt"
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"log"
)

const (
	// create ad query
	CreateAd = "INSERT INTO ad (author_id, hidden, header, text, region, district, ad_type, ls_enabled, comments_enabled, extra_enabled, category, subcat_list, subcat, metro, full_adress%s)" +
		" VALUES($1, true,  $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14%s) RETURNING ad_id"
	ExtraField                = ", extra_field"
	GeoPosition               = ", geo_position, lat, long"
	Blank                     = ""
	NoExtraFieldNoGeoPosition = ""
	NoExtraFieldGeoPosition   = ", ST_SetSRID(ST_POINT($15, $16), 4326), $15, $16"
	ExtraFieldNoGeoPosition   = ", $15"
	ExtraFieldGeoPosition     = ", $15, ST_SetSRID(ST_POINT($16, $17), 4326), $16, $17"

	// edit ad query
	EditAd = "UPDATE ad SET header=$1, text=$2, region=$3, district=$4, ls_enabled=$5, comments_enabled=$6, " +
		"extra_enabled=$7, category=$8, subcat_list=$9, subcat=$10, metro=$11, full_adress=$12%s where ad_id=$%d"
	NoExtraFieldNoGeoPositionEdit = ", extra_field=NULL"
	NoExtraFieldGeoPositionEdit   = ", geo_position=ST_SetSRID(ST_POINT($13, $14), 4326), lat=$13, long=$14"
	ExtraFieldNoGeoPositionEdit   = ", extra_field=$13"
	ExtraFieldGeoPositionEdit     = ", extra_field=$13, geo_position=ST_SetSRID(ST_POINT($14, $15), 4326), lat=$14, long=$15"

	// add photo to ad query
	CountAdPhotos = "SELECT COUNT(*) from ad_photos WHERE ad_id = $1"
	checkAdExist = "SELECT author_id FROM ad WHERE ad_id = $1"
	AddPhotoToAd = "INSERT INTO ad_photos (ad_id, photo_url) VALUES ($1, $2)"

	// deleteAd query
	deleteAd   = "DELETE FROM ad WHERE ad_id = $1"
	clearNotes = "DELETE FROM notifications WHERE ad_id = $1"

	// deletePhotos from ad query
	deleteAdPhotos = "DELETE FROM ad_photos WHERE ad_photos_id IN (%s) RETURNING photo_url"

	// check user exists
	checkUserExists = "SELECT EXISTS(SELECT 1 FROM users WHERE vk_id = $1)"

	// set ad hidden
	SetHidden = "UPDATE ad SET hidden = true WHERE ad_id = $1"

	// set ad visible
	SetVisible = "UPDATE ad SET hidden = false WHERE ad_id = $1"

	CheckIfCanDeleteOrUpdate = "SELECT status != 'closed' AND status != 'aborted' FROM ad WHERE ad_id = $1"
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
		if ad.GeoPosition.Latitude < -90 || ad.GeoPosition.Latitude > 90 ||
			ad.GeoPosition.Longitude < -180 || ad.GeoPosition.Longitude > 180 {
			return WRONG_INPUT, models.AdCreationResult{}
		}
		sign += 1
	}
	metro := pgx.NullString{String: ad.Metro}
	if ad.Metro != "" {
		metro.Valid = true
	}
	fullAdress := pgx.NullString{String: ad.FullAdress}
	if ad.FullAdress != "" {
		fullAdress.Valid = true
	}
	subcatList := pgx.NullString{String: ad.SubCatList}
	if ad.SubCatList != "" {
		subcatList.Valid = true
	}

	subcat := pgx.NullString{String: ad.SubCat}
	if ad.SubCat != "" {
		subcat.Valid = true
	}
	switch sign {
	case 0:
		query = fmt.Sprintf(CreateAd, Blank, NoExtraFieldNoGeoPosition)
		err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled, ad.Category, subcatList, subcat,
			metro, fullAdress).Scan(&res.AdId)
	case 1:
		query = fmt.Sprintf(CreateAd, GeoPosition, NoExtraFieldGeoPosition)
		err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, subcatList, subcat,
			metro, fullAdress, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude).Scan(&res.AdId)
	case 10:
		query = fmt.Sprintf(CreateAd, ExtraField, ExtraFieldNoGeoPosition)
		err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, subcatList, subcat, metro, fullAdress, ad.ExtraField).Scan(&res.AdId)
	case 11:
		query = fmt.Sprintf(CreateAd, ExtraField+GeoPosition, ExtraFieldGeoPosition)
		err = db.db.QueryRow(query, ad.AuthorId, ad.Header, ad.Text, ad.Region, ad.District, ad.AdType,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, subcatList, subcat,
			metro, fullAdress, ad.ExtraField, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude).Scan(&res.AdId)
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

	var allow = false
	for _, id := range WHITE_LIST {
		if userId == id {
			allow = true
		}
	}
	if !allow && authorId != userId {
		return FORBIDDEN
	}
	canDelete := false
	err = db.db.QueryRow(CheckIfCanDeleteOrUpdate, adId).Scan(&canDelete)
	if err != nil {
		return DB_ERROR
	}
	if !canDelete {
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
	_, err = tx.Exec(clearNotes, adId)
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
	if ad.ExtraEnabled {
		sign = 10
	}
	if ad.GeoPosition != nil {
		if ad.GeoPosition.Latitude < -90 || ad.GeoPosition.Latitude > 90 ||
			ad.GeoPosition.Longitude < -180 || ad.GeoPosition.Longitude > 180 {
			return WRONG_INPUT
		}
		sign += 1
	}
	metro := pgx.NullString{String: ad.Metro}
	if ad.Metro != "" {
		metro.Valid = true
	}
	fullAdress := pgx.NullString{String: ad.FullAdress}
	if ad.FullAdress != "" {
		fullAdress.Valid = true
	}

	subcatList := pgx.NullString{String: ad.SubCatList}
	if ad.SubCatList != "" {
		subcatList.Valid = true
	}

	subcat := pgx.NullString{String: ad.SubCat}
	if ad.SubCat != "" {
		subcat.Valid = true
	}

	canUpdate := false
	err = db.db.QueryRow(CheckIfCanDeleteOrUpdate, adId).Scan(&canUpdate)
	if err != nil {
		return DB_ERROR
	}
	if !canUpdate {
		return FORBIDDEN
	}

	switch sign {
	case 0:
		query = fmt.Sprintf(EditAd, NoExtraFieldNoGeoPositionEdit, 13)
		_, err = tx.Exec(query, ad.Header, ad.Text, ad.Region, ad.District,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled, ad.Category, subcatList, subcat, adId, metro, fullAdress)
	case 1:
		query = fmt.Sprintf(EditAd, NoExtraFieldGeoPositionEdit, 15)
		_, err = tx.Exec(query, ad.Header, ad.Text, ad.Region, ad.District,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, subcatList, subcat, metro, fullAdress, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude, adId)
	case 10:
		query = fmt.Sprintf(EditAd, ExtraFieldNoGeoPositionEdit, 14)
		_, err = tx.Exec(query, ad.Header, ad.Text, ad.Region, ad.District,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, subcatList, subcat, metro, fullAdress, ad.ExtraField, adId)
	case 11:
		query = fmt.Sprintf(EditAd, ExtraFieldGeoPositionEdit, 16)
		_, err = tx.Exec(query, ad.Header, ad.Text, ad.Region, ad.District,
			ad.LSEnabled, ad.CommentsEnabled, ad.ExtraEnabled,
			ad.Category, subcatList, subcat, metro, fullAdress,
			ad.ExtraField, ad.GeoPosition.Latitude, ad.GeoPosition.Longitude, adId)
	}
	if err != nil {
		return DB_ERROR
	}

	_, err = tx.Exec(SetHidden, adId)
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

	var allow = false
	for _, id := range WHITE_LIST {
		if userId == id {
			allow = true
		}
	}
	log.Println("heello", userId, authorId, WHITE_LIST, allow)
	if !allow {
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

	var allow = false
	for _, id := range WHITE_LIST {
		if userId == id {
			allow = true
		}
	}
	if !allow {
		return FORBIDDEN
	}

	_, err = tx.Exec(SetVisible, adId)

	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) CanUploadPhoto(adId int, maxPhotoNum int) (bool, int) {
	photoNum := 0
	err := db.db.QueryRow(CountAdPhotos, adId).Scan(&photoNum)

	if err == pgx.ErrNoRows {
		log.Println(err)
		return true, OK
	}
	log.Println("has ", photoNum, " photos")
	if err != nil {
		return false, DB_ERROR
	}
	if photoNum < maxPhotoNum {
		return true, OK
	}
	return false, OK
}