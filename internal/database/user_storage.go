package database

import (
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"log"
)

const (
	GetUserById = "SELECT * FROM users WHERE vk_id = $1"
	CreateUser  = "INSERT INTO users (vk_id, name, surname, photo_url) VALUES ($1, $2, $3, $4)"
	GetReceived = "SELECT a.ad_id, u.vk_id, u.carma, u.name, u.surname, u.photo_url, a.header, a.region," +
		" a.district, a.is_auction, a.feedback_type, a.extra_field, a.creation_datetime, a.status," +
		" a.category, a.comments_count, a.hidden FROM deal d JOIN (SELECT deal_id FROM deal WHERE subscriber_id = $1" +
		" AND status = 'success' ORDER BY deal_id LIMIT $2 OFFSET $3) l ON (l.deal_id = d.deal_id) JOIN ad a ON (d.ad_id = a.ad_id)" +
		" JOIN users u ON (a.author_id = u.vk_id) ORDER BY d.deal_id"
	GetGiven = "SELECT a.ad_id, u.vk_id, u.carma, u.name, u.surname, u.photo_url, a.header, a.region," +
		" a.district, a.is_auction, a.feedback_type, a.extra_field, a.creation_datetime, a.status," +
		" a.category, a.comments_count, a.hidden FROM ad a JOIN (SELECT ad_id FROM ad WHERE status = 'closed'" +
		" AND author_id = $1 ORDER BY ad_id LIMIT $2 OFFSET $3) l ON (l.ad_id = a.ad_id) JOIN users u ON (a.author_id = u.vk_id) ORDER BY ad_id"
)

func (db *DB) GetUser(userId int) (models.User, int) {
	row := db.db.QueryRow(GetUserById, userId)
	user := models.User{}
	err := row.Scan(&user.VkId, &user.Carma, &user.Name, &user.Surname, &user.PhotoUrl)
	if err == pgx.ErrNoRows {
		return user, EMPTY_RESULT
	}
	if err != nil {
		log.Println(err.Error())
		return user, DB_ERROR
	}
	return user, FOUND
}

func (db *DB) CreateUser(userId int, name string, surname string, photoURL string) int {
	_, err := db.db.Exec(CreateUser, userId, name, surname, photoURL)
	if err != nil {
		return DB_ERROR
	}
	return CREATED
}

func (db *DB) GetGiven(userId, page, rowsPerPage int) ([]models.AdForUsers, int) {
	offset := rowsPerPage * (page - 1)
	rows, err := db.db.Query(GetGiven, userId, rowsPerPage, offset)
	if err == pgx.ErrNoRows {
		return nil, EMPTY_RESULT
	}
	if err != nil {
		return nil, DB_ERROR
	}
	ads := make([]models.AdForUsers, 0)
	defer rows.Close()
	for rows.Next() {
		ads, err = db.WorkWithOneAd(rows, ads)
		if err != nil {
			return nil, DB_ERROR
		}
	}
	if len(ads) == 0 {
		return ads, EMPTY_RESULT
	}
	return ads, FOUND
}

func (db *DB) GetReceived(userId, page, rowsPerPage int) ([]models.AdForUsers, int) {
	offset := rowsPerPage * (page - 1)
	rows, err := db.db.Query(GetReceived, userId, rowsPerPage, offset)
	if err == pgx.ErrNoRows {
		return nil, EMPTY_RESULT
	}
	if err != nil {
		return nil, DB_ERROR
	}
	ads := make([]models.AdForUsers, 0)
	defer rows.Close()
	for rows.Next() {
		ads, err = db.WorkWithOneAd(rows, ads)
		if err != nil {
			return nil, DB_ERROR
		}
	}
	if len(ads) == 0 {
		return ads, EMPTY_RESULT
	}
	return ads, FOUND
}
