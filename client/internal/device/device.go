package device

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//TODO: Subscribe to commands topic and add to buffered channel

// Device interface is mainly used for ensuring testability mock's of device.
type Device interface {
	// GetFullDeviceName, concatenate deviceID, regionID, projectID and registry
	// and creates the google path for the device.
	GetFullDeviceName() string
	// LoadPrivateKey load a privatekey and return it
	//LoadPrivateKey(*string) *rsa.PrivateKey
	//	// InitDevice init a new device
	//	InitDevice()
	// Disconnect the device
	SendEvent(*Event, string) error
	DeviceID() *string
	// Buffer for event recived from "/devices/{deviceID}/config" topic
	ConfigQ() *chan string
	// Buffer for event recived from "/devices/{deviceID}/command" topic
	CmdQ() *chan string
	SubscribeToConfigTopic()
	Disconnect()
	Connect() error
}

type device struct {
	deviceID   string
	regionID   string
	projectID  string
	registry   string
	privateKey *rsa.PrivateKey
	client     MQTT.Client
	configQ    chan string
	cmdQ       chan string
	certsCA    string
}

const (
	host     = "mqtt.googleapis.com"
	port     = "443"
	idPrefix = "eid"
	certsCA  = "root-ca.pem"
)

// NewDevice constructor for device
func NewDevice(deviceID string, regionID string, projectID string, registry string, certsCA string, key string) Device {
	log.Printf("Creating Device %s:", deviceID)
	dev := &device{deviceID: deviceID,
		regionID:  regionID,
		projectID: projectID,
		registry:  registry,
		certsCA:   certsCA,
	}

	dev.configQ = make(chan string, 10)
	dev.cmdQ = make(chan string, 10)

	dev.privateKey = dev.loadPrivateKey(&key)

	client := dev.setupMqttClient()
	dev.client = client
	return dev

}

func (dev *device) DeviceID() *string {
	return &dev.deviceID
}

func (dev *device) ConfigQ() *chan string {
	return &dev.configQ
}

func (dev *device) CmdQ() *chan string {
	return &dev.cmdQ
}

func (dev *device) GetFullDeviceName() string {
	return fmt.Sprintf("projects/%v/locations/%v/registries/%v/devices/%v",
		dev.projectID,
		dev.regionID,
		dev.registry,
		dev.deviceID)
}

func (dev *device) Disconnect() {
	if dev.client.IsConnected() == true {
		msg := fmt.Sprintf("Disconnect device: %s", dev.deviceID)
		log.Println(msg)
	}

}

func (dev *device) Connect() error {
	log.Printf("Connecting device: %s", dev.deviceID)
	if token := dev.client.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Device:", *dev.DeviceID(), "could not connect, exited with error:", token.Error())
		return token.Error()
	}
	return nil
}

func (dev *device) loadPrivateKey(privateKey *string) *rsa.PrivateKey {
	log.Println("Loading private key")
	keyBytes, err := ioutil.ReadFile(*privateKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Parsing private key")
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		log.Fatal(err)
	}
	return key
}

// setupMqttClient creates and configures a mqtt client
func (dev *device) setupMqttClient() MQTT.Client {
	log.Println("Setup a MQTT client")
	// create MQTT options
	opts := dev.createMQTTOpts()
	client := MQTT.NewClient(opts)
	return client
}

func (dev *device) createMQTTOpts() *MQTT.ClientOptions {
	opts := MQTT.NewClientOptions()

	log.Println("Setup TLS config")
	config := dev.createTLSConfig()

	broker := fmt.Sprintf("ssl://%v:%v", host, port)
	log.Printf("Broker '%v'", broker)

	opts.AddBroker(broker)
	opts.SetClientID(dev.GetFullDeviceName()).SetTLSConfig(config)
	opts.SetUsername("unused")
	opts.SetProtocolVersion(4)

	log.Println("Setting Password")
	token := dev.getToken()
	opts.SetPassword(dev.signToken(token))

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("[handler] Topic: %v\n", msg.Topic())
		fmt.Printf("[handler] Payload: %v\n", msg.Payload())
	})

	return opts
}

func (dev *device) getToken() *jwt.Token {
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = jwt.StandardClaims{
		Audience:  dev.projectID,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	return token
}

func (dev *device) signToken(token *jwt.Token) string {
	log.Println("Signing token")
	tokenString, err := token.SignedString(dev.privateKey)
	if err != nil {
		log.Fatal(err)
	}
	return tokenString
}

func (dev *device) createTLSConfig() *tls.Config {
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(dev.certsCA)
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	config := &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{},
		MinVersion:         tls.VersionTLS12,
	}
	return config
}

func (dev *device) SendEvent(event *Event, topic string) error {
	data := event.ToJSONString()
	token := dev.client.Publish(
		topic,
		0,
		false,
		data,
	)
	token.WaitTimeout(5 * time.Second)
	if token.Error() != nil {
		fmt.Printf("Error publishing: %s", token.Error())
		return token.Error()
	}
	log.Println("sent event", event)
	return nil
}

// subscribeToTopic rakes the topic and a messageHandler function.
// If nil is provided instead of a messageHandler the default vill be used.
func (dev *device) subscribeToTopic(topic string, messageHandler MQTT.MessageHandler) {
	if messageHandler == nil {
		messageHandler = onIncomingDataReceivedDefault
	}

	for {
		// subscribe the topic, "#" wildcard topic
		token := dev.client.Subscribe(topic, 0, messageHandler)
		if token.Wait() && token.Error() != nil {
			log.Println("Fail to sub... ", token.Error())
			time.Sleep(5 * time.Second)

			log.Printf("Retry to subscribe to topic: %s", topic)
			continue
		} else {
			log.Printf("Subscribe successfult to topic: %s", topic)
			break
		}
	}
}

func (dev *device) SubscribeToConfigTopic() {
	topic := fmt.Sprintf("/devices/%s/config", *dev.DeviceID())
	dev.subscribeToTopic(topic, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf(msg.Topic(), " ", string(msg.Payload()))
		dev.configQ <- string(msg.Payload())
	})
}

func (dev *device) subscribeToCmdTopic() {
	topic := fmt.Sprintf("/devices/%s/commands/#", *dev.DeviceID())
	dev.subscribeToTopic(topic, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf(msg.Topic(), " ", string(msg.Payload()))
		dev.cmdQ <- string(msg.Payload())
	})
}

func onIncomingDataReceivedDefault(client mqtt.Client, message mqtt.Message) {
	log.Printf(message.Topic(), " ", string(message.Payload()))
}
