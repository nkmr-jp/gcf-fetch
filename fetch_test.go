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
	"github.com/nkmr-jp/zl"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
)

func TestRun(t *testing.T) {
	test := NewTestFetch(t)

	t.Run("single url", func(t *testing.T) {
		objPath := "api.github.com/users/github"
		pubsubData := "https://api.github.com/users/github"
		ctx := context.Background()
		test.deleteObjects(ctx, "api.github.com")

		// Send pubsub message1
		if err := fetch.Fetch(ctx, test.event(pubsubData)); err != nil {
			assert.Fail(t, err.Error())
		}
		reader1 := test.getObject(ctx, objPath)

		// Send pubsub message2
		if err := fetch.Fetch(ctx, test.event(pubsubData)); err != nil {
			assert.Fail(t, err.Error())
		}
		reader2 := test.getObject(ctx, objPath)
		// Get generation after send pubsub message.
		assert.NotEqual(t, reader1.Attrs.Generation, reader2.Attrs.Generation)
	})

	t.Run("multi url", func(t *testing.T) {
		pubsubData := `
https://api.github.com/users/github
https://api.github.com/users/github/followers
`
		ctx := context.Background()
		test.deleteObjects(ctx, "api.github.com")
		if err := fetch.Fetch(ctx, test.event(pubsubData)); err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.NotNilf(t, test.getObject(ctx, "api.github.com/users/github"), "reader1")
		assert.NotNilf(t, test.getObject(ctx, "api.github.com/users/github/followers"), "reader2")
	})
}

type TestFetch struct {
	t *testing.T
}

func NewTestFetch(t *testing.T) *TestFetch {
	zl.SetRotateFileName("./test.jsonl")
	zl.Init()
	return &TestFetch{t}
}

func (f *TestFetch) deleteObjects(ctx context.Context, objPath string) {
	client, _ := storage.NewClient(ctx)
	query := storage.Query{Prefix: objPath, Versions: true}
	bucket := client.Bucket(os.Getenv("BUCKET_NAME"))
	it := bucket.Objects(ctx, &query)
	for {
		objAttrs, err := it.Next()
		if err != nil && err != iterator.Done {
			assert.Fail(f.t, err.Error())
		}
		if err == iterator.Done {
			break
		}
		if err := bucket.Object(objAttrs.Name).Generation(objAttrs.Generation).Delete(ctx); err != nil {
			assert.Fail(f.t, err.Error())
		}
	}
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
