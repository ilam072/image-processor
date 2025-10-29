package main

import (
	"context"
	"github.com/ilam072/image-processor/intenal/config"
	storage "github.com/ilam072/image-processor/intenal/image/filestorage/minio"
	"github.com/ilam072/image-processor/intenal/image/kafka/consumer"
	"github.com/ilam072/image-processor/intenal/image/kafka/handler"
	"github.com/ilam072/image-processor/intenal/image/kafka/producer"
	"github.com/ilam072/image-processor/intenal/image/processor"
	imagerepo "github.com/ilam072/image-processor/intenal/image/repo/postgres"
	imagerest "github.com/ilam072/image-processor/intenal/image/rest"
	imageservice "github.com/ilam072/image-processor/intenal/image/service"
	"github.com/ilam072/image-processor/intenal/middlewares"
	taskrepo "github.com/ilam072/image-processor/intenal/task/repo/postgres"
	taskrest "github.com/ilam072/image-processor/intenal/task/rest"
	taskservice "github.com/ilam072/image-processor/intenal/task/service"
	"github.com/ilam072/image-processor/pkg/db"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize logger
	zlog.Init()

	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Initialize config
	cfg := config.MustLoad()

	// Connect to DB
	DB, err := db.OpenDB(cfg.DB)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to DB")
	}

	// Initialize minio client
	minioClient, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.RootUser, cfg.Minio.RootPassword, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to instantiate minio client")
	}

	exists, err := minioClient.BucketExists(ctx, cfg.Minio.BucketName)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to check if bucket exists")
	}

	if !exists {
		if err := minioClient.MakeBucket(ctx, cfg.Minio.BucketName, minio.MakeBucketOptions{}); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to create bucket")
		}
	}

	// Initialize file image storage
	imageStorage := storage.New(minioClient, cfg.Minio.BucketName)

	// Initialize image processor
	opts := processor.Opts{
		Width:  800,
		Height: 800,
	}
	imageProcessor := processor.NewProcessor(opts)

	// Initialize producer and consumer
	producerr := producer.New(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	consumerr := consumer.New(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)

	// Initialize image and task repositories
	imageRepo := imagerepo.New(DB)
	taskRepo := taskrepo.New(DB)

	// Initialize image and task services
	image := imageservice.New(imageRepo, imageStorage)
	task := taskservice.New(imageRepo, taskRepo, producerr)

	// Initialize Kafka handler
	kafkaHandler := handler.New(task, image, imageStorage, imageProcessor, consumerr)

	// Initialize image and task rest api handlers
	taskHandler := taskrest.NewTaskHandler(task)
	imageHandler := imagerest.NewImageHandler(image, imageStorage)

	// Start Kafka handler
	go func() {
		if err := kafkaHandler.Start(ctx); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to start kafka handler")
		}
	}()

	// Initialize Gin engine
	engine := ginext.New("")
	engine.Use(ginext.Logger())
	engine.Use(ginext.Recovery())
	engine.Use(middlewares.CORS())

	// POST /api/upload
	// GET /api/image/:id?processed=
	// GET /api/image/:id/task/resize
	// GET /api/image/:id/task/thumbnail
	// GET /api/image/:id/task/watermark
	apiGroup := engine.Group("/api")
	apiGroup.POST("/upload", imageHandler.UploadImage)
	apiGroup.GET("/image/:id", imageHandler.GetImage) // query ?processed=resize;thumbnail;watermark
	apiGroup.GET("/image/:id/task/resize", taskHandler.EnqueueTask("resize"))
	apiGroup.GET("/image/:id/task/thumbnail", taskHandler.EnqueueTask("thumbnail"))
	apiGroup.GET("/image/:id/task/watermark", taskHandler.EnqueueTask("watermark"))
	apiGroup.DELETE("/image/:id", imageHandler.DeleteImage)

	// Initialize and start http server
	server := &http.Server{
		Addr:    cfg.Server.HTTPPort,
		Handler: engine,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to listen start http server")
		}
	}()

	<-ctx.Done()

	// Graceful shutdown
	withTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(withTimeout); err != nil {
		zlog.Logger.Error().Err(err).Msg("server shutdown failed")
	}

	if err := DB.Master.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close master database")
	}

	if err := consumerr.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close kafka consumer")
	}

	if err := producerr.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close kafka producer")
	}
}
