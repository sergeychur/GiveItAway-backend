package database

import (
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
)

const (
	// Subscribe to ad query
	SubscribeToAd = "INSERT INTO ad_subscribers (ad_id, subscriber_id) VALUES ($1, $2)"

	// Unsubscribe from ad query
	UnsubscribeFromAd = "DELETE FROM ad_subscribers WHERE ad_id = $1 AND subscriber_id = $2"

	// Get ad subscribers query
	GetAdSubscribers = "SELECT u.vk_id, u.name, u.surname, u.carma, u.photo_url FROM ad_subscribers a_s JOIN" +
		" (SELECT ad_subscribers_id FROM ad_subscribers WHERE ad_id = $1 ORDER BY ad_subscribers_id LIMIT $2 OFFSET $3) " +
		" l ON (l.ad_subscribers_id = a_s.ad_subscribers_id) JOIN users u ON (u.vk_id = a_s.subscriber_id) " +
		"ORDER BY a_s.ad_subscribers_id"

	// Make deal query
	CheckIfSubscriber = "SELECT EXISTS(SELECT 1 FROM ad_subscribers WHERE ad_id = $1 AND subscriber_id = $2)"
	CheckIfDealExists = "SELECT EXISTS(SELECT 1 FROM deal WHERE ad_id = $1)"
	CreateDeal = "SELECT make_deal($1, $2)"
	GetDeal = "SELECT * FROM deal WHERE deal_id = $1"

	// Fulfill deal
	GetDealWithAuthor = "SELECT d.*, a.author_id FROM deal d JOIN ad a ON (a.ad_id = d.Ad_id) WHERE d.deal_id = $1"
	FulfillDeal = "SELECT close_deal_success($1)"

	// CancelDeal
	CancelDealAuthor = "SELECT close_deal_fail_by_author($1)"
	CancelDealSubscriber = "SELECT close_deal_fail_by_subscriber($1)"

	// Get Deal
	GetDealForAd = "SELECT * FROM deal WHERE ad_id = $1"
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

	if authorId == userId {
		 return FORBIDDEN
	}
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

func (db *DB) GetAdSubscribers(adId int, page int, rowsPerPage int) ([]models.User, int) {
	offset := rowsPerPage * (page - 1)
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

func (db *DB) UnsubscribeFromAd(adId int, userId int) int {
	_, err := db.db.Exec(UnsubscribeFromAd, adId, userId)
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) MakeDeal(adId int, subscriberId int, initiatorId int) (int, int) {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR, 0
	}
	defer func() {
		_ = tx.Rollback()
	}()

	authorId := 0
	// TODO: maybe add check if the ad is open(you cannot deal with closed ad)
	err = tx.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT, 0
	}
	if err != nil {
		return DB_ERROR, 0
	}

	if authorId != initiatorId {
		return FORBIDDEN, 0
	}
	isSubscriber := false
	err = tx.QueryRow(CheckIfSubscriber, adId, subscriberId).Scan(&isSubscriber)
	if err != nil {
		return DB_ERROR, 0
	}
	dealExists := true
	err = tx.QueryRow(CheckIfDealExists, adId).Scan(&dealExists)
	if err != nil {
		return DB_ERROR, 0
	}
	if dealExists || !isSubscriber {
		return CONFLICT, 0
	}
	dealId := 0
	err = tx.QueryRow(CreateDeal, adId, subscriberId).Scan(&dealId)
	if err != nil {
		return DB_ERROR, 0
	}
	_ = tx.Commit()
	return CREATED, dealId
}

func (db *DB) FulfillDeal(dealId int, userId int) int {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR
	}
	defer func() {
		_ = tx.Rollback()
	}()
	dealIdGot := 0
	subscriberId := 0
	adId := 0
	status := ""
	err = tx.QueryRow(GetDeal, dealId).Scan(&dealIdGot, &adId, &subscriberId, &status)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT
	}
	if err != nil {
		return DB_ERROR
	}
	if userId != subscriberId || status != "open" {
		return FORBIDDEN
	}
	_, err = tx.Exec(FulfillDeal, dealId)
	if err != nil {
		return DB_ERROR
	}
	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) CancelDeal(dealId int, userId int) int {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR
	}
	defer func() {
		_ = tx.Rollback()
	}()
	dealIdGot := 0
	subscriberId := 0
	authorId := 0
	adId := 0
	status := ""
	err = tx.QueryRow(GetDealWithAuthor, dealId).Scan(&dealIdGot, &adId, &subscriberId, &status, &authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT
	}
	if err != nil {
		return DB_ERROR
	}
	if status != "open" {
		return FORBIDDEN
	}
	if subscriberId == userId {
		_, err = tx.Exec(CancelDealSubscriber, dealId)
		if err != nil {
			return DB_ERROR
		}
		err = tx.Commit()
		if err != nil {
			return DB_ERROR
		}
		return OK
	}

	if authorId == userId {
		_, err = tx.Exec(CancelDealAuthor, dealId)
		if err != nil {
			return DB_ERROR
		}
		err = tx.Commit()
		if err != nil {
			return DB_ERROR
		}
		return OK
	}
	return FORBIDDEN
}

func (db *DB) GetDealForAd(adId int) (models.DealDetails, int) {
	deal := models.DealDetails{}
	err := db.db.QueryRow(GetDealForAd, adId).Scan(&deal.DealId, &deal.AdId, &deal.SubscriberId, &deal.Status)
	if err == pgx.ErrNoRows {
		return deal, EMPTY_RESULT
	}
	if err != nil {
		return deal, DB_ERROR
	}
	return deal, FOUND
}
