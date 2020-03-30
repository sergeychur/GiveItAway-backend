package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/models"
	"net/http"
	"strconv"
)

func (serv *Server) FindAds(w http.ResponseWriter, r *http.Request) {
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
	queryArr, ok := params["query"]
	status := 0
	ads := make([]models.AdForUsers, 0)
	if ok && len(queryArr) == 1 {
		query := queryArr[0]
		ads, status = serv.db.FindAds(query, page, rowsPerPage, params)
	} else {
		ads, status = serv.db.GetAds(page, rowsPerPage, params)
	}
	DealRequestFromDB(w, ads, status)
}

func (serv *Server) GetAdInfo(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	ad, status := serv.db.GetAd(adId)
	DealRequestFromDB(w, &ad, status)
}
