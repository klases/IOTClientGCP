package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	host     = "mqtt.googleapis.com"
	port     = "443"
	idPrefix = "eid"
)

var (
	deviceID   = flag.String("device", "", "Cloud IoT Core Device ID")
	projectID  = flag.String("project", "", "GCP Project ID")
	registryID = flag.String("registry", "", "Cloud IoT Registry ID (short form)")
	region     = flag.String("region", "us-central1", "GCP Region")
	numEvents  = flag.Int("events", 10, "Number of events to sent")
	eventSrc   = flag.String("src", "", "Event source")
	certsCA    = flag.String("ca", "root-ca.pem", "Download https://pki.google.com/roots.pem")
	privateKey = flag.String("key", "", "Path to private key file")
)

func main() {
	flag.Parse()

	fmt.Println("Loading Google's roots...")
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(*certsCA)
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

	clientID := fmt.Sprintf("projects/%v/locations/%v/registries/%v/devices/%v",
		*projectID,
		*region,
		*registryID,
		*deviceID,
	)
	fmt.Println(clientID)

	fmt.Println("Creating MQTT client options...")
	opts := MQTT.NewClientOptions()

	broker := fmt.Sprintf("ssl://%v:%v", host, port)
	fmt.Printf("Broker '%v'", broker)

	opts.AddBroker(broker)
	opts.SetClientID(clientID).SetTLSConfig(config)
	opts.SetUsername("unused")
	opts.SetProtocolVersion(4)

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = jwt.StandardClaims{
		Audience:  *projectID,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	fmt.Println("Loading private key...")
	keyBytes, err := ioutil.ReadFile(*privateKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Parsing private key...")
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Signing token")
	tokenString, err := token.SignedString(key)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Setting password...")
	opts.SetPassword(tokenString)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("[handler] Topic: %v\n", msg.Topic())
		fmt.Printf("[handler] Payload: %v\n", msg.Payload())
	})

	fmt.Println("Connecting...")
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	topic := fmt.Sprintf("/devices/%s/events", *deviceID)
	fmt.Println("Publishing messages...")
	for i := 0; i < *numEvents; i++ {
		data := makeEvent()
		fmt.Printf("Publishing to topic '%s':\n   %v \n", topic, data)
		token := client.Publish(
			topic,
			0,
			false,
			data)
		token.WaitTimeout(5 * time.Second)
		if token.Error() != nil {
			fmt.Printf("Error publishing: %s", token.Error())
		}
	}

	fmt.Println("Disconnecting...")
	client.Disconnect(250)

	fmt.Println("Done")
}

func makeEvent() string {

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	event := struct {
		SourceID string  `json:"source_id"`
		EventID  string  `json:"event_id"`
		EventTs  int64   `json:"event_ts"`
		Metric   float32 `json:"metric"`
	}{
		SourceID: *eventSrc,
		//EventID:  fmt.Sprintf("%s-%s", idPrefix, uuid.NewV4().String(),
		EventID: fmt.Sprintf("%s-%s", idPrefix, "57be638d-7d51-4d8c-b653-2439b8cb5651"),
		EventTs: time.Now().UTC().Unix(),
		Metric:  r1.Float32(),
	}

	data, _ := json.Marshal(event)

	return string(data)

}
