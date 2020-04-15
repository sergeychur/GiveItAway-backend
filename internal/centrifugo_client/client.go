package centrifugo_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/centrifugal/gocent"
	"github.com/sergeychur/give_it_away/internal/models"
	"log"
)

type CentrifugoClient struct {
	client *gocent.Client
}

func NewClient(host string, port string, apiKey string) *CentrifugoClient {
	cl := new(CentrifugoClient)
	cl.client = gocent.New(gocent.Config{
		Addr: host + ":" + port,
		Key:  apiKey,
	})
	return cl
}

func (cl *CentrifugoClient) PublishTest(ctx context.Context) {
	log.Println("centrifugo test")
	// maybe change for with cancel
	channel := "test"
	err := cl.client.Publish(ctx, channel, []byte(`{"input": "test"}`))
	if err != nil {
		log.Println(err)
	}
}

func (cl *CentrifugoClient) SendOneClient(ctx context.Context, notification models.Notification, whomId int) {
	channel := fmt.Sprintf("user#%d", whomId)
	data, err := json.Marshal(notification)
	if err != nil {
		log.Println(err)
		return
	}
	err = cl.client.Publish(ctx, channel, data)
	if err != nil {
		log.Println(err)
	}
}

func (cl *CentrifugoClient) SendAllFromList(ctx context.Context, notification models.Notification, whomIds []int) {
	channels := make([]string, 0)
	for _, whomId := range whomIds {
		channels = append(channels, fmt.Sprintf("user#%d", whomId))
	}
	data, err := json.Marshal(notification)
	if err != nil {
		log.Println(err)
		return
	}
	err = cl.client.Broadcast(ctx, channels, data)
	if err != nil {
		log.Println(err)
	}
}

func (cl *CentrifugoClient) SendToChannel(ctx context.Context, notification interface{}, channelName string) {
	data, err := json.Marshal(notification)
	if err != nil {
		log.Println(err)
		return
	}
	err = cl.client.Publish(ctx, channelName, data)
	if err != nil {
		log.Println(err)
	}
}

func (cl *CentrifugoClient) SendAllNotifications(ctx context.Context, notifications []models.Notification) {
	channels := make([]string, 0)
	for _, note := range notifications {
		channels = append(channels, fmt.Sprintf("user#%d", note.WhomId))
	}

	data, err := json.Marshal(notifications[0])
	if err != nil {
		log.Println(err)
		return
	}
	err = cl.client.Broadcast(ctx, channels, data)
	if err != nil {
		log.Println(err)
	}
}