package device

import (
	"encoding/json"
	"log"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

//TODO: Add a struct which the config from google IOT could follow, would be good if it using a standard which the server could use, as contract.

// Configuration struct which is used for setting config push from the back-end
type Configuration struct {
	Config string `json:"Config"`
}

// UpdateConfig parse string and update configuration
func (config *Configuration) UpdateConfig(msg MQTT.Message) {
	log.Println("Updating configuration")
	err := json.Unmarshal(msg.Payload(), config)
	if err != nil {
		log.Println(err)
	}
}
