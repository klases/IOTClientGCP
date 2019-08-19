package device

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	uuid "github.com/satori/go.uuid"
)

// Event type used for sending a MQTT msg
type Event struct {
	SourceID string            `json:"source_id"`
	EventID  string            `json:"event_id"`
	EventTs  int64             `json:"event_ts"`
	Data     map[string]string `json:"Metrics"`
}

// NewEvent this is only a template event
func NewEvent(eventSrc *string) *Event {

	// s1 := rand.NewSource(time.Now().UnixNano())
	// r1 := rand.New(s1)
	uuidV4, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}
	event := &Event{
		SourceID: *eventSrc,
		EventID:  fmt.Sprintf("%s-%s", idPrefix, uuidV4.String()),
		EventTs:  time.Now().UTC().Unix(),
	}
	return event
}

// ToJSONString convert Event to JSON String
func (e *Event) ToJSONString() string {
	data, _ := json.Marshal(e)
	return string(data)
}
