package server

import (
	"fmt"
	"github.com/go-chi/chi"
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
	log.Print(dealId)	// here we can can do some notification to user
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
	status := server.db.FulfillDeal(dealId, userId)
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
	status := server.db.CancelDeal(dealId, userId)
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