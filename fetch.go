package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

// Entry defines a log entry.
type Entry struct {
	Message  string `json:"message"`
	Severity string `json:"severity,omitempty"`
	Trace    string `json:"logging.googleapis.com/trace,omitempty"`

	// Logs Explorer allows filtering and display of this as `jsonPayload.component`.
	Component string `json:"component,omitempty"`
}

func (e Entry) String() string {
	if e.Severity == "" {
		e.Severity = "INFO"
	}
	out, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	return string(out)
}

func init() {
	functions.CloudEvent("HelloPubSub", helloPubSub)
}

// MessagePublishedData contains the full Pub/Sub message
// See the documentation for more details:
// https://cloud.google.com/eventarc/docs/cloudevents#pubsub
type MessagePublishedData struct {
	Message pubsub.Message
}

func helloPubSub(ctx context.Context, e event.Event) error {
	loggingTest()
	var msg MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %v", err)
	}

	name := string(msg.Message.Data) // Automatically decoded from base64.
	if name == "" {
		name = "World"
	}

	log.Printf("Hello, %s!", name)
	return nil
}

func loggingTest() {
	log.Println("This is stderr")
	fmt.Println("This is stdout")

	// Structured logging can be used to set severity levels.
	// See https://cloud.google.com/logging/docs/structured-logging.
	fmt.Println(`{"message": "This has ERROR severity", "severity": "error"}`)

	log.Println(Entry{
		Severity:  "NOTICE",
		Message:   "This is the default display field.",
		Component: "arbitrary-property",
	})
}
