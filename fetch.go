package fetch

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/nkmr-jp/zl"
)

// MessagePublishedData
// See: https://cloud.google.com/eventarc/docs/cloudevents#pubsub
// See: https://googleapis.github.io/google-cloudevents/examples/binary/pubsub/MessagePublishedData-complex.json
type MessagePublishedData struct {
	Message pubsub.Message
}

func init() {
	initLogger()
	functions.CloudEvent("Fetch", Run)
}

func Run(ctx context.Context, e event.Event) error {
	zl.Info("RUN_GCF_FETCH")
	var msg MessagePublishedData

	if err := e.DataAs(&msg); err != nil {
		zl.Error("DATA_AS_ERROR", err)
	}
	name := string(msg.Message.Data)
	if name == "" {
		name = "World"
	}
	fmt.Printf("Hello, %s!", name)

	return nil
}

func initLogger() {
	if os.Getenv("FUNCTION_TARGET") != "" {
		zl.SetOutput(zl.ConsoleOutput)
	}
	zl.SetLevel(zl.DebugLevel)
	zl.Init()
}
