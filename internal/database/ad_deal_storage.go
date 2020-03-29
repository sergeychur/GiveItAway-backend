package database

import (
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
)

const (
	// Subscribe to ad query
	SubscribeToAd = "INSERT INTO ad_subscribers (ad_id, subscriber_id) VALUES ($1, $2)"

	// Get ad subscribers query
	GetAdSubscribers = "SELECT u.vk_id, u.name, u.surname, u.carma, u.photo_url FROM ad_subscribers a_s JOIN" +
		" (SELECT ad_subscribers_id FROM ad_subscribers WHERE ad_id = $1 ORDER BY ad_subscribers_id LIMIT $2 OFFSET $3)" +
		" l ON (l.ad_subscribers_id = a_s.ad_subscribers_id) JOIN users u ON (u.vk_id = a_s.subscriber_id) ORDER BY a_s.ad_subscribers_id"
)

func (db *DB) SubscribeToAd(adId int, userId int) int {
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
	// TODO: uncomment when we can take userId from cookies

	/*if authorId == userId {
		 return CONFLICT
	}*/
	_, err = tx.Exec(SubscribeToAd, adId, userId)
	if err != nil {
		return DB_ERROR
	}
	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) GetAdSubscribers(adId int, page int, rowsPerPage int) ([]models.User, int){
	offset := rowsPerPage * ( page - 1)
	rows, err := db.db.Query(GetAdSubscribers, adId, rowsPerPage, offset)
	if err == pgx.ErrNoRows {
		return nil, EMPTY_RESULT
	}
	if err != nil {
		return nil, DB_ERROR
	}
	users := make([]models.User, 0)
	defer rows.Close()
	for rows.Next() {
		user := models.User{}
		err = rows.Scan(&user.VkId, &user.Name, &user.Surname, &user.Carma, &user.PhotoUrl)
		if err != nil {
			return nil, DB_ERROR
		}
		users = append(users, user)
	}
	if len(users) == 0 {
		return users, EMPTY_RESULT
	}
	return users, FOUND
}