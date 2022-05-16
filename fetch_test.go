package fetch_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/cloudevents/sdk-go/v2/event"
	fetch "github.com/nkmr-jp/gcf-fetch"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	// Prepare for test
	test := NewTestFetch()
	test.setup(t)
	objPath := "api.github.com/users/github"
	ctx := context.Background()
	client, _ := storage.NewClient(ctx)

	// Get generation before send pubsub message.
	var preGeneration int64
	rc, err := client.Bucket(os.Getenv("BUCKET_NAME")).Object(objPath).NewReader(ctx)
	defer rc.Close()
	if err == nil {
		preGeneration = rc.Attrs.Generation
	}

	// Send pubsub message
	if err := fetch.Run(ctx, test.event); err != nil {
		assert.Fail(t, err.Error())
	}

	// Get generation after send pubsub message.
	rc2, err := client.Bucket(os.Getenv("BUCKET_NAME")).Object(objPath).NewReader(ctx)
	defer rc.Close()
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.NotEqual(t, preGeneration, rc2.Attrs.Generation)
}

type TestFetch struct {
	event event.Event
}

func NewTestFetch() *TestFetch {
	return &TestFetch{}
}

func (f *TestFetch) setup(t *testing.T) {
	msg := fetch.MessagePublishedData{
		Message: pubsub.Message{
			Data: []byte("https://api.github.com/users/github"),
		},
	}
	f.event = event.New()
	f.event.SetDataContentType("application/json")
	if err := f.event.SetData(f.event.DataContentType(), msg); err != nil {
		assert.Fail(t, err.Error())
	}
}
