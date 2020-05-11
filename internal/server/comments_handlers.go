package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/models"
	"log"
	"net/http"
	"strconv"
)

func (server *Server) GetAdComments(w http.ResponseWriter, r *http.Request) {
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

	adIdStr := chi.URLParam(r, "ad_id")
	adId, err := strconv.Atoi(adIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}

	comments, status := server.db.GetComments(adId, page, rowsPerPage)
	DealRequestFromDB(w, comments, status)
}

func (server *Server) CommentAd(w http.ResponseWriter, r *http.Request) {
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

	comment := models.Comment{}
	err = ReadFromBody(r, w, &comment)
	if err != nil {
		return
	}
	retVal, status := server.db.CreateComment(adId, userId, comment)
	DealRequestFromDB(w, retVal, status)
	// TODO: mb go func
	if status == database.CREATED {
		// todo: check notification to author
		note := FormNewCommentUpdate(retVal)
		server.NotificationSender.SendToChannel(r.Context(), note, fmt.Sprintf("ad_%d", adId))
		noteToAuthor, err := server.db.FormNewCommentNotif(retVal, adId)

		if err != nil {
			log.Println(err)
			return
		}
		if noteToAuthor.WhomId != userId {
			err = server.db.InsertNotification(noteToAuthor.WhomId, noteToAuthor)
			if err != nil {
				log.Println(err)
				return
			}
			server.NotificationSender.SendOneClient(r.Context(), noteToAuthor, noteToAuthor.WhomId)
		}
	}
}

func (server *Server) EditComment(w http.ResponseWriter, r *http.Request) {
	commentIdStr := chi.URLParam(r, "comment_id")
	commentId, err := strconv.Atoi(commentIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}

	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}

	comment := models.Comment{}
	err = ReadFromBody(r, w, &comment)
	if err != nil {
		return
	}
	// todo: check
	retVal, status := server.db.EditComment(commentId, userId, comment)
	adId, err := server.db.GetAdIdForComment(int(retVal.CommentId))
	if err != nil {
		log.Println(err)
	}
	if status == database.CREATED {
		// TODO: mb go func
		note := FormEditCommentUpdate(retVal)
		server.NotificationSender.SendToChannel(r.Context(), note, fmt.Sprintf("ad_%d", adId))
	}
	DealRequestFromDB(w, retVal, status)
}

func (server *Server) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentIdStr := chi.URLParam(r, "comment_id")
	commentId, err := strconv.Atoi(commentIdStr)
	if err != nil {
		WriteToResponse(w, http.StatusBadRequest, fmt.Errorf("id should be int"))
		return
	}

	userId, err := server.GetUserIdFromCookie(r)
	if err != nil {
		WriteToResponse(w, http.StatusInternalServerError, fmt.Errorf("server cannot get userId from cookie"))
	}
	adId, err := server.db.GetAdIdForComment(int(commentId))
	status := server.db.DeleteComment(commentId, userId)
	if err != nil {
		log.Println(err)
	}
	// Todo: check
	if status == database.OK {
		// todo: mb go func
		note := FormDeleteCommentUpdate(models.CommentId{
			CommentId: commentId,
		})
		server.NotificationSender.SendToChannel(r.Context(), note, fmt.Sprintf("ad_%d", adId))
	}
	DealRequestFromDB(w, "OK", status)
}
