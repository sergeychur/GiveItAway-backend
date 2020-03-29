package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
	"strconv"
)

func (serv *Server) SubscribeToAd(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
	}
	userId := 51000329
	// TODO: later take it from cookie
	status := serv.db.SubscribeToAd(adId, userId)
	DealRequestFromDB(w, "OK", status)
}

func (serv *Server) GetAdSubscribers(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
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
	users, status := serv.db.GetAdSubscribers(adId, page, rowsPerPage)
	DealRequestFromDB(w, users, status)
}
