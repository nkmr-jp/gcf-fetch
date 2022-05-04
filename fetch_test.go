package fetch_test

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
	fetch "github.com/nkmr-jp/gcf-fetch"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	msg := fetch.MessagePublishedData{
		Message: pubsub.Message{
			Data: []byte("test message"),
		},
	}

	e := event.New()
	e.SetDataContentType("application/json")
	if err := e.SetData(e.DataContentType(), msg); err != nil {
		assert.Fail(t, err.Error())
	}
	// Send message
	if err := fetch.Run(context.Background(), e); err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, "Something", "Something")
}
