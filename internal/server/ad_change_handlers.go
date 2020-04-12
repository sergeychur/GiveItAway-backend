package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/filesystem"
	"github.com/sergeychur/give_it_away/internal/models"
	"github.com/sergeychur/give_it_away/internal/notifications"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
)

func (server *Server) CreateAd(w http.ResponseWriter, r *http.Request) {
	ad := models.Ad{}
	err := ReadFromBody(r, w, &ad)
	if err != nil {
		return
	}
	ad.AuthorId, err = server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
		return
	}
	if ad.FeedbackType != database.Comments && ad.FeedbackType != database.LS && ad.FeedbackType != database.Other {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("wrong feedback type"))
		return
	}
	status, adId := server.db.CreateAd(ad)
	DealRequestFromDB(w, &adId, status)
}

func (server *Server) AddPhotoToAd(w http.ResponseWriter, r *http.Request) {
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
		server.config.UploadPath, fmt.Sprintf("ad_%d", adId))
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	status := server.db.AddPhotoToAd(server.config.Host+pathToPhoto, adId, userId)
	DealRequestFromDB(w, "OK", status)
}

func (server *Server) DeleteAd(w http.ResponseWriter, r *http.Request) {
	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
		return
	}
	notificationsArr, errNotif := server.db.FormStatusChangedNotificationsByAd(adId,
		true, notifications.AD_DELETED)

	status := server.db.DeleteAd(adId, userId)
	DealRequestFromDB(w, "OK", status)
	{

		if errNotif == nil {
			server.NotificationSender.SendAllNotifications(r.Context(), notificationsArr)
			err = server.db.InsertNotifications(notificationsArr)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}

	}
	if status == database.OK {
		err = filesystem.DeleteAdPhotos(server.config.UploadPath, adId)
		if err != nil {
			log.Printf("Didn't delete photos for ad %d\n", adId)
		}
	}
}

func (server *Server) DeleteAdPhoto(w http.ResponseWriter, r *http.Request) {
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
	params := r.URL.Query()
	photoIds, ok := params["ad_photo_id"]
	if !ok || len(photoIds) < 1 {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("photo id has to be int "))
		return
	}
	status, photoUrls := server.db.DeletePhotosFromAd(adId, userId, photoIds)
	DealRequestFromDB(w, "OK", status)
	if status != database.FORBIDDEN {
		for _, photoUrl := range photoUrls {
			err = filesystem.DeleteAdPhoto(server.config.UploadPath, photoUrl)
			if err != nil {
				log.Printf("Didn't delete photo: %s\nbecause of %v", photoUrl, err)
			}
		}
	}
}

func (server *Server) EditAd(w http.ResponseWriter, r *http.Request) {
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
	ad := models.Ad{}
	err = ReadFromBody(r, w, &ad)
	if err != nil {
		return
	}
	if ad.FeedbackType != database.Comments && ad.FeedbackType != database.LS && ad.FeedbackType != database.Other {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("wrong feedback type"))
		return
	}
	// TODO(EDIT): send notification to all, who are on the page of ad
	status := server.db.EditAd(adId, userId, ad)
	DealRequestFromDB(w, "OK", status)
}

func (server *Server) SetHidden(w http.ResponseWriter, r *http.Request) {
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
	status := server.db.SetAdHidden(adId, userId)
	DealRequestFromDB(w, "OK", status)
}

func (server *Server) SetVisible(w http.ResponseWriter, r *http.Request) {
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
	status := server.db.SetAdVisible(adId, userId)
	DealRequestFromDB(w, "OK", status)
}
