package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

	//pgx v5 connection
	pgpool := pgxpool.New(context.Background(), "postgres://user:password@localhost:5432/dbname")
	if pgpool == nil {
		panic("Unable to connect to database")
	}
	defer pgpool.Close()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
