package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	repository "github.com/wilbyang/law-docs/internal/db"
	"github.com/wilbyang/law-docs/internal/models"
	"github.com/wilbyang/law-docs/internal/services"
)

func main() {
	ctx := context.TODO()
	//pgx v5 connection
	connPool, err := pgxpool.New(ctx, "postgres://boya:@localhost:28813/law_docs")
	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}
	defer connPool.Close()
	repo := repository.New(connPool)

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           "http://localhost:4566", // LocalStack 的默认端口
			SigningRegion: "us-east-1",             // LocalStack 默认区域
		}, nil
	})

	// 加载配置
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test", // LocalStack 默认的测试凭据
				SecretAccessKey: "test",
			}, nil
		})),
	)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	notifier, err := services.NewNotifier(ctx, cfg)
	if err != nil {
		slog.Error("Failed to create notifier", "error", err)
	}
	gofakeit.Seed(time.Now().UnixNano())

	notifier.ReceiveMessage(func(message string) error {
		slog.Info("Received message", "message", message)
		// parse message
		var notification models.Notification
		err := json.Unmarshal([]byte(message), &notification)
		if err != nil {
			slog.Error("Failed to unmarshal message", "error", err)
			return err
		}
		//sleep random 1-10 seconds
		time.Sleep(time.Duration(gofakeit.IntRange(1, 10)) * time.Second)
		err = processDocument(ctx, repo, connPool, notification)
		if err != nil {
			slog.Error("Failed to process document", "error", err)
			return err
		}

		return nil
	})

}
func processDocument(ctx context.Context, repo *repository.Queries, connPool *pgxpool.Pool, notification models.Notification) error {

	doc, err := repo.GetDocumentById(ctx, notification.DocID)

	doc.Title = gofakeit.Name()
	doc.Content = gofakeit.Paragraph(10, 10, 10, " ")
	doc.Meta = models.Meta{
		Key:   gofakeit.Name(),
		Value: gofakeit.Name(),
	}
	//doc.Status = pgtype.Text{String: "pre-processed", Valid: true}

	repo.UpdateDocument(ctx, repository.UpdateDocumentParams{
		ID:      doc.ID,
		Title:   doc.Title,
		Content: doc.Content,
		Meta:    doc.Meta,
		Status:  doc.Status,
	})

	// sqlc raw query
	query := `
	UPDATE documents SET title = $1, content = $2, meta = $3, status = $4 WHERE id = $5
	`
	_, err = connPool.Query(ctx, query, doc.Title, doc.Content, doc.Meta, pgtype.Text{String: "pre-processed"}, doc.ID)

	if err != nil {
		slog.Error("Failed to get document", "error", err)
		return err
	}
	//slog.Info("Document", "document", doc)
	if err != nil {
		return err
	}
	return nil
}
