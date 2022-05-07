package fetch

import (
	"bytes"
	"context"
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

func Run(ctx context.Context, e event.Event) error {
	var msg MessagePublishedData
	objectName := "obj_test220507" // ファイル名は後で生成する

	if err := e.DataAs(&msg); err != nil {
		zl.Error("DATA_AS_ERROR", err)
	}
	url := msg.Message.Attributes["url"]
	bucket := msg.Message.Attributes["bucket"]

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

// See: https://cloud.google.com/storage/docs/streaming#code-samples
func save(ctx context.Context, bucketName, objectName string, buf *bytes.Buffer) {
	// Client 作成
	client, err := storage.NewClient(ctx)
	if err != nil {
		zl.Error("storage.NewClient", err)
	}
	defer client.Close()

	// タイムアウト設定
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// 書き込み
	writer := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	writer.ChunkSize = 0
	if _, err := io.Copy(writer, buf); err != nil {
		zl.Error("io.Copy", err)
	}

	// 書き込み終了
	if err := writer.Close(); err != nil {
		zl.Error("Writer.Close", err)
	}
	zl.Debug("SAVED")
}

func initLogger() {
	if os.Getenv("FUNCTION_TARGET") != "" {
		zl.SetOutput(zl.ConsoleOutput)
	}
	zl.SetLevel(zl.DebugLevel)
	zl.Init()
}
