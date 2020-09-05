package server

import (
	"github.com/go-chi/chi"
	"github.com/go-vk-api/vk"
	"github.com/sergeychur/give_it_away/internal/models"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	GET_PROFILE = "users.get"
)

var (
	TypeGetMap = map[string]func() interface{}{
		GET_PROFILE: func() interface{} {
			return &[]models.Profile{}
		},
	}

	HandlersMap = map[string]func(interface{}) interface{}{
		GET_PROFILE: func(param interface{}) interface{} {
			casted, ok := param.(*[]models.Profile)
			if !ok {
				log.Println("Cannot convert vk answer to Profile")
				return nil
			}
			if len(*casted) != 0 {
				return &((*casted)[0])
			}
			log.Println("VK returned empty slice")
			return nil
		},
	}
)


func (server *Server) ProxyToVK(w http.ResponseWriter, r *http.Request) {
	methodName := chi.URLParam(r, "method_name")
	params := r.URL.Query()
	paramsRemastered := RemasterParams(params) 
	response := GetResponseTypeFromVK(methodName)
	if response == nil {
		WriteToResponse(w, http.StatusBadRequest, "cannot handle this vk request")
		return
	}
	err := server.VKClient.CallMethod(methodName, paramsRemastered, response)
	handled := HandleVKResponse(response, methodName)
	if handled == nil {
		WriteToResponse(w, http.StatusBadRequest, "cannot handle this vk request")
		return
	}
	if err != nil {
		log.Println("VK cannot serve that because of: ", err)
		WriteToResponse(w, http.StatusForbidden, err)
	}
	WriteToResponse(w, http.StatusOK, handled)
}

func GetResponseTypeFromVK(methodName string) interface{} {

	function, ok := TypeGetMap[methodName]
	if !ok {
		log.Println("we cannot handle that vk method")
		return nil
	}
	return function()
}

func HandleVKResponse(response interface{}, methodName string) interface{} {
	handler, ok := HandlersMap[methodName]
	if !ok {
		log.Println("No handler for type: ", methodName)
		return nil
	}
	return handler(response)
}

func RemasterParams(params url.Values) vk.RequestParams {
	retVal := make(vk.RequestParams, 0)
	for key, value := range params {
		if len(value) == 1 {
			retVal[key] = value[0]
		}
		if len(value) > 1 {
			retVal[key] = strings.Join(value, ",")
		}
	}
	retVal["lang"]  = "ru"
	return retVal
}