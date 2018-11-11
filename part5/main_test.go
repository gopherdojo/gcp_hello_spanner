package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/spanner"
)

var testClient *spanner.Client

func TestMain(m *testing.M) {
	ctx := context.Background()

	// CreateSessionに時間がかかると最初に実行されたテストに時間がかかってるように見えるので、Sessionのウォームアップをしておく
	testClient = CreateClient(ctx, fmt.Sprintf("projects/gcpug-public-spanner/instances/merpay-sponsored-instance/databases/%s", databaseName))
	if err := testClient.Single().Query(ctx, spanner.NewStatement("SELECT 1")).Do(func(*spanner.Row) error { return nil }); err != nil {
		log.Fatalf("failed %+v", err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestInsertItem(t *testing.T) {
	ctx := context.Background()

	if err := InsertItem(ctx, testClient); err != nil {
		t.Fatal(err)
	}
}
