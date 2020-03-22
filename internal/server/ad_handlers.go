package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/filesystem"
	"github.com/sergeychur/give_it_away/internal/models"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
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

func (serv *Server) AddPhotoToAd(w http.ResponseWriter, r *http.Request) {
	function := func(header multipart.FileHeader) error {
		re := regexp.MustCompile(`image/.*`)
		if !re.MatchString(header.Header.Get("Content-Type")) {
			log.Print(header.Header.Get("Content-Type"))
			return fmt.Errorf("not an image")
		}
		return nil
	}
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
	}
	pathToPhoto, err := filesystem.UploadFile(w, r, function,
		serv.config.UploadPath, fmt.Sprintf("ad_%d", adId))
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userId := 0
	// TODO: take it from cookies later
	status := serv.db.AddPhotoToAd(pathToPhoto, adId, userId)
	DealRequestFromDB(w, "OK", status)
}