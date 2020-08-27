package centrifugo_client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/centrifugal/gocent"
	"github.com/go-vk-api/vk"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/models"
	"github.com/sergeychur/give_it_away/internal/notifications"
)

type CentrifugoClient struct {
	client   *gocent.Client
	vkClient *vk.Client
	db       *database.DB
}

func NewClient(host, port, apiKey, vkApiKey string, db *database.DB) *CentrifugoClient {
	cl := new(CentrifugoClient)

	cl.client = gocent.New(gocent.Config{
		Addr: host + ":" + port,
		Key:  apiKey,
	})

	vkClient, err := vk.NewClientWithOptions(
		vk.WithToken(vkApiKey),
	)
	if err != nil {
		log.Print("init error", err)
		return cl
	}
	cl.vkClient = vkClient
	cl.db = db

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
	log.Print("prepare", cl.vkClient, cl.db)
	if cl.vkClient != nil && cl.db != nil {
		log.Print("get is", notification)
		var b, status = cl.db.GetPermissoinToPM(whomId)
		if status == database.FOUND && b {
			var text, err = notifications.NotificationToText(notification)
			if err != nil {
				log.Print("notification payload error is", err)
			} else {
				err = cl.vkClient.CallMethod("messages.send", vk.RequestParams{
					"peer_id":   whomId,
					"message":   text,
					"random_id": 0,
				}, nil)
				if err != nil {
					log.Print("vkClient error is", err)
				}
			}

		} else {
			log.Print("database status", status)
		}

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
