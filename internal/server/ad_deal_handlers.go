package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/global_constants"
	notifications2 "github.com/sergeychur/give_it_away/internal/notifications"
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
	status, maxBidNote := server.db.SubscribeToAd(adId, userId, global_constants.PriceCoeff)
	if status == database.OK {
		server.SubscribeToAdSendUpd(userId, adId, r)
		if maxBidNote != nil {
			server.NewMaxBidUpd(*maxBidNote, r)
		}
	}
	DealRequestFromDB(w, "OK", status)
}

func (server *Server) IncreaseBid(w http.ResponseWriter, r *http.Request) {
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
	note, status := server.db.IncreaseBid(adId, userId)
	if status == database.OK {
		server.NewMaxBidUpd(note, r)
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
	// todo maybe send notification to ad viewers here too
	status := server.db.UnsubscribeFromAd(adId, userId)
	if status == database.OK {
		server.UnsubscribeToAdSendUpd(userId, adId, r)
		server.db.DeleteInvalidNotesDelete(adId)
	}
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
	typeArr, ok := params["type"]
	if !ok || len(typeArr) != 1 {
		WriteToResponse(w, http.StatusBadRequest,
			fmt.Errorf("type has to be in query"))
	}

	status, dealId, subscriberId := server.db.MakeDeal(adId, initiatorId, typeArr[0], params)

	if status == database.CREATED { // TODO: probably go func
		server.MakeDealSendUpd(dealId, initiatorId, subscriberId, adId, r)
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
	if status == database.OK { // TODO: mb go func
		server.FulFillDealSendUpd(dealId, notifications, r)
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

	if cancelInfo.CancelType == notifications2.AUTHOR_CANCELLED {
		// TODO: check if works
		server.db.DeleteInvalidNotesCancelDeal(notifications[0].AdId)
	}

	if status == database.OK { // TODO: mb go func
		server.CancelDealSendUpd(err, cancelInfo, userId, notifications, r)
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

func (server *Server) GetBidForUser(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("ad_id should be int"))
		return
	}
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	maxBid, status := server.db.GetUserBidForAd(adId, userId)
	DealRequestFromDB(w, maxBid, status)
}

func (server *Server) GetMaxBid(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("ad_id should be int"))
		return
	}
	maxBid, status := server.db.GetMaxBidForAd(adId)
	DealRequestFromDB(w, maxBid, status)
}

func (server *Server) GetMaxBidUser(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("ad_id should be int"))
		return
	}
	maxBid, status := server.db.GetMaxBidUserForAd(adId)
	DealRequestFromDB(w, maxBid, status)
}

func (server *Server) GetReturnSize(w http.ResponseWriter, r *http.Request) {
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("ad_id should be int"))
		return
	}
	maxBid, status := server.db.GetReturnBid(adId, userId)
	DealRequestFromDB(w, maxBid, status)
}
