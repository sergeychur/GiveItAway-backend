package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/filesystem"
	"github.com/sergeychur/give_it_away/internal/global_constants"
	"github.com/sergeychur/give_it_away/internal/models"
	"github.com/sergeychur/give_it_away/internal/notifications"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"time"
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
	//if ad.FeedbackType != database.Comments && ad.FeedbackType != database.LS && ad.FeedbackType != database.Other {
	//	WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("wrong feedback type"))
	//	return
	//}
	now := time.Now()
	_, ok := server.AntiFloodAdMap[ad.AuthorId]
	if !ok {
		server.AntiFloodAdMap[ad.AuthorId] = make([]time.Time, 1)
		server.AntiFloodAdMap[ad.AuthorId][0] = time.Now()
	} else {
		n := 0
		// filter slice in place
		for _, x := range server.AntiFloodAdMap[ad.AuthorId] {
			if now.Sub(x) <= time.Duration(server.config.MinutesAntiFlood) * time.Minute {
				server.AntiFloodAdMap[ad.AuthorId][n] = x
				n++
			}
		}
		server.AntiFloodAdMap[ad.AuthorId] = server.AntiFloodAdMap[ad.AuthorId][:n]

		// add new request time
		server.AntiFloodAdMap[ad.AuthorId] = append(server.AntiFloodAdMap[ad.AuthorId], now)
		if len(server.AntiFloodAdMap[ad.AuthorId]) > server.config.MaxAdsAntiFlood {
			WriteToResponse(w, http.StatusTooManyRequests, nil)
			return
		}
	}
	err, httpStatus := validateFields(ad)
	if err != nil {
		WriteToResponse(w, httpStatus, nil)
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
		return
	}
	canUpload, stat := server.db.CanUploadPhoto(adId, server.config.MaxPhotosAd)
	if stat != database.OK {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("database error"))
		return
	}
	if !canUpload {
		WriteToResponse(w, http.StatusForbidden, fmt.Errorf("too musch photos"))
		return
	}
	pathToPhoto, err := filesystem.UploadFile(w, r, function,
		server.config.UploadPath, fmt.Sprintf("post_%d", adId))
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
	//if ad.FeedbackType != database.Comments && ad.FeedbackType != database.LS && ad.FeedbackType != database.Other {
	//	WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("wrong feedback type"))
	//	return
	//}
	err, httpStatus := validateFields(ad)
	if err != nil {
		WriteToResponse(w, httpStatus, nil)
		return
	}
	// TODO(EDIT): check this

	status := server.db.EditAd(adId, userId, ad)
	if status == database.OK {
		retVal, getStatus := server.db.GetAd(adId, userId, server.config.MinutesAntiFlood, server.config.MaxViewsAd)
		if getStatus != database.OK {
			log.Println("cannot get ad, strange")
		} else {
			note := FormEditAdUpdate(retVal)
			server.NotificationSender.SendToChannel(r.Context(), note, fmt.Sprintf("ad_%d", adId))
		}
	}
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

func validateFields(ad models.Ad) (error, int) {
	validationMap := map[string]int{
		ad.Header: global_constants.MaxHeaderLen,
		ad.Text: global_constants.MaxTextLen,
		ad.SubCat: global_constants.MaxCategoryLen,
		ad.SubCatList: global_constants.MaxCategoryLen,
		ad.FullAdress: global_constants.MaxFullAdressLen,
		ad.Metro: global_constants.MaxMetroLen,
		ad.AdType: global_constants.MaxAdType,
		ad.Category: global_constants.MaxCategoryLen,
		ad.CreationDate: global_constants.MaxCreationDate,
		ad.District: global_constants.MaxDistrict,
	}
	for field, length := range validationMap {
		if len(field) > length {
			log.Println(field, " is too large for ad")
			return fmt.Errorf("too large"), http.StatusRequestEntityTooLarge
		}
	}
	return nil, http.StatusOK
}