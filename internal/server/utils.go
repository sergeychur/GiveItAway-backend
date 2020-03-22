package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/models"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func CheckUserAuth(info models.AuthInfo, secret string) (int, bool) {
	// This verifies by url of app
	// TODO(sergeychur): check if works
	const vkSign = "vk_"
	const signKey = "sign"
	const userIdKey = "vk_user_id"
	urlObject, err := url.Parse(info.Url) // parsed url
	if err != nil {
		return 0, false
	}
	queryMap, err := url.ParseQuery(urlObject.RawQuery) // we get query
	digestArr, ok := queryMap[signKey]                  // we get the digest to compare
	if !ok || len(digestArr) == 0 {
		return 0, false
	}
	digest := digestArr[0]

	userIdArr, ok := queryMap[userIdKey]
	if !ok || len(userIdArr) == 0 {
		return 0, false
	}
	userId, err := strconv.Atoi(userIdArr[0])
	// userId must be int
	if err != nil {
		return 0, false
	}
	// we find only fields starting with "vk_"
	for key := range queryMap {
		if key[0:3] != vkSign {
			queryMap.Del(key)
		}
	}
	queryString := queryMap.Encode()                     // we build new quert string
	h := hmac.New(sha256.New, []byte(secret))            // HMAC init
	h.Write([]byte(queryString))                         // HMAC counting
	str := base64.StdEncoding.EncodeToString(h.Sum(nil)) // base64 encoding

	// because of RFC 4648 which says these symbols are changed during encoding
	str = strings.ReplaceAll(str, "+", "-")
	str = strings.ReplaceAll(str, "/", "_")
	// maybe we should just remove all '='. But seems it is only once at the end
	return userId, str[0:len(str)-1] == digest
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
}
