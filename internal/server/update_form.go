package server

import "github.com/sergeychur/give_it_away/internal/models"

const (
	NEW_COMMENT = "new_comment"
	EDIT_COMMENT = "edit_comment"
	DELETE_COMMENT = "delete_comment"
	EDIT_AD = "edit_ad"
	NEW_SUBSCRIBER = "new_subscriber"
	AD_CLOSE = "ad_close"
	STATUS_CHANGED = "status_changed"
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

func FormEditAdUpdate(ad models.AdForUsersDetailed) models.AdUpdate {
	return models.AdUpdate{
		Payload: ad,
		Type: EDIT_AD,
	}
}

func FormNewSubscriberUpdate(note models.Notification) *models.AdUpdate {
	subscribed, ok := note.Payload.(models.UserSubscribed)
	if !ok {
		return nil
	}
	user := subscribed.Author
	return &models.AdUpdate{
		Payload: user,
		Type: NEW_SUBSCRIBER,
	}
}

func FormDealcreatedUpdate(deal models.DealDetails) models.AdUpdate {
	return models.AdUpdate{
		Payload: deal,
		Type: AD_CLOSE,
	}
}

func FormFulfillDealUpdate(note models.Notification) (*models.AdUpdate, int) {
	fulfillInfo, ok := note.Payload.(*models.AdStatusChanged)
	if !ok {
		return nil, 0
	}

	newStatus := models.AdNewStatus{
		NewStatus: fulfillInfo.Ad.Status,
	}

	return &models.AdUpdate{
		Payload: newStatus,
		Type: STATUS_CHANGED,
	}, int(fulfillInfo.Ad.AdId)
}

func FormCancelDealUpdate(note models.Notification) (*models.AdUpdate, int) {
	cancelInfoAuthor, ok := note.Payload.(models.AuthorCancelled)
	if ok {
		newStatus := models.AdNewStatus{
			NewStatus: cancelInfoAuthor.Ad.Status,
		}

		return &models.AdUpdate{
			Payload: newStatus,
			Type: STATUS_CHANGED,
		}, int(cancelInfoAuthor.Ad.AdId)
	} else {
		cancelInfoSubscriber, ok := note.Payload.(models.SubscriberCancelled)
		if ok {
			newStatus := models.AdNewStatus{
				NewStatus: cancelInfoSubscriber.Ad.Status,
			}

			return &models.AdUpdate{
				Payload: newStatus,
				Type: STATUS_CHANGED,
			}, int(cancelInfoSubscriber.Ad.AdId)
		}
	}
	return nil, 0
}
