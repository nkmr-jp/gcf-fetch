//
package fetch_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/googleapis/google-cloudevents-go/cloud/pubsub/v1"
	fetch "github.com/nkmr-jp/gcf-fetch"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	test := NewTestFetch(t)

	t.Run("single url", func(t *testing.T) {
		objPath := "api.github.com/users/github"
		pubsubData := "https://" + objPath
		ctx := context.Background()

		// Send pubsub message1
		if err := fetch.Run(ctx, test.event(pubsubData)); err != nil {
			assert.Fail(t, err.Error())
		}
		reader1 := test.getObject(ctx, objPath)

		// Send pubsub message2
		if err := fetch.Run(ctx, test.event(pubsubData)); err != nil {
			assert.Fail(t, err.Error())
		}
		reader2 := test.getObject(ctx, objPath)

		// Get generation after send pubsub message.
		assert.NotEqual(t, reader1.Attrs.Generation, reader2.Attrs.Generation)
	})

	t.Run("multi url", func(t *testing.T) {
		t.Skip()
		// if got := add(tt.args.a, tt.args.b); got != tt.want {
		// 	t.Errorf("add() = %v, want %v", got, tt.want)
		// }
	})
}

type TestFetch struct {
	t *testing.T
}

func NewTestFetch(t *testing.T) *TestFetch {
	return &TestFetch{t}
}

func (f *TestFetch) getObject(ctx context.Context, objPath string) *storage.Reader {
	client, _ := storage.NewClient(ctx)
	reader, err := client.Bucket(os.Getenv("BUCKET_NAME")).Object(objPath).NewReader(ctx)
	defer func(rc *storage.Reader) {
		err := rc.Close()
		if err != nil {
			assert.Fail(f.t, err.Error())
		}
	}(reader)
	if err != nil {
		assert.Fail(f.t, err.Error())
	}
	return reader
}

func (f *TestFetch) event(data string) event.Event {
	msg := pubsub.MessagePublishedData{
		Message: &pubsub.Message{
			Data: []byte(data),
		},
	}
	e := event.New()
	e.SetDataContentType("application/json")
	if err := e.SetData(e.DataContentType(), msg); err != nil {
		assert.Fail(f.t, err.Error())
	}
	return e
}
