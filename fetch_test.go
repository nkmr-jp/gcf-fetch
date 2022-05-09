package fetch_test

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
	fetch "github.com/nkmr-jp/gcf-fetch"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	test := NewTestFetch()
	test.setup(t)

	// Send message
	if err := fetch.Run(context.Background(), test.event); err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, "Something", "Something")
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
			Attributes: map[string]string{
				"url": "https://api.github.com/users/defunkt",
			},
		},
	}
	f.event = event.New()
	f.event.SetDataContentType("application/json")
	if err := f.event.SetData(f.event.DataContentType(), msg); err != nil {
		assert.Fail(t, err.Error())
	}
}
