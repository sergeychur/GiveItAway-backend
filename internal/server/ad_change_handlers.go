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

func (serv *Server) DeleteAd(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	userId := 0
	// TODO: take it from cookies later
	status := serv.db.DeleteAd(adId, userId)
	DealRequestFromDB(w, "OK", status)

	if status != database.FORBIDDEN {
		err = filesystem.DeleteAdPhotos(serv.config.UploadPath, adId)
		if err != nil {
			log.Printf("Didn't delete photos for ad %d\n", adId)
		}
	}
}

func (serv *Server) DeleteAdPhoto(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	userId := 0
	// TODO: take it from cookies later
	params := r.URL.Query()
	photoIds, ok := params["ad_photo_id"]
	if !ok || len(photoIds) < 1 {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("photo id has to be int "))
		return
	}
	status, photoUrls := serv.db.DeletePhotosFromAd(adId, userId, photoIds)
	DealRequestFromDB(w, "OK", status)
	if status != database.FORBIDDEN {
		for _, photoUrl := range photoUrls {
			err = filesystem.DeleteAdPhoto(serv.config.UploadPath, photoUrl)
			if err != nil {
				log.Printf("Didn't delete photo: %s\nbecause of %v", photoUrl, err)
			}
		}
	}
}

func (serv *Server) EditAd(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	userId := 0
	// TODO: take it from cookies later
	ad := models.Ad{}
	err = ReadFromBody(r, w, &ad)
	if err != nil {
		return
	}
	if ad.FeedbackType != database.Comments && ad.FeedbackType != database.LS && ad.FeedbackType != database.Other {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("wrong feedback type"))
		return
	}
	status := serv.db.EditAd(adId, userId, ad)
	DealRequestFromDB(w, "OK", status)
}
