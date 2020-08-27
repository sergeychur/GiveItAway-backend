package database

import (
	"github.com/sergeychur/give_it_away/internal/models"
	"github.com/sergeychur/give_it_away/internal/notifications"
	"time"
)

const (
	GetUnreadNotesNumber = "SELECT COUNT(*) FROM notifications WHERE is_read = false AND user_id = $1"
)

func (db *DB) FormNewCommentNotif(comment models.CommentForUser, adId int) (models.Notification, error) {
	note := models.Notification{}
	timeStamp := time.Now()
	loc, _ := time.LoadLocation("UTC")
	timeStamp.In(loc)
	note.CreationDateTime = timeStamp.Format("02 Jan 06 15:04 UTC")
	note.IsRead = false
	authorId := 0
	err := db.db.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err != nil {
		return models.Notification{}, err
	}
	ad := models.AdForNotification{}
	whomId := 0
	err = db.db.QueryRow(GetAdForNotif, adId).Scan(&ad.AdId, &ad.Header, &ad.Status, &whomId)
	if err != nil {
		return models.Notification{}, err
	}
	note.Payload = models.NewComment{
		Ad:      ad,
		Comment: comment,
	}
	note.WhomId = authorId
	note.NotificationType = notifications.COMMENT_CREATED
	return note, nil
}

func (db *DB) GetUnreadNotesCount(userId int) (models.NotesNumber, int) {
	num := models.NotesNumber{}
	err := db.db.QueryRow(GetUnreadNotesNumber, userId).Scan(&num.Number)
	if err != nil {
		return models.NotesNumber{}, DB_ERROR
	}
	return num, FOUND
}
