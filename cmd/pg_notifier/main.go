package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.TODO()
	//pgx v5 connection
	pgpool, err := pgxpool.New(ctx, "postgres://boya:@localhost:28813/law_docs")

	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}
	defer pgpool.Close()
	conn, err := pgpool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Unable to acquire connection: %v\n", err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), "LISTEN test_topic")
	if err != nil {
		log.Fatalf("Unable to listen to task_queue: %v\n", err)
	}

	for {
		notification, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			log.Printf("Error waiting for notification: %v\n", err)
			continue
		}

		log.Printf("Received notification: %s\n", notification.Payload)

	}

}
