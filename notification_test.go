package goapollo

import (
	"context"
	"os"
	"testing"
)

func TestNotificationRepo_Watch(t *testing.T) {
	host := os.Getenv("APOLLO_HOST")
	appId := os.Getenv("APOLLO_APP_ID")
	client := newNotificationRepo(host, appId, "default")

	client.AddNamespace("application")

	select {
	case notify := <-client.Watch(context.Background()):
		t.Logf("%s", &notify)
	}
}
