package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	nu "net/url"
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
	defer zl.Sync() // Flush log file buffer. for debug in mac local.

	// objectName := "test"
	bucket := getEnv()
	url := parseEvent(event)
	gcsPath := parseURL(url)
	buf := get(url)

	if err := save(ctx, bucket, gcsPath, buf); err != nil {
		return err
	}

	zl.Info("RUN_GCF_FETCH",
		zap.String("url", url),
		zap.String("bucket", bucket),
	)
	return nil
}

func get(url string) *bytes.Buffer {
	res, err := http.Get(url)
	defer func() {
		if err := res.Body.Close(); err != nil {
			zl.Error("HTTP_CLOSE_ERROR", err)
		}
	}()
	if err != nil {
		zl.Error("HTTP_GET_ERROR", err)
		return nil
	}

	ret, err := io.ReadAll(res.Body)
	if err != nil {
		zl.Error("READ_BODY_ERROR", err)
		return nil
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, ret, "", "  "); err != nil {
		zl.Error("INDENT_ERROR", err)
		return nil
	}

	return &buf
}

func parseURL(s string) string {
	url, err := nu.Parse(s)
	if err != nil {
		zl.Error("URL_PARSE_ERROR", err)
	}
	return url.Host + "/" + url.Path
}

func parseEvent(event event.Event) (url string) {
	var msg MessagePublishedData
	if err := event.DataAs(&msg); err != nil {
		zl.Error("DATA_AS_ERROR", err)
		return
	}
	url = msg.Message.Attributes["url"]
	if url == "" {
		zl.Error("ATTRIBUTE_ERROR", fmt.Errorf("url is empty"))
		return
	}
	return url
}

func getEnv() (bucket string) {
	bucket = os.Getenv("BUCKET_NAME")
	if bucket == "" {
		zl.Error("GETENV_ERROR", fmt.Errorf("bucket is empty"))
		return ""
	}
	return bucket
}

// See: https://cloud.google.com/storage/docs/streaming#code-samples
func save(ctx context.Context, bucket, object string, buf *bytes.Buffer) error {
	// create client
	client, err := storage.NewClient(ctx)
	if err != nil {
		zl.Error("NEW_CLIENT_ERROR", err)
		return err
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
		return err
	}

	// writer close
	fields := []zap.Field{
		zap.String("bucket", bucket),
		zap.String("object", object),
	}
	if err := writer.Close(); err != nil {
		zl.Error("WRITER_CLOSE_ERROR", err, fields...)
		return err
	}
	zl.Debug("SAVED", fields...)
	return nil
}

func initLogger() {
	if os.Getenv("FUNCTION_TARGET") != "" {
		zl.SetOutput(zl.ConsoleOutput)
		zl.SetOmitKeys(zl.PIDKey, zl.HostnameKey)
	}
	zl.SetLevel(zl.DebugLevel)
	zl.Init()
}
