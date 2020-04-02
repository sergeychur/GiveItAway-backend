package database

import (
	"encoding/json"
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
	GetAdTitleAndAuthorForDeal = "SELECT a.header, u.vk_id, u.name, u.surname, u.photo_url " +
		"FROM ad a JOIN users u ON (a.author_id = u.vk_id) JOIN deal d ON (d.ad_id = a.ad_id) WHERE d.deal_id = $1"
	InsertNotification = "INSERT INTO notifications (user_id, notification_type, payload, creation_datetime) VALUES ($1, $2, $3, $4)"
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
		notification.CreationDateTime = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
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
	note.CreationDateTime = time.Now().Format("2006-01-02T15:04:05.999999999Z07:00")
	note.IsRead = false
	val := &models.AuthorClosedAd{}
	val.DealId = dealId
	err := db.db.QueryRow(GetAdTitleAndAuthorForDeal, dealId).Scan(&val.AdTitle, &val.Author.VkId,
		&val.Author.Name, &val.Author.Surname, &val.Author.PhotoUrl)
	if err != nil {
		return models.Notification{}, err
	}
	note.Payload = val
	return note, nil
}

func (db *DB) InsertNotification(whomId int, notification models.Notification) error {
	creation, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", notification.CreationDateTime)
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