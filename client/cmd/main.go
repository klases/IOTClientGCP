package main

import (
	"fmt"
	"time"

	device "github.com/klases/IOTClientGCP/client/internal/device"
)

// TODO: Add arg handling

func main() {
	//d := &device.Device{DeviceID: "claes_test_device", RegionID: "europe-west1", ProjectID: "atkiotcore", Registry: "iot_device_raspberry_pi_sensors"}
	d := device.NewDevice("claes_test_device", "europe-west1", "atkiotcore", "iot_device_raspberry_pi_sensors")
	keyPath := "device.key.pem"
	key := d.LoadPrivateKey(&keyPath)
	d.SetPrivateKey(key)
	d.InitDevice()
	eventSrc := "test"

	//topicCMD := fmt.Sprintf("/devices/%s/config", d.DeviceID())
	//fmt.Println(topicCMD)
	//d.SubscribeToTopic(topicCMD)
	for i := 0; i < 5; i++ {
		event := device.NewEvent(&eventSrc)
		topic := fmt.Sprintf("/devices/%s/events", *d.DeviceID())
		d.SendEvent(event, topic)
		time.Sleep(30 * time.Second)
	}
	d.Disconnect()
}
