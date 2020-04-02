package notifications

import (
	"encoding/json"
	"fmt"
	"github.com/sergeychur/give_it_away/internal/models"
)

const (
	AD_CLOSE = "ad_close"
)

var (
	funcsmap = map[string]func() interface{} {
		AD_CLOSE: func() interface{} {
			return &models.AuthorClosedAd{}
		},
	}
)

func FormPayLoad(payload []byte, notificationType string) (interface{}, error) {
	//funcsmap := make(map[string] func([]byte) interface{})
	neededFunc, ok := funcsmap[notificationType]
	if !ok {
		return nil, fmt.Errorf("unable to detect notification type")
	}
	val := neededFunc()
	err := json.Unmarshal(payload, &val)
	return val, err
}
