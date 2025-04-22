package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wilbyang/law-docs/internal/api"
	repository "github.com/wilbyang/law-docs/internal/db"
)

func main() {
	ctx := context.TODO()
	//pgx v5 connection
	pgpool, err := pgxpool.New(ctx, "postgres://boya:@localhost:28813/law_docs")
	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}
	defer pgpool.Close()
	queries := repository.New(pgpool)

	api := api.NewAPI(queries)
	api.Start(":8080")

}
