package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func (server *Server) GetNotifications(w http.ResponseWriter, r *http.Request) {
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

	notifications, status := server.db.GetNotifications(userId, page, rowsPerPage)
	DealRequestFromDB(w, notifications, status)
}

func (server *Server) CountUnreadNotes(w http.ResponseWriter, r *http.Request) {
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	num, status := server.db.GetUnreadNotesCount(userId)
	DealRequestFromDB(w, num, status)
}

func (server *Server) GetCentrifugoToken(w http.ResponseWriter, r *http.Request) {
	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
		return
	}
	token, err := GenerateCentrifugoToken(userId, 60*24, []byte(server.config.Secret))
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, err)
		return
	}
	WriteToResponse(w, http.StatusOK, token)
}

func (server *Server) TestCentrifugo(w http.ResponseWriter, r *http.Request) {
	server.NotificationSender.PublishTest(r.Context())
}
