package middlewares

import (
	"github.com/go-chi/cors" // here was rs instead of go-chi
	"log"
	"net/http"
)

func CreateCorsMiddleware(allowedHosts []string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		c := cors.New(cors.Options{
			AllowedHeaders:     []string{"Access-Control-Allow-Origin", "Charset", "Content-Type", "Access-Control-Allow-Credentials"},
			AllowedOrigins:     allowedHosts,
			AllowCredentials:   true,
			AllowedMethods:     []string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE", "PATCH"},
			OptionsPassthrough: true,
			Debug:              false,
		})
		return c.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			handler.ServeHTTP(w, r)
		}))
	}
}

func CreateCheckAuthMiddleware(secret []byte, cookieField string,
	checkFunc func(request *http.Request, secret []byte, cookieField string) bool) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if checkFunc(r, secret, cookieField) {
				handler.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			log.Println("Check auth middleware fail")
		})
	}
}

