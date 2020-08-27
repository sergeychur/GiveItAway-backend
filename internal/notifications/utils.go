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
	COMMENT_CREATED      = "new_comment"
	MAX_BID_UPDATED      = "max_bid_upd"
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
			return &models.NewComment{}
		},
		MAX_BID_UPDATED: func() interface{} {
			return &models.MaxBidUpdated{}
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

const ERR_FORM = "CANT_CONVERT"
const ERR_UNKNOWN_TYPE = "Unknown type"

func AdToText(m models.AdForNotification) string {
	var header = m.Header
	if header != "" {
		header = "'" + header + "'"
	}
	return fmt.Sprintf("%s (https://vk.com/app7360033_45863670#%d)", header, m.AdId)

}

func AuthorToText(user models.User) string {
	return fmt.Sprintf("@id%d(%s %s)", user.VkId, user.Name, user.Surname)

}

func NotificationToText(m models.Notification) (string, error) {
	var text string
	switch v := m.Payload.(type) {
	case models.NewComment:
		{
			var comment = m.Payload.(models.NewComment)

			var author = comment.Comment.Author
			text = "Новый комментарий под объявлением " + AdToText(comment.Ad) + "\n"
			text += AuthorToText(author) + " : '" + comment.Comment.Text + "'"
			break
		}
	case *models.UserSubscribed:
		{
			var user = m.Payload.(*models.UserSubscribed)
			var author = user.Author
			text = "Кое-кто (" +
				AuthorToText(author) +
				") откликнулся на объявление " +
				AdToText(user.Ad) + "\n"
			break
		}
	case *models.MaxBidUpdated:
		{
			var maxBid = m.Payload.(*models.MaxBidUpdated)
			text = "Новая максимальная ставка в объявлении" +
				AdToText(models.AdForNotification{AdId: m.AdId}) + "\n"
			text += fmt.Sprintf("%s: %d💰", AuthorToText(maxBid.User), maxBid.NewBid)
			break
		}
	case models.AdStatusChanged:
		switch m.NotificationType {

		case STATUS_CHANGED:
			{
				var status = m.Payload.(models.AdStatusChanged)
				switch status.Ad.Status {
				case "chosen": //!! не смог найти константу
					text = "Автор выбрал получателя в объявлении " + AdToText(status.Ad)
					break
				case "closed": //!! не смог найти константу
					text = "Передача вещи в объявлении " + AdToText(status.Ad) + " проведена"
					break
				}
				break
			}

		case AD_DELETED:
			{
				var status = m.Payload.(models.AdStatusChanged)
				text = "Объявление " + AdToText(status.Ad) + " удалено."
				break
			}
		case DEAL_FULFILL:
			{
				var ad = m.Payload.(models.AdStatusChanged)
				text = "Получатель из объявления " + AdToText(ad.Ad) + " подтвердил получение вещи."
				break
			}
		}
		break
	case *models.AuthorClosedAd:
		var ad = m.Payload.(*models.AuthorClosedAd)
		text = "Автор объявления " + AdToText(ad.Ad) + " предлагает вам забрать вещь. Свяжитесь" +
			" с ним для уточнения деталей! Не забудьте подтведить получение вещи в приложении " +
			"после того как заберёте её."
		break
	case models.AuthorCancelled:
		var ad = m.Payload.(models.AuthorCancelled)
		text = "Автор объявления " + AdToText(ad.Ad) + " отменил запрос на передачу вещи вам."
		break
	case models.SubscriberCancelled:
		var ad = m.Payload.(models.SubscriberCancelled)

		text = "Получатель из объявления " + AdToText(ad.Ad) + " отказался подтвердить получение вещи."
		break
	default:
		err := fmt.Errorf("Unknown type: %T", v)
		fmt.Printf("error is %e", err)
		return "", err
	}
	return text, nil
}
