package api

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/wilbyang/law-docs/docs"

	entity "github.com/wilbyang/law-docs/internal/db"
	"github.com/wilbyang/law-docs/internal/models"
	"github.com/wilbyang/law-docs/internal/services"
)

var (
	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_upload_total",
			Help: "Total number of file uploads",
		},
		[]string{"doctype", "priority"},
	)
)

func init() {
	prometheus.MustRegister(counter)
}

type API struct {
	router   *gin.Engine
	repo     *entity.Queries
	notifier *services.Notifier
	uploader *services.S3Uploader
}

func NewAPI(queries *entity.Queries) *API {
	ctx := context.TODO()
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {

		switch service {
		case s3.ServiceID:
			return aws.Endpoint{
				URL:           "http://s3.localhost.localstack.cloud:4566",
				SigningRegion: "us-east-1",
			}, nil

		case sqs.ServiceID:
			return aws.Endpoint{
				URL:               "http://sqs.eu-west-1.localhost.localstack.cloud:4566",
				SigningRegion:     "us-east-1",
				HostnameImmutable: true,
			}, nil

		default:
			return aws.Endpoint{
				URL:           "http://localhost:4566",
				SigningRegion: "us-east-1",
			}, nil
		}

	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test",
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
	uploader, err := services.NewS3Uploader(ctx, cfg, "test")
	if err != nil {
		slog.Error("Failed to create uploader", "error", err)
	}
	api := &API{
		repo:     queries,
		router:   gin.Default(),
		notifier: notifier,
		uploader: uploader,
	}

	api.setupRoutes()
	return api
}
func (api *API) Start(addr string) error {
	return api.router.Run(addr)
}

func (api *API) setupRoutes() {
	api.router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	v1 := api.router.Group("/api/v1")
	{
		// Device management
		v1.GET("/docs", api.listDocuments)
		v1.POST("/docs", api.addDocument)
		v1.PUT("/docs/:id", api.updateDocument)
		v1.DELETE("/docs/:id", api.deleteDocument)
		v1.POST("/upload", api.uploadFile)
	}
	api.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

}

// @Summary List all law documents
// @Description Get a list of all law documents
// @Tags documents
// @Produce json
// @Success 200 {object} map[string]interface{} "List of documents"
// @Router /api/v1/docs [get]
func (api *API) listDocuments(c *gin.Context) {
	documents, err := api.repo.GetDocuments(c, pgtype.Int4{Int32: 1, Valid: true})
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "Failed to list documents",
		})
		return
	}
	c.JSON(200, gin.H{
		"status":    "ok",
		"documents": documents,
	})

}

// @Summary Add a new law document
// @Description Create a new law document with the provided details
// @Tags documents
// @Accept json
// @Param document body repository.CreateDocumentParams true "Document details"
// @Produce json
// @Success 201 {object} map[string]interface{} "Document created"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 500 {object} map[string]interface{} "Failed to create document"
// @Router /api/v1/docs [post]
func (api *API) addDocument(c *gin.Context) {
	api.repo.CreateDocument(c, entity.CreateDocumentParams{
		Title:     "test",
		Content:   "test",
		DocSize:   1,
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		Meta:      models.Meta{Key: "test", Value: "test"},
	})
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Document created successfully",
	})
}
func (api *API) uploadFile(c *gin.Context) {
	// Source
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Failed to get file",
		})
		return
	}
	filePath, err := api.uploader.UploadFile(fileHeader)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "Failed to upload file",
		})
		return
	}
	counter.With(prometheus.Labels{"doctype": "pdf", "priority": "high"}).Inc()

	newdoc, err := api.repo.CreateDocument(c, entity.CreateDocumentParams{
		Title:     "",
		Content:   "",
		DocSize:   int32(fileHeader.Size),
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		Status:    pgtype.Text{String: "draft", Valid: true},
		AuthorID:  pgtype.Int4{Int32: 1, Valid: true},
		FilePath:  pgtype.Text{String: filePath, Valid: true},
	})
	if err != nil {
		slog.Error("Failed to create document", "error", err, "filePath", filePath)
		return
	}

	notification := models.Notification{
		DocID: newdoc.ID,
	}
	sending, _ := json.Marshal(notification)

	err = api.notifier.SendMessage(context.TODO(), string(sending))
	if err != nil {
		slog.Error("Failed to send message", "error", err)
		return
	}
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "File uploaded successfully",
	})
}

// @Summary Manage metadata of a law document
// @Description Update the metadata of a law document with the provided details
// @Tags documents
// @Accept json
// @Param id path int true "Document ID"
// @Produce json
// @Success 200 {object} map[string]interface{} "Document updated"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 500 {object} map[string]interface{} "Failed to update document"
// @Router /api/v1/docs/:id [put]
func (api *API) updateDocument(c *gin.Context) {
}
func (api *API) deleteDocument(c *gin.Context) {
}
