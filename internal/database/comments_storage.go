package database

import (
	"time"

	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
)

const (
	GetAdComments = "SELECT c.comment_id, c.creation_datetime, c.text, u.vk_id, u.name, u.surname, u.photo_url " +
		"FROM comment c JOIN (SELECT comment_id FROM comment WHERE ad_id = $1 ORDER BY comment_id " +
		"LIMIT $2 OFFSET $3) v ON (v.comment_id = c.comment_id) JOIN users u ON (u.vk_id = c.author_id) " +
		"ORDER BY c.comment_id"

	CreateComment = "INSERT INTO COMMENT (ad_id, text, author_id) VALUES ($1, $2, $3) RETURNING comment_id"
	GetComment    = "SELECT c.comment_id, c.creation_datetime, c.text, u.vk_id, u.name, u.surname, u.photo_url " +
		"FROM comment c JOIN users u ON (c.author_id = u.vk_id) WHERE c.comment_id = $1"

	CheckCommentExists = "SELECT author_id FROM comment WHERE comment_id = $1"

	UpdateComment  = "UPDATE comment SET text = $2 WHERE comment_id = $1"
	DeleteComment  = "DELETE FROM comment WHERE comment_id = $1"
	GetCommentAdId = "SELECT ad_id FROM comment WHERE comment_id = $1"
)

func (db *DB) GetComments(adId int, page int, rowsPerPage int) ([]models.CommentForUser, int) {
	offset := rowsPerPage * (page - 1)
	rows, err := db.db.Query(GetAdComments, adId, rowsPerPage, offset)
	if err == pgx.ErrNoRows {
		return nil, EMPTY_RESULT
	}
	if err != nil {
		return nil, DB_ERROR
	}
	comments := make([]models.CommentForUser, 0)
	defer rows.Close()
	for rows.Next() {
		timeStamp := time.Time{}
		comment := models.CommentForUser{}
		err = rows.Scan(&comment.CommentId, &timeStamp, &comment.Text, &comment.Author.VkId,
			&comment.Author.Name, &comment.Author.Surname, &comment.Author.PhotoUrl)
		if err != nil {
			return nil, DB_ERROR
		}
		loc, _ := time.LoadLocation("UTC")
		timeStamp.In(loc)
		comment.CreationDateTime = timeStamp.Format("02 Jan 06 15:04 UTC")
		comments = append(comments, comment)
	}

	if len(comments) == 0 {
		return comments, EMPTY_RESULT
	}
	return comments, FOUND
}

func (db *DB) CreateComment(adId int, userId int, comment models.Comment) (models.CommentForUser, int) {
	exists := false
	err := db.db.QueryRow(checkUserExists, userId).Scan(&exists)
	if err == pgx.ErrNoRows || !exists {
		return models.CommentForUser{}, EMPTY_RESULT
	} // TODO: mb remove, user_id taken from cookie, useless
	authorId := 0
	err = db.db.QueryRow(checkAdExist, adId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return models.CommentForUser{}, EMPTY_RESULT
	}
	commentId := 0
	err = db.db.QueryRow(CreateComment, adId, comment.Text, userId).Scan(&commentId)
	if err != nil {
		return models.CommentForUser{}, DB_ERROR
	}
	retVal := models.CommentForUser{}
	retVal.Author = models.User{}
	timeStamp := time.Time{}
	err = db.db.QueryRow(GetComment, commentId).Scan(&retVal.CommentId, &timeStamp, &retVal.Text,
		&retVal.Author.VkId, &retVal.Author.Name, &retVal.Author.Surname, &retVal.Author.PhotoUrl)
	if err != nil {
		return models.CommentForUser{}, DB_ERROR
	}
	loc, _ := time.LoadLocation("UTC")
	timeStamp.In(loc)
	retVal.CreationDateTime = timeStamp.Format("02 Jan 06 15:04 UTC")
	return retVal, CREATED
}

func (db *DB) EditComment(commentId int, userId int, comment models.Comment) (models.CommentForUser, int) {
	authorId := 0
	err := db.db.QueryRow(CheckCommentExists, commentId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return models.CommentForUser{}, EMPTY_RESULT
	}

	if authorId != userId {
		return models.CommentForUser{}, FORBIDDEN
	}

	_, err = db.db.Exec(UpdateComment, commentId, comment.Text)
	if err != nil {
		return models.CommentForUser{}, DB_ERROR
	}
	retVal := models.CommentForUser{}
	retVal.Author = models.User{}
	timeStamp := time.Time{}
	err = db.db.QueryRow(GetComment, commentId).Scan(&retVal.CommentId, &timeStamp, &retVal.Text,
		&retVal.Author.VkId, &retVal.Author.Name, &retVal.Author.Surname, &retVal.Author.PhotoUrl)
	if err != nil {
		return models.CommentForUser{}, DB_ERROR
	}
	loc, _ := time.LoadLocation("UTC")
	timeStamp.In(loc)
	retVal.CreationDateTime = timeStamp.Format("02 Jan 06 15:04 UTC")
	return retVal, CREATED
}

func (db *DB) DeleteComment(commentId int, userId int) int {
	authorId := 0
	err := db.db.QueryRow(CheckCommentExists, commentId).Scan(&authorId)
	if err == pgx.ErrNoRows {
		return EMPTY_RESULT
	}

	var allow = false
	for id := range WHITE_LIST {
		if userId == id {
			allow = true
		}
	}
	if !allow && authorId != userId {
		return FORBIDDEN
	}

	_, err = db.db.Exec(DeleteComment, commentId)
	if err != nil {
		return DB_ERROR
	}
	return OK
}

func (db *DB) GetAdIdForComment(commentId int) (int, error) {
	adId := 0
	err := db.db.QueryRow(GetCommentAdId, commentId).Scan(&adId)
	if err != nil {
		return 0, err
	}
	return adId, nil
}
