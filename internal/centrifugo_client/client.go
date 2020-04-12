package centrifugo_client

import (
	"context"
	"github.com/centrifugal/gocent"
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
