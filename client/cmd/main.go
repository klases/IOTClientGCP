package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

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
	defer d.Disconnect()
	if err != nil {
		log.Fatal(err)
	}
	err = d.SubscribeToConfigTopic()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	err = sendAirData(&wg, fmt.Sprintf("/devices/%s/events", *d.DeviceID()), d)
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}

// Send AirDatra from file airData
func sendAirData(wg *sync.WaitGroup, topic string, d device.Device) error {
	f, err := os.Open("airData")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalln(err)
	}
	header := lines[0]
	lines = lines[1:]
	fmt.Println(header)
	for _, line := range lines {
		m := createLineToMap(line, header)
		event := device.NewEvent(eventSrc)
		event.Data = m
		d.SendEvent(event, topic)
		//time.Sleep(1 * time.Second)
	}
	wg.Done()
	return err
}

// Line of the cvs to a map
func createLineToMap(line []string, header []string) map[string]string {
	m := make(map[string]string)
	for _, head := range header {
		for _, value := range line {

			m[head] = value
		}

	}
	return m
}
