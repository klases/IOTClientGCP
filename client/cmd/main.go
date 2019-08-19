package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	device "github.com/klases/IOTClientGCP/client/internal/device"
)

var (
	deviceID   = flag.String("device", "", "Cloud IoT Core Device ID")
	projectID  = flag.String("project", "", "GCP Project ID")
	registryID = flag.String("registry", "", "Cloud IoT Registry ID (short form)")
	region     = flag.String("region", "europe-west1", "GCP Region")
	numEvents  = flag.Int("events", 10, "Number of events to sent")
	eventSrc   = flag.String("src", "", "Event source")
	certsCA    = flag.String("ca", "root-ca.pem", "Download https://pki.google.com/roots.pem")
	privateKey = flag.String("key", "device.key.pem", "Path to private key file")
)

func main() {

	flag.Parse()

	d := device.NewDevice(*deviceID, *region, *projectID, *registryID, *certsCA, *privateKey)
	err := d.Connect()
	if err != nil {
		log.Fatal(err)
	}
	err = d.SubscribeToConfigTopic()
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < *numEvents; i++ {
		event := device.NewEvent(eventSrc)
		topic := fmt.Sprintf("/devices/%s/events", *d.DeviceID())
		d.SendEvent(event, topic)
		time.Sleep(30 * time.Second)
	}
	d.Disconnect()
}
