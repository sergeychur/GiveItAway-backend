package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sergeychur/give_it_away/internal/auth"
	"github.com/sergeychur/give_it_away/internal/models"
)

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

func SetJWTToCookie(secret []byte, userId int, w http.ResponseWriter, minutes int, cookieField string) error {
	expirationTime := time.Now().Add(time.Duration(minutes) * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"expires": expirationTime.Unix(),
		"userId":  userId,
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieField,
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode, // https://web.dev/samesite-cookies-explained/
	})
	return nil
}

func (server *Server) IsLogined(r *http.Request, secret []byte, cookieField string) bool {
	_, err := server.GetUserIdFromCookie(r)
	return err == nil
}

func (server *Server) GetUserIdFromCookie(r *http.Request) (int, error) {
	cookie, err := r.Cookie(server.CookieField)
	if err != nil {
		return 0, err
	}
	ctx := context.Background()
	StrUserId, err := server.AuthClient.GetUserId(ctx,
		&auth.AuthCookie{
			Data:   cookie.Value,
			Secret: server.config.Secret,
		})
	if err != nil {
		log.Println("GetUserIdFromCookie ", err)
		return int(0), err
	}
	return int(StrUserId.Id), nil
}

func GenerateCentrifugoToken(userId int, minutes int, secret []byte) (models.CentInfo, error) {
	expirationTime := time.Now().Add(time.Duration(minutes) * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": expirationTime.Unix(),
		"sub": strconv.Itoa(userId),
	})
	tokenStr, err := token.SignedString(secret)
	return models.CentInfo{
		Token: tokenStr,
	}, err
}
