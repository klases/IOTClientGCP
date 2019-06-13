package gcpfunc

import (
	"context"
	"log"
)

type PubSubMessage struct {
	Data       []byte `json:"data"`
	Attributes *StateChangeMessage
}

type StateChangeMessage struct {
	DeviceID  string
	Latitude  string
	Longitude string
}

func StateListener(ctx context.Context, m PubSubMessage) error {
	data := string(m.Data)
	if data != "" {
		log.Printf("data: %s", data)
	}

	if m.Attributes == nil {
		log.Println("No Attr")
	} else if m.Attributes.Latitude == "" && m.Attributes.Longitude == "" {
		log.Println("No lat or long")
	} else if m.Attributes.Latitude == "" {
		log.Printf("long: %s", m.Attributes.Longitude)
	} else if m.Attributes.Longitude == "" {
		log.Printf("lat: %s", m.Attributes.Latitude)
	} else {
		log.Printf("(%s, %s)", m.Attributes.Latitude, m.Attributes.Longitude)
	}

	return nil
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
