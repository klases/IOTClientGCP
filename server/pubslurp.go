package gcpfunc

import (
	"context"
	"log"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

/*
To deploy run:
gcloud functions deploy HelloPubSub --runtime go111 --trigger-topic iot_device_telematics
*/

func HelloPubSub(ctx context.Context, m PubSubMessage) error {
	name := string(m.Data)
	if name == "" {
		name = "World"
	}
	log.Printf("Hello, %s!", name)
	return nil
}
