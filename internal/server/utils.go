package server

import (
	"encoding/json"
	"fmt"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/models"
	"io/ioutil"
	"log"
	"net/http"
)

func WriteToResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	response, _ := json.Marshal(v)
	_, err := w.Write(response)
	if err != nil {
		log.Println("heh, unable to write to response, starve")
	}
}

func ReadFromBody(r *http.Request, w http.ResponseWriter, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errText := models.Error{Message: "Cannot read body"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return fmt.Errorf(errText.Message)
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		errText := models.Error{Message: "Cannot unmarshal json"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return fmt.Errorf(errText.Message)
	}
	return nil
}

func DealRequestFromDB(w http.ResponseWriter, v interface{}, status int) {
	if status == database.DB_ERROR {
		errText := models.Error{Message: "Error in DB"}
		WriteToResponse(w, http.StatusInternalServerError, errText)
		return
	}

	if status == database.FOUND {
		WriteToResponse(w, http.StatusOK, v)
		return
	}

	if status == database.CREATED {
		WriteToResponse(w, http.StatusCreated, v)
		return
	}

	if status == database.EMPTY_RESULT {
		errText := models.Error{Message: "No such item"}
		WriteToResponse(w, http.StatusNotFound, errText)
		return
	}

	if status == database.FORBIDDEN {
		errText := models.Error{Message: "This action is forbidden"}
		WriteToResponse(w, http.StatusForbidden, errText)
		return
	}

	if status == database.CONFLICT {
		WriteToResponse(w, http.StatusConflict,
			fmt.Errorf("conflict happened while performing these actions"))
		return
	}

	if status == database.WRONG_INPUT {
		WriteToResponse(w, http.StatusBadRequest,
			fmt.Errorf("input is incorrect"))
		return
	}

	if status == database.TOO_MUCH_TIMES {
		WriteToResponse(w, http.StatusTooManyRequests,
			fmt.Errorf("flood"))
		return
	}
}

