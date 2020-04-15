package database

import (
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"log"
	"time"
)

const (
	GetUserProfileById = "SELECT u.vk_id, u_c.current_carma, u.name, u.surname, u.photo_url, u.registration_date_time, u_c.frozen_carma, " +
		"u_s.total_earned_carma, u_s.total_spent_carma, u_s.total_given_ads, u_s.total_received_ads, " +
		"u_s.total_aborted_ads FROM users u JOIN users_stats u_s ON u_s.user_id = u.vk_id " +
		"JOIN users_carma u_c ON u_c.user_id = u.vk_id WHERE u.vk_id = $1"

	GetUserById = "SELECT u.vk_id, u.name, u.surname, u.photo_url" +
		" FROM users u WHERE vk_id = $1"

	CreateUser  = "INSERT INTO users (vk_id, name, surname, photo_url) VALUES ($1, $2, $3, $4)"
	GrantInitialCarma = "UPDATE users_carma SET current_carma = $1"
	GetReceived = "SELECT a.ad_id, u.vk_id, u.name, u.surname, u.photo_url, a.header, a.region," +
		" a.district, a.is_auction, a.feedback_type, a.extra_field, a.creation_datetime, a.status," +
		" a.category, a.comments_count, a.hidden FROM deal d JOIN (SELECT deal_id FROM deal WHERE subscriber_id = $1" +
		" AND status = 'success' ORDER BY deal_id LIMIT $2 OFFSET $3) l ON (l.deal_id = d.deal_id) JOIN ad a ON (d.ad_id = a.ad_id)" +
		" JOIN users u ON (a.author_id = u.vk_id) ORDER BY d.deal_id"
	GetGiven = "SELECT a.ad_id, u.vk_id, u.name, u.surname, u.photo_url, a.header, a.region," +
		" a.district, a.is_auction, a.feedback_type, a.extra_field, a.creation_datetime, a.status," +
		" a.category, a.comments_count, a.hidden FROM ad a JOIN (SELECT ad_id FROM ad WHERE status = 'closed'" +
		" AND author_id = $1 ORDER BY ad_id LIMIT $2 OFFSET $3) l ON (l.ad_id = a.ad_id) JOIN users u ON (a.author_id = u.vk_id) ORDER BY ad_id"
)

func (db *DB) GetUserProfile(userId int) (models.UserForProfile, int) {
	row := db.db.QueryRow(GetUserProfileById, userId)
	user := models.UserForProfile{}
	timeStamp := time.Time{}
	err := row.Scan(&user.VkId, &user.Carma, &user.Name, &user.Surname, &user.PhotoUrl, &timeStamp, &user.FrozenCarma,
		&user.TotalEarnedCarma, &user.TotalSpentCarma, &user.TotalGivenAds, &user.TotalReceivedAds, &user.TotalAbortedAds)
	if err == pgx.ErrNoRows {
		return user, EMPTY_RESULT
	}
	if err != nil {
		log.Println(err.Error())
		return user, DB_ERROR
	}
	loc, _ := time.LoadLocation("UTC")
	timeStamp.In(loc)
	user.RegistrationDate = timeStamp.Format("02 Jan 06 15:04 UTC")
	return user, FOUND
}

func (db *DB) GetUser(userId int) (models.User, int) {
	row := db.db.QueryRow(GetUserById, userId)
	user := models.User{}
	err := row.Scan(&user.VkId, &user.Name, &user.Surname, &user.PhotoUrl)
	if err == pgx.ErrNoRows {
		return user, EMPTY_RESULT
	}
	if err != nil {
		log.Println(err.Error())
		return user, DB_ERROR
	}
	return user, FOUND
}

func (db *DB) CreateUser(userId int, name string, surname string, photoURL string, initialCarma int) int {
	_, err := db.db.Exec(CreateUser, userId, name, surname, photoURL)
	if err != nil {
		return DB_ERROR
	}
	_, err = db.db.Exec(GrantInitialCarma, initialCarma)
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
