package fetch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/nkmr-jp/zl"
	"go.uber.org/zap"
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

func Run(ctx context.Context, event event.Event) error {
	objectName := "test"
	url := parse(event)
	bucket := os.Getenv("BUCKET_NAME")

	// TODO: ここでAPIからデータ取る 220507 土 07:39:39
	b := []byte("Hello world.")
	buf := bytes.NewBuffer(b)
	save(ctx, bucket, objectName, buf)

	zl.Info("RUN_GCF_FETCH",
		zap.String("url", url),
		zap.String("bucket", bucket),
	)
	return nil
}

func parse(event event.Event) string {
	var msg MessagePublishedData
	if err := event.DataAs(&msg); err != nil {
		zl.Error("DATA_AS_ERROR", err)
	}
	url := msg.Message.Attributes["url"]
	if url == "" {
		zl.Error("ATTRIBUTE_ERROR", fmt.Errorf("url is empty"))
	}
	return url
}

// See: https://cloud.google.com/storage/docs/streaming#code-samples
func save(ctx context.Context, bucket, object string, buf *bytes.Buffer) {
	// create client
	client, err := storage.NewClient(ctx)
	if err != nil {
		zl.Error("NEW_CLIENT_ERROR", err)
	}
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			zl.Error("CLIENT_CLOSE_ERROR", err)
		}
	}(client)

	// timeout setting
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// write buffer
	writer := client.Bucket(bucket).Object(object).NewWriter(ctx)
	writer.ChunkSize = 0
	if _, err := io.Copy(writer, buf); err != nil {
		zl.Error("IO_COPY_ERROR", err)
	}

	// writer close
	if err := writer.Close(); err != nil {
		zl.Error("WRITER_CLOSE_ERROR", err)
	}

	zl.Debug("SAVED",
		zap.String("bucket", bucket),
		zap.String("object", object),
	)
}

func initLogger() {
	if os.Getenv("FUNCTION_TARGET") != "" {
		zl.SetOutput(zl.ConsoleOutput)
	}
	zl.SetOmitKeys(zl.PIDKey, zl.HostnameKey)
	zl.SetLevel(zl.DebugLevel)
	zl.Init()
}
