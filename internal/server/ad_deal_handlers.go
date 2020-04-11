package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/database"
	"log"
	"net/http"
	"strconv"
)

func (server *Server) SubscribeToAd(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	status := server.db.SubscribeToAd(adId, userId)
	if status == database.OK {
		notification, err := server.db.FormRespondNotification(userId, adId)
		if err == nil {
			err = server.db.InsertNotification(notification.WhomId, notification)
			if err != nil {
				log.Println(err)
			}
		}
	}
	DealRequestFromDB(w, "OK", status)
}

func (server *Server) GetAdSubscribers(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}

	params := r.URL.Query()
	pageArr, ok := params["page"]
	if !ok {
		WriteToResponse(w, http.StatusBadRequest,
			fmt.Errorf("page and rows per page have to be in get params and positive int"))
		return
	}
	page, err := strconv.Atoi(pageArr[0])
	if err != nil || page < 1 {
		WriteToResponse(w, http.StatusBadRequest,
			fmt.Errorf("page and rows per page have to be in get params and positive int"))
		return
	}
	rowsPerPageArr, ok := params["rows_per_page"]
	if !ok || len(rowsPerPageArr) != 1 {
		WriteToResponse(w, http.StatusBadRequest,
			fmt.Errorf("page and rows per page have to be in get params and positive int"))
		return
	}
	rowsPerPage, err := strconv.Atoi(rowsPerPageArr[0])
	if err != nil || rowsPerPage < 1 {
		WriteToResponse(w, http.StatusBadRequest,
			fmt.Errorf("page and rows per page have to be in get params and positive int"))
		return
	}
	users, status := server.db.GetAdSubscribers(adId, page, rowsPerPage)
	DealRequestFromDB(w, users, status)
}

func (server *Server) UnsubscribeFromAd(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	status := server.db.UnsubscribeFromAd(adId, userId)
	DealRequestFromDB(w, "OK", status)
}

func (server *Server) MakeDeal(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("ad_id should be int"))
		return
	}
	initiatorId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	params := r.URL.Query()
	subscriberArr, ok := params["subscriber_id"]
	if !ok || len(subscriberArr) != 1 {
		WriteToResponse(w, http.StatusBadRequest,
			fmt.Errorf("subscriber_id has to be in get params and positive int"))
		return
	}
	subscriberId, err := strconv.Atoi(subscriberArr[0])
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("subscriber_id should be int"))
		return
	}
	status, dealId := server.db.MakeDeal(adId, subscriberId, initiatorId)
	if status == database.CREATED {	// TODO: probably go func
		notification, err := server.db.FormAdClosedNotification(dealId, initiatorId, subscriberId)
		if err == nil {
			err = server.db.InsertNotification(subscriberId, notification)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}
		notifications, err := server.db.FormStatusChangedNotificationsByDeal(dealId)
		if err == nil {
			err = server.db.InsertNotifications(notifications)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}

	}

	DealRequestFromDB(w, "OK", status)
}

func (server *Server) FulfillDeal(w http.ResponseWriter, r *http.Request) {
	dealStr := chi.URLParam(r, "deal_id")
	dealId, err := strconv.Atoi(dealStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("deal_id should be int"))
		return
	}
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	notifications, err := server.db.FormStatusChangedNotificationsByDeal(dealId)
	status := server.db.FulfillDeal(dealId, userId)
	if status == database.OK {	// TODO: mb go func
		notification, err := server.db.FormFulfillDealNotification(dealId)
		if err == nil {
			err = server.db.InsertNotification(notification.WhomId, notification)
			if err != nil {
				log.Println(err)
			}
		}

		if err == nil {
			err = server.db.InsertNotifications(notifications)
		} else {
			log.Println(err)
		}
	}
	DealRequestFromDB(w, "OK", status)
}

func (server *Server) CancelDeal(w http.ResponseWriter, r *http.Request) {
	dealStr := chi.URLParam(r, "deal_id")
	dealId, err := strconv.Atoi(dealStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("deal_id should be int"))
		return
	}
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	notifications, err := server.db.FormStatusChangedNotificationsByDeal(dealId)
	status, cancelInfo := server.db.CancelDeal(dealId, userId)
	if status == database.OK {	// TODO: mb go func
		if err == nil {
			err = server.db.InsertNotifications(notifications)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}
		note, err := server.db.FormCancelNotification(cancelInfo.CancelType, userId, cancelInfo.AdId)
		if err == nil {
			err = server.db.InsertNotification(cancelInfo.WhomId, note)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}
	}
	DealRequestFromDB(w, "OK", status)
}

func (server *Server) GetDealForAd(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("ad_id should be int"))
		return
	}
	deal, status := server.db.GetDealForAd(adId)
	DealRequestFromDB(w, deal, status)
}