package database

import (
	"fmt"
	"github.com/sergeychur/give_it_away/internal/global_constants"
	"gopkg.in/jackc/pgx.v2"
)

const (
	ZerofyIfExceeded = "UPDATE users_carma SET cost_frozen = 1, casback_frozen = 1," +
		" last_updated = (now() at time zone 'utc') WHERE user_id = $1" +
		" AND last_updated - (now() at time zone 'utc') >= '%s'::interval AND frozen_carma = 0"
	checkIfAuction        = "SELECT is_auction FROM ad where ad_id = $1"
	checkIfEnoughCarma    = "SELECT current_carma - frozen_carma >= $1 * cost_frozen FROM users_carma WHERE user_id = $2"

	getMaxBidInAuction = "select bid from ad_subscribers where ad_id = $1 order by bid desc limit 1" // todo: повесить индекс

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

)

func (db *DB) DealWithCarmaSubscribe(tx *pgx.Tx, adId, userId int) (bool, int, error){
	_, err := tx.Exec(fmt.Sprintf(ZerofyIfExceeded, global_constants.ZeroingTime), userId)
	if err != nil {
		return false, 0, err
	}
	isAuction := false
	err = tx.QueryRow(checkIfAuction, adId).Scan(&isAuction)
	if err != nil {
		return false, 0, err
	}
	if isAuction {
		return db.DealWithCarmaSubscribeAuct(tx, adId, userId)
	}
	return db.DealWithCarmaSubscribeNonAuct(tx, adId, userId)
}

func (db *DB) DealWithCarmaSubscribeAuct(tx *pgx.Tx, adId, userId int) (bool, int, error) {
	carmaForAuct := 0
	err := tx.QueryRow(getMaxBidInAuction, adId).Scan(&carmaForAuct)
	if err != nil {
		return false, 0, err
	}
	enoughCarma := false
	err = tx.QueryRow(checkIfEnoughCarmaAuction, carmaForAuct, userId).Scan(&enoughCarma)
	if err != nil {
		return false, 0, err
	}
	if !enoughCarma {
		return false, 0, nil
	}

	frozenCarma := 0
	err = tx.QueryRow(updateFrozenSubscribeAuct, carmaForAuct, userId).Scan(&frozenCarma)
	if err != nil {
		return false, 0, err
	}
	// took auction out of usual flow
	return true, frozenCarma, nil
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
