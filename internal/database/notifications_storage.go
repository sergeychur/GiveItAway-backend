package database

import (
	"encoding/json"
	"fmt"
	"github.com/sergeychur/give_it_away/internal/models"
	"github.com/sergeychur/give_it_away/internal/notifications"
	"gopkg.in/jackc/pgx.v2"
	"time"
)

const (
	GetNotifications = "SELECT n.notification_id, n.notification_type, n.creation_datetime, n.payload, " +
		"n.is_read FROM notifications n JOIN (SELECT notification_id FROM notifications" +
		" WHERE user_id = $1 ORDER BY is_read, notification_id DESC LIMIT $2 OFFSET $3) l" +
		" ON (n.notification_id = l.notification_id) ORDER BY n.is_read, n.notification_id DESC"

	SetReadTrue = "UPDATE notifications SET is_read = true WHERE notification_id = $1"

	GetAdNotifWithUserAndDeal = "SELECT a.ad_id, a.header, a.status, u.vk_id, u.name, u.surname, u.photo_url " +
		"FROM ad a JOIN users u ON (a.author_id = u.vk_id) JOIN deal d ON (d.ad_id = a.ad_id) WHERE d.deal_id = $1"

	InsertNotification = "INSERT INTO notifications (user_id, notification_type, payload, creation_datetime) VALUES ($1, $2, $3, $4)"

	GetAdForNotif = "SELECT a.ad_id, a.header, a.status, a.author_id FROM ad a WHERE a.ad_id = $1"

	GetDealAdInfoById = "SELECT d.ad_id FROM deal d JOIN ad a ON (a.ad_id = d.ad_id) WHERE deal_id = $1"
)

func (db *DB) GetNotifications(userId int, page int, rowsPerPage int) ([]models.Notification, int) {
	offset := rowsPerPage * (page - 1)
	rows, err := db.db.Query(GetNotifications, userId, rowsPerPage, offset)
	if err == pgx.ErrNoRows {
		return nil, EMPTY_RESULT
	}
	if err != nil {
		return nil, DB_ERROR
	}
	defer func() {
		rows.Close()
	}()
	notificationArr := make([]models.Notification, 0)
	for rows.Next() {
		notification := models.Notification{}
		timeStamp := time.Time{}
		var payload []byte
		id := 0
		err = rows.Scan(&id, &notification.NotificationType, &timeStamp, &payload, &notification.IsRead)
		if err != nil {
			return nil, DB_ERROR
		}
		notification.CreationDateTime = timeStamp.Format("01.02.2006 15:04")
		notification.Payload, err = notifications.FormPayLoad(payload, notification.NotificationType)
		if err != nil {
			return nil, DB_ERROR
		}
		_, err = db.db.Exec(SetReadTrue, id)
		if err != nil {
			return nil, DB_ERROR
		}
		notificationArr = append(notificationArr, notification)
	}
	return notificationArr, FOUND
}

func (db *DB) FormAdClosedNotification(dealId int, initiatorId int, subscriberId int) (models.Notification, error) {
	note := models.Notification{}
	note.NotificationType = notifications.AD_CLOSE
	note.CreationDateTime = time.Now().Format("01.02.2006 15:04")
	note.IsRead = false
	val := &models.AuthorClosedAd{}
	val.DealId = dealId
	err := db.db.QueryRow(GetAdNotifWithUserAndDeal, dealId).Scan(&val.Ad.AdId, &val.Ad.Header, &val.Ad.Status, &val.Author.VkId,
		&val.Author.Name, &val.Author.Surname, &val.Author.PhotoUrl)
	if err != nil {
		return models.Notification{}, err
	}

	val.Ad.PathesToPhoto, err = db.GetAdPhotos(val.Ad.AdId)
	if err != nil {
		return models.Notification{}, err
	}
	note.Payload = val
	return note, nil
}

func (db *DB) FormRespondNotification(subscriberId int, adId int) (models.Notification, error) {
	note := models.Notification{}
	note.NotificationType = notifications.AD_RESPOND
	note.CreationDateTime = time.Now().Format("01.02.2006 15:04")
	note.IsRead = false
	val := &models.UserSubscribed{}
	err := db.db.QueryRow(GetAdForNotif, adId).Scan(&val.Ad.AdId, &val.Ad.Header, &val.Ad.Status, &note.WhomId)
	if err != nil {
		return models.Notification{}, err
	}
	val.Ad.PathesToPhoto, err = db.GetAdPhotos(val.Ad.AdId)
	user, status := db.GetUser(subscriberId)
	if status == DB_ERROR {
		return models.Notification{}, fmt.Errorf("get user failed")
	}
	val.Author = user
	note.Payload = val
	return note, nil
}

func (db *DB) FormStatusChangedNotification(adId int, isDeleted bool, noteType string) (models.Notification, error) {
	note := models.Notification{}
	note.NotificationType = noteType
	note.CreationDateTime = time.Now().Format("01.02.2006 15:04")
	note.IsRead = false
	val := models.AdStatusChanged{}
	err := db.db.QueryRow(GetAdForNotif, adId).Scan(&val.Ad.AdId, &val.Ad.Header, &val.Ad.Status, &note.WhomId)
	if err == pgx.ErrNoRows {
		return models.Notification{}, fmt.Errorf("no ad")
	}
	if err != nil {
		return models.Notification{}, err
	}
	if !isDeleted {
		val.Ad.PathesToPhoto, err = db.GetAdPhotos(val.Ad.AdId)
	}
	note.Payload = val
	return note, nil
}

func (db *DB) FormFulfillDealNotification(dealId int) (models.Notification, error) {
	adId := 0
	err := db.db.QueryRow(GetDealAdInfoById, dealId).Scan(&adId)
	if err != nil {
		return models.Notification{}, err
	}
	note, err := db.FormStatusChangedNotification(adId, false, notifications.DEAL_FULFILL)
	if err != nil {
		return models.Notification{}, err
	}
	return note, nil
}

func (db *DB) FormStatusChangedNotificationsByDeal(dealId int) ([]models.Notification, error) {
	adId := 0
	err := db.db.QueryRow(GetDealAdInfoById, dealId).Scan(&adId)
	if err != nil {
		return nil, err
	}
	subscriberIds, err := db.GetAllAdSubscribersIDs(adId)
	if err != nil {
		return nil, err
	}
	note, err := db.FormStatusChangedNotification(adId, false, notifications.STATUS_CHANGED)
	if err != nil {
		return nil, err
	}
	notes := make([]models.Notification, 0)
	for _, subscriberId := range subscriberIds {
		curNote := note
		curNote.WhomId = subscriberId
		notes = append(notes, curNote)
	}
	return notes, nil
}

func (db *DB) FormStatusChangedNotificationsByAd (adId int, isDeleted bool, noteType string) ([]models.Notification, error) {
	subscriberIds, err := db.GetAllAdSubscribersIDs(adId)
	if err != nil {
		return nil, err
	}
	if len(subscriberIds) == 0 {
		return nil, fmt.Errorf("no subscribers")
	}
	note, err := db.FormStatusChangedNotification(adId, isDeleted, noteType)
	if err != nil {
		return nil, err
	}
	notes := make([]models.Notification, 0)
	for _, subscriberId := range subscriberIds {
		curNote := note
		curNote.WhomId = subscriberId
		notes = append(notes, curNote)
	}
	return notes, nil
}

func (db *DB) InsertNotification(whomId int, notification models.Notification) error {
	creation, err := time.Parse("01.02.2006 15:04", notification.CreationDateTime)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(notification.Payload)
	if err != nil {
		return err
	}
	_, err = db.db.Exec(InsertNotification, whomId, notification.NotificationType, payload, creation)
	return err
}

func (db *DB) InsertNotifications(notes []models.Notification) error {
	stmt, err := db.db.Prepare("insert_notif", InsertNotification)
	if err != nil {
		return err
	}
	for _, notification := range notes {
		creation, err := time.Parse("01.02.2006 15:04", notification.CreationDateTime)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(notification.Payload)
		if err != nil {
			return err
		}
		_, err = db.db.Exec(stmt.Name, notification.WhomId, notification.NotificationType, payload, creation)
		if err != nil {
			return err
		}
	}
	return nil
}