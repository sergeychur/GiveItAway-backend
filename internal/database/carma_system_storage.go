package database

import (
	"fmt"
	"github.com/sergeychur/give_it_away/internal/global_constants"
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
)

const (
	ZerofyIfExceeded = "UPDATE users_carma SET cost_frozen = 1, casback_frozen = 1," +
		" last_updated = (now() at time zone 'utc') WHERE user_id = $1" +
		" AND last_updated - (now() at time zone 'utc') >= '%s'::interval AND frozen_carma = 0"
	checkIfAuction        = "SELECT ad_type = 'auction' FROM ad where ad_id = $1"
	checkIfEnoughCarma    = "SELECT current_carma - frozen_carma >= $1 * cost_frozen FROM users_carma WHERE user_id = $2"

	getMaxBidInAuction = "select bid from ad_subscribers where ad_id = $1 order by bid desc limit 1" // todo: повесить индекс
	getMaxBidInAuctionWithUserId = "select bid, subscriber_id from ad_subscribers where ad_id = $1 order by bid desc limit 1"
	getMaxBidUserInAuction = "select a_s.bid, u.vk_id, u.name, u.surname, u.photo_url from ad_subscribers a_s" +
		" join users u on (u.vk_id = a_s.subscriber_id) where a_s.ad_id = $1 order by a_s.bid desc limit 1"
	getUserBid = "select bid from ad_subscribers where ad_id = $1 and subscriber_id = $2"

	checkIfEnoughCarmaAuction = "SELECT current_carma - frozen_carma > $1 FROM users_carma WHERE user_id = $2"
	updateFrozenSubscribe = "UPDATE users_carma SET frozen_carma = frozen_carma + $1 * cost_frozen WHERE user_id = $2 " +
		"RETURNING $1 * cost_frozen"
	updateFrozenSubscribeAuct = "UPDATE users_carma SET frozen_carma = frozen_carma + $1 WHERE user_id = $2"
	updateCostFrozenSubscribe = "UPDATE users_carma SET cost_frozen = cost_frozen + 1 WHERE user_id = $1"

	updateCarmaUnsubscribe = "UPDATE users_carma SET  cost_frozen = cost_frozen - 1, frozen_carma = frozen_carma - (cost_frozen - 1) * $1" +
		"WHERE user_id = $2"

	GetCarmaToReturnUnsubscribe = "SELECT bid FROM ad_subscribers WHERE ad_id = $1 AND subscriber_id = $2"
	updateCarmaUnsubscribeAuct = "UPDATE users_carma SET frozen_carma = frozen_carma - $1 WHERE user_id = $2"

	updateCarmaDeleteNonAuct = "UPDATE users_carma SET cost_frozen = cost_frozen -1, frozen_carma = frozen_carma - (cost_frozen - 1) * $1 " +
		"WHERE user_id IN (SELECT subscriber_id FROM ad_subscribers WHERE ad_id = $2);"
	updateCarmaDeleteAuct = "UPDATE users_carma SET frozen_carma = frozen_carma - a_s.bid FROM ad_subscribers a_s " +
		"WHERE users_carma.user_id = a_s.subscriber_id and a_s.ad_id = $1"

	GetUserCostFreeze = "SELECT cost_frozen FROM users_carma WHERE user_id = $1"

	updateBid = "update ad_subscribers set bid=$1 where ad_id=$2 and subscriber_id=$3"

)

func (db *DB) DealWithCarmaSubscribe(tx *pgx.Tx, adId, userId int) (bool, int, error, *models.Notification){
	_, err := tx.Exec(fmt.Sprintf(ZerofyIfExceeded, global_constants.ZeroingTime), userId)
	if err != nil {
		return false, 0, err, nil
	}
	isAuction := false
	err = tx.QueryRow(checkIfAuction, adId).Scan(&isAuction)
	if err != nil {
		return false, 0, err, nil
	}
	if isAuction {
		return db.DealWithCarmaSubscribeAuct(tx, adId, userId)
	}
	canSubscribe, carma, err := db.DealWithCarmaSubscribeNonAuct(tx, adId, userId)
	return canSubscribe, carma, err, nil
}

func (db *DB) DealWithCarmaSubscribeAuct(tx *pgx.Tx, adId, userId int) (bool, int, error, *models.Notification) {
	carmaForAuct := 0
	prevMaxBidId := 0
	err := tx.QueryRow(getMaxBidInAuctionWithUserId, adId).Scan(&carmaForAuct, &prevMaxBidId)
	if err == pgx.ErrNoRows {
		err = nil
		carmaForAuct = global_constants.InitialBid
	}
	if err != nil {
		return false, 0, err, nil
	}
	enoughCarma := false
	err = tx.QueryRow(checkIfEnoughCarmaAuction, carmaForAuct, userId).Scan(&enoughCarma)
	if err != nil {
		return false, 0, err, nil
	}
	if !enoughCarma {
		return false, 0, nil, nil
	}

	frozenCarma := carmaForAuct + 1
	_, err = tx.Exec(updateFrozenSubscribeAuct, frozenCarma, userId)
	if err != nil {
		return false, 0, err, nil
	}
	// took auction out of usual flow
	note, err := db.FormMaxBidUpdatedNote(adId, prevMaxBidId, frozenCarma, userId)
	if err != nil {
		return false, 0, err, nil
	}
	return true, frozenCarma, nil, &note
}

func (db *DB) DealWithCarmaSubscribeNonAuct(tx *pgx.Tx, adId, userId int) (bool, int, error) {
	enoughCarma := false
	err := tx.QueryRow(checkIfEnoughCarma, global_constants.PriceCoeff, userId).Scan(&enoughCarma)
	if err != nil {
		return false, 0, err
	}
	if !enoughCarma {
		return false, 0, nil
	}
	frozenCarma := 0
	err = tx.QueryRow(updateFrozenSubscribe, global_constants.PriceCoeff, userId).Scan(&frozenCarma)
	if err != nil {
		return false, 0, err
	}
	_, err = tx.Exec(updateCostFrozenSubscribe, userId)
	if err != nil {
		return false, 0, err
	}

	return true, frozenCarma, nil
}

func (db *DB) DealWithCarmaUnsubscribe(tx *pgx.Tx, adId, userId int) error {
	isAuction := false
	err := tx.QueryRow(checkIfAuction, adId).Scan(&isAuction)
	if err != nil {
		return err
	}
	if isAuction {
		return db.DealWithCarmaUnsubscribeAuct(tx, adId, userId)
	}
	return db.DealWithCarmaUnsubscribeNonAuct(tx, adId, userId)
}

func (db *DB) DealWithCarmaUnsubscribeAuct(tx *pgx.Tx, adId, userId int) error {
	priceToReturn := 0
	err := tx.QueryRow(GetCarmaToReturnUnsubscribe, adId, userId).Scan(&priceToReturn)
	_, err = tx.Exec(updateCarmaUnsubscribeAuct, priceToReturn, userId)
	return err
}

func (db *DB) DealWithCarmaUnsubscribeNonAuct(tx *pgx.Tx, adId, userId int) error {
	_, err := tx.Exec(updateCarmaUnsubscribe, global_constants.PriceCoeff, userId)
	return err
}

func (db *DB) GiveCarmaBackDelete(tx *pgx.Tx, adId, userId int) error {
	isAuction := false
	err := tx.QueryRow(checkIfAuction, adId).Scan(&isAuction)
	if err != nil {
		return err
	}
	if isAuction {

		_, err = tx.Exec(updateCarmaDeleteAuct, adId)
	} else {
		_, err = tx.Exec(updateCarmaDeleteNonAuct, global_constants.PriceCoeff, adId)
	}
	return err
}

func (db *DB) GetMaxBidForAd (adId int) (models.Bid, int) {
	bid := models.Bid{}
	err := db.db.QueryRow(getMaxBidInAuction, adId).Scan(&bid.Bid)
	if err == pgx.ErrNoRows {
		return models.Bid{Bid: 0}, FOUND
	}
	if err != nil {
		return models.Bid{}, DB_ERROR
	}
	return bid, FOUND
}

func (db *DB) GetMaxBidUserForAd (adId int) (models.BidUser, int) {
	bid := models.BidUser{}
	isAuction := false
	err := db.db.QueryRow(checkIfAuction, adId).Scan(&isAuction)
	if !isAuction {
		return models.BidUser{}, FORBIDDEN
	}
	err = db.db.QueryRow(getMaxBidUserInAuction, adId).Scan(&bid.Bid,
		&bid.User.VkId, &bid.User.Name, &bid.User.Surname, &bid.User.PhotoUrl)
	if err == pgx.ErrNoRows {
		return models.BidUser{}, EMPTY_RESULT
	}
	if err != nil {
		return models.BidUser{}, DB_ERROR
	}
	return bid, FOUND
}

func (db *DB) GetUserBidForAd(adId, userId int) (models.Bid, int) {
	isAuction := false
	err := db.db.QueryRow(checkIfAuction, adId).Scan(&isAuction)
	if err == pgx.ErrNoRows {
		return models.Bid{}, EMPTY_RESULT
	}
	if err != nil {
		return models.Bid{}, DB_ERROR
	}
	if isAuction {
		bid, status := db.GetMaxBidForAd(adId)
		bid.Bid += 1
		return bid, status
	}
	bid := models.Bid{}
	err = db.db.QueryRow(GetUserCostFreeze, userId).Scan(&bid.Bid)
	if err == pgx.ErrNoRows {
		return models.Bid{}, EMPTY_RESULT
	}
	if err != nil {
		return models.Bid{}, DB_ERROR
	}
	return bid, FOUND
}

func (db *DB) GetReturnBid(adId, userId int) (models.Bid, int) {
	bid := models.Bid{}
	err := db.db.QueryRow(getUserBid, adId, userId).Scan(&bid.Bid)
	if err == pgx.ErrNoRows {
		return models.Bid{}, EMPTY_RESULT
	}
	if err != nil {
		return models.Bid{}, DB_ERROR
	}
	return bid, FOUND
}

func (db *DB) IncreaseBid(adId, userId int) (models.Notification, int) {
	isAuction := false
	tx, err := db.StartTransaction()
	if err != nil {
		return models.Notification{}, DB_ERROR
	}
	defer func() {
		_ = tx.Rollback()
	}()
	err = tx.QueryRow(checkIfAuction, adId).Scan(&isAuction)
	if err == pgx.ErrNoRows {
		return models.Notification{}, EMPTY_RESULT
	}
	if err != nil {
		return models.Notification{}, DB_ERROR
	}
	if !isAuction {
		return models.Notification{}, FORBIDDEN
	}
	isSubscriber := false
	err = tx.QueryRow(CheckIfSubscriber, adId, userId).Scan(&isSubscriber)
	if err == pgx.ErrNoRows {
		return models.Notification{}, EMPTY_RESULT
	}
	if err != nil {
		return models.Notification{}, DB_ERROR
	}
	if !isSubscriber {
		return models.Notification{}, EMPTY_RESULT
	}
	maxBid := 0
	prevMaxId := 0
	err = tx.QueryRow(getMaxBidInAuctionWithUserId, adId).Scan(&maxBid, &prevMaxId)
	if err == pgx.ErrNoRows {
		maxBid = 0
	}
	enoughCarma := false
	err = tx.QueryRow(checkIfEnoughCarmaAuction, maxBid, userId).Scan(&enoughCarma)
	if !enoughCarma {
		return models.Notification{}, CONFLICT
	}
	_, err = tx.Exec(updateBid, maxBid+1, adId, userId)
	err = tx.Commit()
	if err != nil {
		return models.Notification{}, DB_ERROR
	}
	note, err := db.FormMaxBidUpdatedNote(adId, prevMaxId, maxBid + 1, userId)
	if err != nil {
		return models.Notification{}, DB_ERROR
	}
	return note, OK
}
