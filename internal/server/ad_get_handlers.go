package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
	"strconv"
)

func (server *Server) FindAds(w http.ResponseWriter, r *http.Request) {
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	params := r.URL.Query()
	pageArr, ok := params["page"]
	if !ok {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("page and rows per page have to be in get params and int"))
		return
	}
	page, err := strconv.Atoi(pageArr[0])
	if err != nil || page < 1 {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("page and rows per page have to be in get params and int"))
		return
	}
	rowsPerPageArr, ok := params["rows_per_page"]
	if !ok || len(rowsPerPageArr) != 1 {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("page and rows per page have to be in get params and int"))
		return
	}
	rowsPerPage, err := strconv.Atoi(rowsPerPageArr[0])
	if err != nil || rowsPerPage < 1 {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("page and rows per page have to be in get params and int"))
		return
	}
	//queryArr, ok := params["query"]
	//status := 0
	//ads := make([]models.AdForUsers, 0)
	/*if ok && len(queryArr) == 1 {
		query := queryArr[0]
		ads, status = server.db.FindAds(query, page, rowsPerPage, params, userId)
	} else {

	}*/
	ads, status := server.db.GetAds(page, rowsPerPage, params, userId)
	DealRequestFromDB(w, ads, status)
}

func (server *Server) GetAdInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	ad, status := server.db.GetAd(adId, userId, server.config.MinutesAntiFlood, server.config.MaxViewsAd)
	DealRequestFromDB(w, &ad, status)
}
