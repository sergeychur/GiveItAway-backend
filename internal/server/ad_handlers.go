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
	params := r.URL.Query()
	pageArr, ok := params["page"]
	if !ok {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("page and rows per page have to be in get params and int"))
		return
	}
	page, err := strconv.Atoi(pageArr[0])
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("page and rows per page have to be in get params and int"))
	}
	rowsPerPageArr, ok := params["rows_per_page"]
	if !ok || len(rowsPerPageArr) != 1 {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("page and rows per page have to be in get params and int"))
		return
	}
	rowsPerPage, err := strconv.Atoi(rowsPerPageArr[0])
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("page and rows per page have to be in get params and int"))
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
	}
	//ad := models.AdForUsers{}
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