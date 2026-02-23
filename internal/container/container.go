package container

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	rediscache "github.com/EduGoGroup/edugo-shared/cache/redis"
	"github.com/EduGoGroup/edugo-shared/logger"
	"github.com/EduGoGroup/edugo-shared/messaging/rabbit"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/client"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/config"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/handler"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/messaging"
	mongorepo "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/persistence/mongodb/repository"
	pgrepo "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/persistence/postgres/repository"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/storage"
)

// Handlers groups all HTTP handlers.
type Handlers struct {
	Health     *handler.HealthHandler
	Material   *handler.MaterialHandler
	Assessment *handler.AssessmentHandler
	Progress   *handler.ProgressHandler
	Screen     *handler.ScreenHandler
	Stats      *handler.StatsHandler
	Summary    *handler.SummaryHandler
}

// Container is the root dependency injection container.
type Container struct {
	Logger     logger.Logger
	AuthClient *client.AuthClient
	Handlers   *Handlers

	// Infrastructure handles stored for cleanup
	db          *gorm.DB
	mongoClient *mongo.Client
	rabbitConn  *rabbit.Connection
	redisClient *goredis.Client
}

// New creates and wires the entire dependency graph.
func New(ctx context.Context, cfg *config.Config, log logger.Logger) (*Container, error) {
	c := &Container{Logger: log}

	// --- PostgreSQL (GORM) ---
	db, err := gorm.Open(postgres.Open(cfg.Database.Postgres.GormDSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.Postgres.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.Postgres.MaxIdleConns)
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	c.db = db
	log.Info("PostgreSQL connected", "host", cfg.Database.Postgres.Host, "database", cfg.Database.Postgres.Database)

	// --- MongoDB ---
	var mongoDB *mongo.Database
	var mongoClient *mongo.Client
	if cfg.Database.MongoDB.URI != "" {
		mongoClient, err = mongo.Connect(options.Client().ApplyURI(cfg.Database.MongoDB.URI))
		if err != nil {
			log.Warn("MongoDB connection failed, continuing without MongoDB", "error", err)
		} else {
			if err := mongoClient.Ping(ctx, nil); err != nil {
				log.Warn("MongoDB ping failed, continuing without MongoDB", "error", err)
				mongoClient = nil
			} else {
				mongoDB = mongoClient.Database(cfg.Database.MongoDB.Database)
				c.mongoClient = mongoClient
				log.Info("MongoDB connected", "database", cfg.Database.MongoDB.Database)
			}
		}
	}

	// --- RabbitMQ ---
	var publisher rabbit.Publisher
	if cfg.Messaging.RabbitMQ.URL != "" {
		rabbitConn, rabErr := rabbit.Connect(cfg.Messaging.RabbitMQ.URL)
		if rabErr != nil {
			log.Warn("RabbitMQ connection failed, continuing without messaging", "error", rabErr)
		} else {
			publisher = messaging.NewPublisher(rabbitConn)
			c.rabbitConn = rabbitConn
			log.Info("RabbitMQ connected")
		}
	}

	// --- S3 ---
	var s3Client *storage.S3Client
	if cfg.Storage.S3.AccessKeyID != "" {
		s3Client, err = storage.NewS3Client(ctx, cfg.Storage.S3)
		if err != nil {
			log.Warn("S3 client creation failed, continuing without storage", "error", err)
		} else {
			log.Info("S3 client initialized", "bucket", cfg.Storage.S3.Bucket)
		}
	}

	// --- Redis Cache ---
	var cacheSvc rediscache.CacheService
	if cfg.Cache.RedisURL != "" {
		redisClient, redisErr := rediscache.ConnectRedis(rediscache.RedisConfig{URL: cfg.Cache.RedisURL})
		if redisErr != nil {
			log.Warn("Redis connection failed, continuing without cache", "error", redisErr)
		} else {
			cacheSvc = rediscache.NewCacheService(redisClient)
			c.redisClient = redisClient
			log.Info("Redis connected")
		}
	}

	// --- IAM Platform Client ---
	var iamClient *client.IAMPlatformClient
	if cfg.Auth.APIIamPlatform.BaseURL != "" {
		iamClient = client.NewIAMPlatformClient(client.IAMPlatformConfig{
			BaseURL: cfg.Auth.APIIamPlatform.BaseURL,
			Timeout: cfg.Auth.APIIamPlatform.Timeout,
		})
		log.Info("IAM Platform client initialized", "baseURL", cfg.Auth.APIIamPlatform.BaseURL)
	}

	// --- Auth Client ---
	authClient := client.NewAuthClient(client.AuthClientConfig{
		JWTSecret:       cfg.Auth.JWT.Secret,
		JWTIssuer:       cfg.Auth.JWT.Issuer,
		BaseURL:         cfg.Auth.APIAdmin.BaseURL,
		Timeout:         cfg.Auth.APIAdmin.Timeout,
		RemoteEnabled:   cfg.Auth.APIAdmin.RemoteEnabled,
		FallbackEnabled: cfg.Auth.APIAdmin.FallbackEnabled,
		CacheTTL:        cfg.Auth.APIAdmin.CacheTTL,
		CacheEnabled:    cfg.Auth.APIAdmin.CacheEnabled,
	})
	c.AuthClient = authClient

	// --- PostgreSQL Repositories ---
	materialRepo := pgrepo.NewMaterialRepository(db)
	assessmentRepo := pgrepo.NewAssessmentRepository(db)
	attemptRepo := pgrepo.NewAttemptRepository(db)
	progressRepo := pgrepo.NewProgressRepository(db)
	screenRepo := pgrepo.NewScreenRepository(db)
	statsRepo := pgrepo.NewStatsRepository(db)

	// --- MongoDB Repositories ---
	var mongoAssessmentRepo *mongorepo.MongoAssessmentRepository
	var mongoSummaryRepo *mongorepo.MongoSummaryRepository
	if mongoDB != nil {
		mongoAssessmentRepo = mongorepo.NewMongoAssessmentRepository(mongoDB)
		mongoSummaryRepo = mongorepo.NewMongoSummaryRepository(mongoDB)
	}

	// --- Application Services ---
	materialSvc := service.NewMaterialService(materialRepo, s3Client, publisher, log, cfg.Messaging.RabbitMQ.Exchange)
	assessmentSvc := service.NewAssessmentService(assessmentRepo, attemptRepo, mongoAssessmentRepo, log)
	progressSvc := service.NewProgressService(progressRepo, log)
	screenSvc := service.NewScreenService(screenRepo, iamClient, cacheSvc, log)
	statsSvc := service.NewStatsService(statsRepo, log)
	summarySvc := service.NewSummaryService(mongoSummaryRepo, log)

	// --- HTTP Handlers ---
	c.Handlers = &Handlers{
		Health:     handler.NewHealthHandler(db, mongoClient),
		Material:   handler.NewMaterialHandler(materialSvc),
		Assessment: handler.NewAssessmentHandler(assessmentSvc),
		Progress:   handler.NewProgressHandler(progressSvc),
		Screen:     handler.NewScreenHandler(screenSvc),
		Stats:      handler.NewStatsHandler(statsSvc),
		Summary:    handler.NewSummaryHandler(summarySvc),
	}

	return c, nil
}

// Close releases all infrastructure resources in reverse order.
func (c *Container) Close() {
	if c.redisClient != nil {
		if err := c.redisClient.Close(); err != nil {
			c.Logger.Error("closing Redis", "error", err)
		}
	}
	if c.rabbitConn != nil {
		if err := c.rabbitConn.Close(); err != nil {
			c.Logger.Error("closing RabbitMQ", "error", err)
		}
	}
	if c.mongoClient != nil {
		if err := c.mongoClient.Disconnect(context.Background()); err != nil {
			c.Logger.Error("closing MongoDB", "error", err)
		}
	}
	if c.db != nil {
		sqlDB, err := c.db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				c.Logger.Error("closing PostgreSQL", "error", err)
			}
		}
	}
}
