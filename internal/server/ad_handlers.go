package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/models"
	"net/http"
	"strconv"
)

func (serv *Server) CreateAd(w http.ResponseWriter, r *http.Request) {
	ad := models.Ad{}
	err := ReadFromBody(r, w, &ad)
	if err != nil {
		return
	}
	if ad.FeedbackType != database.Comments && ad.FeedbackType != database.LS && ad.FeedbackType != database.Other {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("wrong feedback type"))
		return
	}
	status, adId := serv.db.CreateAd(ad)
	DealRequestFromDB(w, &adId, status)
}

func (serv *Server) FindAds(w http.ResponseWriter, r *http.Request) {

}

func (serv *Server) GetAdInfo(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
	}
	ad := models.Ad{}
	ad, status := serv.db.GetAd(adId)
	DealRequestFromDB(w, &ad, status)
}