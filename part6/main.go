package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
)

const databaseName = "sinmetal-test1"

const (
	accessCounterTableName = "AccessCounter"
	accessCounterIDPrefix  = "AccessCounter-"
	accessCounterShardMax  = 10
)

func main() {
	fmt.Println("ignite")

	ctx := context.Background()

	sc := CreateClient(ctx, fmt.Sprintf("projects/gcpug-public-spanner/instances/merpay-sponsored-instance/databases/%s", databaseName))

	_, err := IncrementAccessCounter(ctx, sc)
	if err != nil {
		panic(err)
	}

	count, err := GetAccessCounter(ctx, sc)
	if err != nil {
		panic(err)
	}
	fmt.Printf("AccessCounter = %d\n", count)

	fmt.Println("done")
}

type AccessCounter struct {
	ID         string `spanner:"Id"`
	Count      int64
	CommitedAt time.Time
}

// IncrementAccessCounter is カウンターをインクリメントする
// 返し値のint64はShardingCounter全体の値ではなく、更新したカウンタの値
func IncrementAccessCounter(ctx context.Context, client *spanner.Client) (int64, error) {
	var count int64
	shard := rand.Intn(accessCounterShardMax)
	id := fmt.Sprintf("%s%d", accessCounterIDPrefix, shard)
	key := spanner.Key{id}
	_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		row, err := txn.ReadRow(ctx, accessCounterTableName, key, []string{"Count"})
		if err != nil {
			if spanner.ErrCode(err) != codes.NotFound {
				return err
			}
			// noop
		} else {
			if err := row.Column(0, &count); err != nil {
				return err
			}
		}

		count++
		a := AccessCounter{
			ID:         id,
			Count:      count,
			CommitedAt: spanner.CommitTimestamp,
		}
		m, err := spanner.InsertOrUpdateStruct(accessCounterTableName, a)
		if err != nil {
			return err
		}
		if err := txn.BufferWrite([]*spanner.Mutation{m}); err != nil {
			return err
		}
		return nil
	})
	return count, err
}

func GetAccessCounter(ctx context.Context, client *spanner.Client) (int64, error) {
	k := spanner.KeyRange{
		Start: spanner.Key{accessCounterIDPrefix},
		End:   spanner.Key{fmt.Sprintf("%s%d", accessCounterIDPrefix, math.MaxInt64)},
	}
	iter := client.ReadOnlyTransaction().Read(ctx, accessCounterTableName, k, []string{"Count"})
	defer iter.Stop()
	var ac int64
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return ac, err
		}

		var count int64
		if err := row.Column(0, &count); err != nil {
			return ac, err
		}
		ac += count
	}
	return ac, nil
}

func CreateClient(ctx context.Context, db string) *spanner.Client {
	dataClient, err := spanner.NewClient(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	return dataClient
}
