package database

import (
	"fmt"
	"github.com/go-vk-api/vk"
	"log"
	"time"

	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
)

const (
	GetUserProfileById = "SELECT u.vk_id, u_c.current_carma, u.name, u.surname, u.photo_url, u.registration_date_time, u_c.frozen_carma, " +
		"u_s.total_earned_carma, u_s.total_spent_carma, u_s.total_given_ads, u_s.total_received_ads, " +
		"u_s.total_aborted_ads, u.last_change_time FROM users u JOIN users_stats u_s ON u_s.user_id = u.vk_id " +
		"JOIN users_carma u_c ON u_c.user_id = u.vk_id WHERE u.vk_id = $1"

	GetUserById = "SELECT u.vk_id, u.name, u.surname, u.photo_url, u.last_change_time" +
		" FROM users u WHERE vk_id = $1"

	CreateUser = "INSERT INTO users (vk_id, name, surname, photo_url) VALUES ($1, $2, $3, $4)"

	GrantInitialCarma = "UPDATE users_carma SET current_carma = $1"

	GetReceived = "SELECT a.ad_id, u.vk_id, u.name, u.surname, u.photo_url, a.header, a.region," +
		" a.district, a.ad_type, a.ls_enabled, a.comments_enabled, a.extra_enabled, a.extra_field, a.creation_datetime, a.status," +
		" a.category, a.subcat_list, a.subcat, a.comments_count, a.hidden, a.metro, u.last_change_time" +
		" FROM deal d JOIN (SELECT deal_id FROM deal WHERE subscriber_id = $1" +
		" AND status = 'success' ORDER BY deal_id LIMIT $2 OFFSET $3) l ON (l.deal_id = d.deal_id) JOIN ad a ON (d.ad_id = a.ad_id)" +
		" JOIN users u ON (a.author_id = u.vk_id) ORDER BY d.deal_id"

	GetGiven = "SELECT a.ad_id, u.vk_id, u.name, u.surname, u.photo_url, a.header, a.region," +
		" a.district, a.ad_type, a.ls_enabled, a.comments_enabled, a.extra_enabled, a.extra_field, a.creation_datetime, a.status," +
		" a.category, a.subcat_list, a.subcat, a.comments_count, a.hidden, a.metro, u.last_change_time" +
		" FROM ad a JOIN (SELECT ad_id FROM ad WHERE status = 'closed'" +
		" AND author_id = $1 ORDER BY ad_id LIMIT $2 OFFSET $3) l ON (l.ad_id = a.ad_id) JOIN users u ON (a.author_id = u.vk_id) ORDER BY ad_id"

	GetWanted = "SELECT a.ad_id, u.vk_id, u.name, u.surname, u.photo_url, a.header, a.region," +
		" a.district, a.ad_type, a.ls_enabled, a.comments_enabled, a.extra_enabled, a.extra_field, a.creation_datetime, a.status," +
		" a.category, a.subcat_list, a.subcat, a.comments_count, a.hidden, a.metro, u.last_change_time" +
		" FROM ad_subscribers a_s JOIN (SELECT ad_subscribers_id FROM ad_subscribers WHERE subscriber_id = $1" +
		" ORDER BY ad_subscribers_id LIMIT $2 OFFSET $3) l ON (l.ad_subscribers_id = a_s.ad_subscribers_id) JOIN ad a ON (a_s.ad_id = a.ad_id)" +
		" JOIN users u ON (a.author_id = u.vk_id) ORDER BY a_s.ad_subscribers_id"

	GetSendNotificationsToPM = "SELECT send_notifications_to_pm From users where vk_id=$1"
	SetSendNotificationsToPM = "UPDATE users SET send_notifications_to_pm = $1 where vk_id=$2"

	UpdateUserInfo = "UPDATE users SET name = $2, surname = $3, photo_url = $4, last_change_time = $5 WHERE vk_id = $1"
	// updatedUser.VkId, updatedUser.Name, updatedUser.Surname, time.Now()
)

func (db *DB) GetUserProfile(userId int, client *vk.Client, allowedDuration time.Duration) (models.UserForProfile, int) {
	row := db.db.QueryRow(GetUserProfileById, userId)
	user := models.UserForProfile{}
	timeStamp := time.Time{}
	updateTime := time.Time{}
	err := row.Scan(&user.VkId, &user.Carma, &user.Name, &user.Surname, &user.PhotoUrl, &timeStamp, &user.FrozenCarma,
		&user.TotalEarnedCarma, &user.TotalSpentCarma, &user.TotalGivenAds, &user.TotalReceivedAds, &user.TotalAbortedAds, &updateTime)
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
	result, err := db.CheckVKUserCacheConsist(client, user.VkId, updateTime, allowedDuration)
	if err != nil {
		// some needed actions
		log.Println(err)
	} else {
		user.PhotoUrl = result.PhotoURL
		user.Name = result.Name
		user.Surname = result.Surname
	}
	return user, FOUND
}

func (db *DB) GetUser(userId int, client *vk.Client, allowedDuration time.Duration) (models.User, int) {
	row := db.db.QueryRow(GetUserById, userId)
	user := models.User{}
	updateTime := time.Time{}
	err := row.Scan(&user.VkId, &user.Name, &user.Surname, &user.PhotoUrl, &updateTime)
	if err == pgx.ErrNoRows {
		return user, EMPTY_RESULT
	}
	if err != nil {
		log.Println(err.Error())
		return user, DB_ERROR
	}
	result, err := db.CheckVKUserCacheConsist(client, user.VkId, updateTime, allowedDuration)
	if err != nil {
		// some needed actions
		log.Println(err)
	} else {
		user.PhotoUrl = result.PhotoURL
		user.Name = result.Name
		user.Surname = result.Surname
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
		log.Println("Error in GIVEN")
		log.Println(err)
		return nil, DB_ERROR
	}
	ads := make([]models.AdForUsers, 0)
	defer rows.Close()
	for rows.Next() {
		ads, err = db.WorkWithOneAd(rows, ads, nil, 1)
		if err != nil {
			log.Println("Error in GIVEN")
			log.Println(err)
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
		log.Println("Error in received")
		log.Println(err)
		return nil, DB_ERROR
	}
	ads := make([]models.AdForUsers, 0)
	defer rows.Close()
	for rows.Next() {
		ads, err = db.WorkWithOneAd(rows, ads, nil, 1)
		if err != nil {
			log.Println("Error in received")
			log.Println(err)
			return nil, DB_ERROR
		}
	}
	if len(ads) == 0 {
		return ads, EMPTY_RESULT
	}
	return ads, FOUND
}

func (db *DB) GetWanted(userId, page, rowsPerPage int, client *vk.Client, allowedDuration time.Duration) ([]models.AdForUsers, int) {
	offset := rowsPerPage * (page - 1)
	rows, err := db.db.Query(GetWanted, userId, rowsPerPage, offset)
	if err == pgx.ErrNoRows {
		return nil, EMPTY_RESULT
	}
	if err != nil {
		log.Println("Error in wanted")
		log.Println(err)
		return nil, DB_ERROR
	}
	ads := make([]models.AdForUsers, 0)
	defer rows.Close()
	for rows.Next() {
		ads, err = db.WorkWithOneAd(rows, ads, client, allowedDuration)
		if err != nil {
			log.Println("Error in wanted")
			log.Println(err)
			return nil, DB_ERROR
		}
	}
	if len(ads) == 0 {
		return ads, EMPTY_RESULT
	}
	return ads, FOUND
}

func (db *DB) GetPermissoinToPM(userId int) (bool, int) {
	row := db.db.QueryRow(GetSendNotificationsToPM, userId)
	var canSend bool
	err := row.Scan(&canSend)
	if err == pgx.ErrNoRows {
		return canSend, EMPTY_RESULT
	}
	if err != nil {
		log.Println(err.Error())
		return canSend, DB_ERROR
	}
	return canSend, FOUND
}

func (db *DB) ChangePermissoinToPM(userId int, canSend bool) int {
	_, err := db.db.Exec(SetSendNotificationsToPM, canSend, userId)
	if err != nil {
		return DB_ERROR
	}
	return CREATED

}

func (db *DB) CheckVKUserCacheConsist(client *vk.Client, vkID int64, lastUpdated time.Time,
	allowedDuration time.Duration) (models.UserCacheCheck, error) {
	if client == nil {
		return models.UserCacheCheck{}, fmt.Errorf("no client provided")
	}
	duration := time.Now().Sub(lastUpdated)
	if duration > allowedDuration {
		var users []struct{
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Photo100 string `json:"photo_100"`}
		err := client.CallMethod("users.get", vk.RequestParams{
			"user_ids":   vkID,
			"name_case": "nom",
			"fields": "photo_100",
			"lang": "ru",	// TODO: talk about that
		}, &users)
		if err != nil {
			log.Print("vkClient error is", err)
			return models.UserCacheCheck{}, err
		}
		if len(users) == 0 {
			log.Println("got empty result")
			return models.UserCacheCheck{}, fmt.Errorf("got empty result")
		}
		updatedUser := models.UserCacheCheck{
			VkId: users[0].ID,
			Name: users[0].FirstName,
			Surname: users[0].LastName,
			PhotoURL: users[0].Photo100,
		}
		_, err = db.db.Exec(UpdateUserInfo, updatedUser.VkId, updatedUser.Name, updatedUser.Surname,
			updatedUser.PhotoURL, time.Now())
		if err != nil {
			log.Println("Can not update record in database: ", err)
			return models.UserCacheCheck{}, err
		}
		return updatedUser, nil
	}

	return models.UserCacheCheck{}, fmt.Errorf("no need in updating, cache is valid")
}