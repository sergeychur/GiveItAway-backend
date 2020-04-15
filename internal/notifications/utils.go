package notifications

import (
	"encoding/json"
	"fmt"
	"github.com/sergeychur/give_it_away/internal/models"
)

const (
	AD_CLOSE             = "ad_close"
	AD_RESPOND           = "respond"
	DEAL_FULFILL         = "fulfill"
	STATUS_CHANGED       = "status"
	AD_DELETED           = "deleted"
	AUTHOR_CANCELLED     = "authorCancel"
	SUBSCRIBER_CANCELLED = "subscriberCancel"
	COMMENT_CREATED = "new_comment"
)

var (
	FuncsMap = map[string]func() interface{}{
		AD_CLOSE: func() interface{} {
			return &models.AuthorClosedAd{}
		},
		AD_RESPOND: func() interface{} {
			return &models.UserSubscribed{}
		},
		DEAL_FULFILL: func() interface{} {
			return &models.AdStatusChanged{}
		},
		STATUS_CHANGED: func() interface{} {
			return &models.AdStatusChanged{}
		},
		AD_DELETED: func() interface{} {
			return &models.AdStatusChanged{}
		},
		AUTHOR_CANCELLED: func() interface{} {
			return &models.AuthorCancelled{}
		},
		SUBSCRIBER_CANCELLED: func() interface{} {
			return &models.SubscriberCancelled{}
		},
		COMMENT_CREATED: func() interface{} {
			return &models.CommentForUser{}
		},
	}
)

func FormPayLoad(payload []byte, notificationType string) (interface{}, error) {
	neededFunc, ok := FuncsMap[notificationType]
	if !ok {
		return nil, fmt.Errorf("unable to detect notification type")
	}
	val := neededFunc()
	err := json.Unmarshal(payload, &val)
	return val, err
}
