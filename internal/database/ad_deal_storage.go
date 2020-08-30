package database

import (
	"github.com/sergeychur/give_it_away/internal/global_constants"
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"math/rand"
	"net/url"
	"strconv"
)

const (
	// Subscribe to ad query
	SubscribeToAd = "INSERT INTO ad_subscribers (ad_id, subscriber_id, bid) VALUES ($1, $2, $3)"

	// Unsubscribe from ad query
	UnsubscribeFromAd = "DELETE FROM ad_subscribers WHERE ad_id = $1 AND subscriber_id = $2"

	// Make deal query
	CheckIfSubscriber = "SELECT EXISTS(SELECT 1 FROM ad_subscribers WHERE ad_id = $1 AND subscriber_id = $2)"
	CheckIfDealExists = "SELECT EXISTS(SELECT 1 FROM deal WHERE ad_id = $1)"
	CreateDeal        = "SELECT make_deal($1, $2)"
	GetDeal           = "SELECT * FROM deal WHERE deal_id = $1"
	GetRichest        = "select subscriber_id from ad_subscribers where ad_id = $1 order by bid desc limit 1"

	// Fulfill deal
	GetDealWithAuthor = "SELECT d.*, a.author_id FROM deal d JOIN ad a ON (a.ad_id = d.Ad_id) WHERE d.deal_id = $1"
	FulfillDeal       = "SELECT close_deal_success($1, $2)"

	// CancelDeal
	CancelDealAuthor     = "SELECT close_deal_fail_by_author($1)"
	CancelDealSubscriber = "SELECT close_deal_fail_by_subscriber($1, $2)"

	// Get Deal
	GetDealForAd = "SELECT * FROM deal WHERE ad_id = $1"

	// Get ad subscribers query
	GetAdSubscribers = "SELECT u.vk_id, u.name, u.surname, u.photo_url FROM ad_subscribers a_s JOIN" +
		" (SELECT ad_subscribers_id FROM ad_subscribers WHERE ad_id = $1 ORDER BY ad_subscribers_id LIMIT $2 OFFSET $3) " +
		" l ON (l.ad_subscribers_id = a_s.ad_subscribers_id) JOIN users u ON (u.vk_id = a_s.subscriber_id) " +
		"ORDER BY a_s.ad_subscribers_id"

	GetAdSubscribersIds = "SELECT a_s.subscriber_id FROM ad_subscribers a_s WHERE a_s.ad_id = $1"

	CheckAdHidden = "SELECT hidden FROM ad WHERE ad_id = $1"
	CheckAdOffer = "SELECT status = 'offer' FROM ad WHERE ad_id = $1"
)

func (db *DB) SubscribeToAd(adId int, userId int, priceCoeff int) (int, *models.Notification) {
	// todo check if user can subscribe; two different functions for auction and usual ad
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

	hidden := false
	err = tx.QueryRow(CheckAdHidden, adId).Scan(&hidden)
	if hidden {
		return FORBIDDEN, nil
	}

	isOffer := false
	err = tx.QueryRow(CheckAdOffer, adId).Scan(&isOffer)
	if !isOffer {
		return FORBIDDEN, nil
	}

	if authorId == userId {
		return FORBIDDEN, nil
	}
	isSubscribed := false
	err = tx.QueryRow(CheckIfSubscriber, adId, userId).Scan(&isSubscribed)
	if isSubscribed {
		return FORBIDDEN, nil
	}
	canSubscribe, frozencarma, err, note := db.DealWithCarmaSubscribe(tx, adId, userId)
	if err != nil {
		return DB_ERROR, nil
	}

	if !canSubscribe {
		return CONFLICT, nil
	}
	_, err = tx.Exec(SubscribeToAd, adId, userId, frozencarma)
	if err != nil {
		return DB_ERROR, nil
	}
	err = tx.Commit()
	if err != nil {
		return DB_ERROR, nil
	}
	return OK, note
}

func (db *DB) UnsubscribeFromAd(adId int, userId int) int {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR
	}
	defer func() {
		_ = tx.Rollback()
	}()

	isSubscriber := false
	err = db.db.QueryRow(CheckIfSubscriber, adId, userId).Scan(&isSubscriber)
	if !isSubscriber {
		return FORBIDDEN
	}
	err = db.DealWithCarmaUnsubscribe(tx, adId, userId)
	if err != nil {
		return DB_ERROR
	}
	_, err = tx.Exec(UnsubscribeFromAd, adId, userId)
	if err != nil {
		return DB_ERROR
	}
	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) MakeDeal(adId int, initiatorId int, dealType string, params url.Values) (int, int, int) {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR, 0, 0
	}
	defer func() {
		_ = tx.Rollback()
	}()
	authorId := 0
	// TODO: maybe add check if the ad is open(you cannot deal with closed ad)
	err = tx.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT, 0, 0
	}
	if err != nil {
		return DB_ERROR, 0, 0
	}

	if authorId != initiatorId {
		return FORBIDDEN, 0, 0
	}
	dealExists := true
	err = tx.QueryRow(CheckIfDealExists, adId).Scan(&dealExists)
	if err != nil {
		return DB_ERROR, 0, 0
	}
	subscriberId, status := db.GetSubscriberIdForDeal(adId, dealType, params)
	if status != OK {
		return status, 0, 0
	}
	isSubscriber := false
	err = tx.QueryRow(CheckIfSubscriber, adId, subscriberId).Scan(&isSubscriber)
	if err != nil {
		return DB_ERROR, 0, 0
	}
	if dealExists || !isSubscriber {
		return CONFLICT, 0, 0
	}

	dealId := 0
	err = tx.QueryRow(CreateDeal, adId, subscriberId).Scan(&dealId)
	if err != nil {
		return DB_ERROR, 0, 0
	}
	_ = tx.Commit()
	return CREATED, dealId, subscriberId
}

func (db *DB) FulfillDeal(dealId int, userId int) int {
	// todo: check if current_carma or frozen_carma can be less then zero
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
	_, err = tx.Exec(FulfillDeal, dealId, global_constants.PriceCoeff)
	if err != nil {
		return DB_ERROR
	}
	err = tx.Commit()
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) CancelDeal(dealId int, userId int) (int, models.CancelInfo) {
	tx, err := db.StartTransaction()
	if err != nil {
		return DB_ERROR, models.CancelInfo{}
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
		return EMPTY_RESULT, models.CancelInfo{}
	}
	if err != nil {
		return DB_ERROR, models.CancelInfo{}
	}
	if status != "open" {
		return FORBIDDEN, models.CancelInfo{}
	}
	if subscriberId == userId {
		_, err = tx.Exec(CancelDealSubscriber, dealId, global_constants.PriceCoeff)
		if err != nil {
			return DB_ERROR, models.CancelInfo{}
		}
		err = tx.Commit()
		if err != nil {
			return DB_ERROR, models.CancelInfo{}
		}
		return OK, models.CancelInfo{WhomId: authorId, CancelType: "subscriber", AdId: adId}
	}

	if authorId == userId {
		_, err = tx.Exec(CancelDealAuthor, dealId)
		if err != nil {
			return DB_ERROR, models.CancelInfo{}
		}
		err = tx.Commit()
		if err != nil {
			return DB_ERROR, models.CancelInfo{}
		}
		return OK, models.CancelInfo{WhomId: subscriberId, CancelType: "author", AdId: adId}
	}
	return FORBIDDEN, models.CancelInfo{}
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

func (db *DB) GetDealById(dealId int) (models.DealDetails, int) {
	deal := models.DealDetails{}
	err := db.db.QueryRow(GetDeal, dealId).Scan(&deal.DealId, &deal.AdId, &deal.SubscriberId, &deal.Status)
	if err == pgx.ErrNoRows {
		return deal, EMPTY_RESULT
	}
	if err != nil {
		return deal, DB_ERROR
	}
	return deal, FOUND
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
		err = rows.Scan(&user.VkId, &user.Name, &user.Surname, &user.PhotoUrl)
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

func (db *DB) GetAllAdSubscribersIDs(adId int) ([]int, error) {
	rows, err := db.db.Query(GetAdSubscribersIds, adId)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0)
	defer rows.Close()
	for rows.Next() {
		id := 0
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (db *DB) GetSubscriberIdForDeal(adId int, dealType string, params url.Values) (int, int) {
	choicer := map[string]func(values url.Values) (int, int){
		"auction": func(values url.Values) (int, int) {
			subscriberId := 0
			err := db.db.QueryRow(GetRichest, adId).Scan(&subscriberId)
			if err != nil {
				return 0, DB_ERROR
			}
			return subscriberId, OK
		},
		"random": func(values url.Values) (int, int) {
			rows, err := db.db.Query(GetAdSubscribersIds, adId)
			if err != nil {
				return 0, DB_ERROR
			}
			defer func() {
				rows.Close()
			}()
			ids := make([]int, 0)
			for rows.Next() {
				id := 0
				err = rows.Scan(&id)
				if err != nil {
					return 0, DB_ERROR
				}
				ids = append(ids, id)
			}
			length := len(ids)
			index := rand.Intn(length)
			return ids[index], OK
		},
		"choice": func(values url.Values) (int, int) {
			subscriberArr, ok := params["subscriber_id"]
			if !ok || len(subscriberArr) != 1 {
				return 0, WRONG_INPUT
			}
			subscriberId, err := strconv.Atoi(subscriberArr[0])
			if err != nil {
				return 0, WRONG_INPUT
			}
			return subscriberId, OK
		},
	}
	fun, ok := choicer[dealType]
	if !ok {
		return 0, WRONG_INPUT
	}
	return fun(params)
}
