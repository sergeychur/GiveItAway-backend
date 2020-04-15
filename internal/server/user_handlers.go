package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/global_constants"
	"github.com/sergeychur/give_it_away/internal/models"
	"net/http"
	"strconv"
)

func (server *Server) AuthUser(w http.ResponseWriter, r *http.Request) {
	info := models.AuthInfo{}
	err := ReadFromBody(r, w, &info)
	if err != nil {
		return
	}
	userId, isUserCorrect := CheckUserAuth(info, server.config.Secret)
	if !isUserCorrect {
		WriteToResponse(w, http.StatusUnauthorized, fmt.Errorf("auth data is invalid"))
		return
	}
	user, status := server.db.GetUser(userId)
	if status == database.EMPTY_RESULT {
		status = server.db.CreateUser(userId, info.Name, info.Surname, info.PhotoURL, global_constants.InitialCarma)
		if status == database.CREATED {
			newStatus := 0
			user, newStatus = server.db.GetUser(userId)
			if newStatus != database.FOUND {
				// mb not that way, dunno. but after creation there should be a result
				status = database.DB_ERROR
			}
		}
	}
	err = SetJWTToCookie([]byte(server.config.Secret), userId, w, 60, server.CookieField)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("auth failed"))
		return
	}
	DealRequestFromDB(w, &user, status)
}

func (server *Server) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}
	user, status := server.db.GetUserProfile(userId)
	DealRequestFromDB(w, &user, status)
}

func (server *Server) GetGiven(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
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

	ads, status := server.db.GetGiven(userId, page, rowsPerPage)
	DealRequestFromDB(w, ads, status)
}

func (server *Server) GetReceived(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
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

	ads, status := server.db.GetReceived(userId, page, rowsPerPage)
	DealRequestFromDB(w, ads, status)
}

func (server *Server) GetWanted(w http.ResponseWriter, r *http.Request) {
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

	ads, status := server.db.GetWanted(userId, page, rowsPerPage)
	DealRequestFromDB(w, ads, status)
}


