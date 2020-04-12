package server

import (
	"fmt"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/models"
	"log"
	"net/http"
)

func (server *Server) MakeDealSendUpd(dealId, initiatorId, subscriberId, adId int, r *http.Request) {
	notification, err := server.db.FormAdClosedNotification(dealId, initiatorId, subscriberId)
	if err == nil {
		err = server.db.InsertNotification(subscriberId, notification)
		// TODO :done
		server.NotificationSender.SendOneClient(r.Context(), notification, subscriberId)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	notifications, err := server.db.FormStatusChangedNotificationsByDeal(dealId)
	if err == nil {
		// TODO:done
		server.NotificationSender.SendAllNotifications(r.Context(), notifications)
		err = server.db.InsertNotifications(notifications)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}

	deal, dealGetStatus := server.db.GetDealById(dealId)
	if dealGetStatus == database.FOUND {
		upd := FormDealcreatedUpdate(deal)
		server.NotificationSender.SendToChannel(r.Context(), upd, fmt.Sprintf("ad_%d", adId))
	}

}

func (server *Server) FulFillDealSendUpd(dealId int, notifications []models.Notification, r *http.Request) {
	notification, err := server.db.FormFulfillDealNotification(dealId)
	if err == nil {
		// TODO(FULFILL): done
		server.NotificationSender.SendOneClient(r.Context(), notification, notification.WhomId)
		err = server.db.InsertNotification(notification.WhomId, notification)
		if err != nil {
			log.Println(err)
		}
	}

	if err == nil {
		// TODO(FULFILL): done
		server.NotificationSender.SendAllNotifications(r.Context(), notifications)
		err = server.db.InsertNotifications(notifications)
	} else {
		log.Println(err)
	}
}

func (server *Server) CancelDealSendUpd(err error, cancelInfo models.CancelInfo, userId int,
	notifications []models.Notification, r *http.Request) {
	if err == nil {
		// TODO(CANCEL): done
		server.NotificationSender.SendAllNotifications(r.Context(), notifications)
		err = server.db.InsertNotifications(notifications)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	note, err := server.db.FormCancelNotification(cancelInfo.CancelType, userId, cancelInfo.AdId)
	if err == nil {
		// TODO(CANCEL): done
		server.NotificationSender.SendOneClient(r.Context(), note, cancelInfo.WhomId)
		err = server.db.InsertNotification(cancelInfo.WhomId, note)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
}

func (server *Server) SubscribeToAdSendUpd(userId, adId int, r *http.Request) {
	notification, err := server.db.FormRespondNotification(userId, adId)
	if err == nil {
		err = server.db.InsertNotification(notification.WhomId, notification)
		// todo: done
		if err != nil {
			log.Println(err)
		} else {

			server.NotificationSender.SendOneClient(r.Context(), notification, notification.WhomId)
			// todo: maybe remake, talk with Artyom
			newSubUpd := FormNewSubscriberUpdate(notification)
			if newSubUpd != nil {
				server.NotificationSender.SendToChannel(r.Context(), *newSubUpd, fmt.Sprintf("ad_%d", adId))
			}
		}
	}
}