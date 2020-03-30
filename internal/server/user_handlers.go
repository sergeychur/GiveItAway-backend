package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/models"
	"net/http"
	"strconv"
)

func (serv *Server) AuthUser(w http.ResponseWriter, r *http.Request) {
	info := models.AuthInfo{}
	err := ReadFromBody(r, w, &info)
	if err != nil {
		return
	}
	userId, isUserCorrect := CheckUserAuth(info, serv.config.Secret)
	if !isUserCorrect {
		WriteToResponse(w, http.StatusUnauthorized, fmt.Errorf("auth data is invalid"))
		return
	}
	user, status := serv.db.GetUser(userId)
	if status == database.EMPTY_RESULT {
		status = serv.db.CreateUser(userId, info.Name, info.Surname, info.PhotoURL)
		if status == database.CREATED {
			newStatus := 0
			user, newStatus = serv.db.GetUser(userId)
			if newStatus != database.FOUND {
				// mb not that way, dunno. but after creation there should be a result
				status = database.DB_ERROR
			}
		}
	}
	// TODO(sergeychur): here should be placing some token to cookies
	DealRequestFromDB(w, &user, status)
}

func (serv *Server) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	user := models.User{}
	user, status := serv.db.GetUser(userId)
	DealRequestFromDB(w, &user, status)
}
