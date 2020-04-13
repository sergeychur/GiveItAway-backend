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
	note.CreationDateTime = time.Now().Format("01.02.2006 15:04")
	note.IsRead = false
	authorId := 0
	err := db.db.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err != nil {
		return models.Notification{}, err
	}
	note.Payload = comment
	note.WhomId = authorId
	note.NotificationType = notifications.COMMENT_CREATED
	return note, nil
}

func (db *DB) GetUnreadNotesCount(userId int) (models.NotesNumber, int){
	num := models.NotesNumber{}
	err := db.db.QueryRow(GetUnreadNotesNumber, userId).Scan(&num.Number)
	if err != nil {
		return models.NotesNumber{}, DB_ERROR
	}
	return num, FOUND
}
