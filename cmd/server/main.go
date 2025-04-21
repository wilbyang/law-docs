package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	repository "github.com/wilbyang/law-docs/internal/db"
	"github.com/wilbyang/law-docs/internal/models"
)

func main() {
	ctx := context.Background()
	//pgx v5 connection
	pgpool, err := pgxpool.New(ctx, "postgres://boya:@localhost:28813/law_docs")
	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}
	defer pgpool.Close()
	queries := repository.New(pgpool)

	inserted, _ := queries.CreateDocument(ctx, repository.CreateDocumentParams{
		Title:     "test",
		Content:   "test",
		DocSize:   1,
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		Meta:      models.Meta{Key: "test", Value: "test"},
	})

	slog.Info("inserted", "id", inserted.ID, "title", inserted.Title, "content", inserted.Content, "doc_size", inserted.DocSize, "created_at", inserted.CreatedAt.Time, "updated_at", inserted.UpdatedAt.Time, "meta", inserted.Meta)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
