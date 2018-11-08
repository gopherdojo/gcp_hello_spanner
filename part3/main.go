package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
)

const databaseName = "{your slack id}"

type Tweet struct {
	ID         string `spanner:"Id"`
	Author     string
	Content    string
	Count      int
	Favos      []string
	Sort       int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CommitedAt time.Time
}

func main() {
	fmt.Println("ignite")

	ctx := context.Background()

	sc := createClient(ctx, fmt.Sprintf("projects/gcpug-public-spanner/instances/merpay-sponsored-instance/databases/%s", databaseName))

	now := time.Now()
	t := Tweet{
		ID : uuid.New().String(),
		Author: "sinmetal",
		Content: "Hello Spanner",
		Favos: []string{},
		CreatedAt: now,
		UpdatedAt: now,
		CommitedAt: spanner.CommitTimestamp,
	}
	if err := insert(ctx, &t, sc); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("insert id = %s\n", t.ID)

	fmt.Println("done")
}

func createClient(ctx context.Context, db string) (*spanner.Client) {
	dataClient, err := spanner.NewClient(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	return dataClient
}

func insert(ctx context.Context, tweet *Tweet, client *spanner.Client) (error) {
	m, err := spanner.InsertStruct("Tweet", tweet)
	if err != nil {
		return err
	}
	ml := []*spanner.Mutation{
		m,
	}
	_, err = client.Apply(ctx, ml)
	return err
}