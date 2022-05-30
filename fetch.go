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

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/googleapis/google-cloudevents-go/cloud/pubsub/v1"
	"github.com/nkmr-jp/zl"
	"go.uber.org/zap"
)

func init() {
	initLogger()
	functions.CloudEvent("Fetch", Run)
}

func Run(ctx context.Context, event event.Event) error {
	defer zl.Sync() // Flush log file buffer. for debug in mac local.

	bucket := getEnv()
	url := parseEvent(event)
	gcsPath := parseURL(url)
	buf := get(url)

	if err := save(ctx, bucket, gcsPath, buf); err != nil {
		return err
	}

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
	return url.Host + url.Path
}

func parseEvent(event event.Event) string {
	var data pubsub.MessagePublishedData

	err := json.Unmarshal(event.Data(), &data)
	if err != nil {
		zl.Error("UNMARSHAL_ERROR", err)
		return ""
	}
	zl.Info("CLOUD_EVENT_RECEIVED",
		zap.String("cloudEventContext", event.Context.String()),
		zap.Any("cloudEventData", data),
	)

	return string(data.Message.Data)
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
	urlFormat := "https://github.com/nkmr-jp/gcf-fetch/blob/%s"
	srcRootDir := "/workspace/serverless_function_source_code"
	version := os.Getenv("VERSION")
	zl.SetVersion(version)
	zl.SetRepositoryCallerEncoder(urlFormat, version, srcRootDir)

	if os.Getenv("FUNCTION_TARGET") != "" {
		zl.SetOutput(zl.ConsoleOutput)
		zl.SetOmitKeys(zl.PIDKey, zl.HostnameKey)
	}
	zl.SetLevel(zl.DebugLevel)
	zl.Init()
}
