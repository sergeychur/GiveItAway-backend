package server

import "github.com/sergeychur/give_it_away/internal/models"

const (
	NEW_COMMENT = "new_comment"
	EDIT_COMMENT = "edit_comment"
	DELETE_COMMENT = "delete_comment"
)

func FormNewCommentUpdate(comment models.CommentForUser) models.AdUpdate {
	return models.AdUpdate{
		Payload: &comment,
		Type: NEW_COMMENT,
	}
}

func FormEditCommentUpdate(comment models.CommentForUser) models.AdUpdate {
	return models.AdUpdate{
		Payload: &comment,
		Type: EDIT_COMMENT,
	}
}

func FormDeleteCommentUpdate() models.AdUpdate {
	return models.AdUpdate{
		Payload: nil,
		Type: DELETE_COMMENT,
	}
}
