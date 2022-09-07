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
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/googleapis/google-cloudevents-go/cloud/pubsub/v1"
	"github.com/nkmr-jp/zl"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	initLogger()
	functions.CloudEvent("Fetch", Fetch)
}

type Results []string

func (r Results) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, u := range r {
		enc.AppendString(u)
	}
	return nil
}

func Fetch(ctx context.Context, event event.Event) error {
	defer zl.Sync() // Flush log file buffer. for debug in mac local.

	bucket := getEnv()
	urls := parseEvent(event)
	if urls == nil {
		return fmt.Errorf("urls is nil")
	}

	var successes, failures Results
	for i := range urls {
		object := parseURL(urls[i])
		buf := get(urls[i])
		if err := save(ctx, bucket, object, buf); err != nil {
			failures = append(failures, object)
		} else {
			successes = append(successes, object)
		}
		time.Sleep(time.Second)
	}

	console := fmt.Sprintf("%d successes. %d failures.", len(successes), len(failures))
	fields := []zap.Field{
		zap.String("bucket", bucket),
		zap.Array("successes", successes),
		zap.Array("failures", failures),
	}
	if len(failures) == 0 {
		zl.Info("FETCH_COMPLETE", append(fields, zl.Console(console))...)
	} else {
		err := fmt.Errorf(console)
		zl.Err("FETCH_COMPLETE_WITH_ERROR", err, fields...)
		return err
	}

	return nil
}

// nolint:funlen
func get(urlStr string) *bytes.Buffer {
	u, err := nu.Parse(urlStr)
	if err != nil {
		return nil
	}
	res, err := http.Get(u.String())
	defer func() {
		if err := res.Body.Close(); err != nil {
			zl.Err("HTTP_CLOSE_ERROR", err)
		}
	}()
	if err != nil {
		zl.Err("HTTP_GET_ERROR", err)
		return nil
	}
	if res.StatusCode != http.StatusOK {
		zl.Err("HTTP_GET_ERROR", fmt.Errorf("status code is %d", res.StatusCode))
		return nil
	}

	ret, err := io.ReadAll(res.Body)
	if err != nil {
		zl.Err("READ_BODY_ERROR", err)
		return nil
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, ret, "", "  "); err != nil {
		zl.Err("INDENT_ERROR", err)
		return nil
	}

	return &buf
}

func parseURL(s string) string {
	url, err := nu.Parse(s)
	if err != nil {
		zl.Err("URL_PARSE_ERROR", err)
	}
	return url.Host + url.Path
}

func parseEvent(event event.Event) []string {
	var data pubsub.MessagePublishedData

	err := json.Unmarshal(event.Data(), &data)
	if err != nil {
		zl.Err("UNMARSHAL_ERROR", err)
		return nil
	}
	zl.Info("CLOUD_EVENT_RECEIVED",
		zap.String("cloudEventContext", event.Context.String()),
		zap.Any("cloudEventData", data),
	)

	// return string(data.Message.Data)
	dataStr := strings.TrimSpace(string(data.Message.Data))
	dataStr = strings.ReplaceAll(dataStr, "\n", " ")
	return strings.Split(dataStr, " ")
}

func getEnv() (bucket string) {
	bucket = os.Getenv("BUCKET_NAME")
	if bucket == "" {
		zl.Err("GETENV_ERROR", fmt.Errorf("bucket is empty"))
		return ""
	}
	return bucket
}

// See: https://cloud.google.com/storage/docs/streaming#code-samples
func save(ctx context.Context, bucket, object string, buf *bytes.Buffer) error {
	if buf == nil {
		err := fmt.Errorf("bytes.Buffer is nil")
		zl.Err("BUFFER_ERROR", err)
		return err
	}
	// create client
	client, err := storage.NewClient(ctx)
	if err != nil {
		zl.Err("NEW_CLIENT_ERROR", err)
		return err
	}
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			zl.Err("CLIENT_CLOSE_ERROR", err)
		}
	}(client)

	// timeout setting
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	return writeBuf(ctx, client, bucket, object, buf)
}

func writeBuf(ctx context.Context, client *storage.Client, bucket, object string, buf *bytes.Buffer) error {
	// write buffer
	writer := client.Bucket(bucket).Object(object).NewWriter(ctx)
	writer.ChunkSize = 0
	if _, err := io.Copy(writer, buf); err != nil {
		zl.Err("IO_COPY_ERROR", err)
		return err
	}

	// writer close
	fields := []zap.Field{
		zl.Console(object),
		zap.String("bucket", bucket),
		zap.String("object", object),
	}
	if err := writer.Close(); err != nil {
		zl.Err("WRITER_CLOSE_ERROR", err, fields...)
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

	// test
	if strings.HasSuffix(os.Args[0], ".test") {
		zl.SetRotateFileName("./log/test.jsonl")
	}
	// production
	if os.Getenv("FUNCTION_TARGET") != "" {
		zl.SetOutput(zl.ConsoleOutput)
		zl.SetOmitKeys(zl.PIDKey, zl.HostnameKey)
	}
	zl.SetLevel(zl.DebugLevel)
	zl.Init()
}
