package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/leventa-backend/database/seeds"
	"github.com/leventa-backend/internal/config"
	"github.com/leventa-backend/internal/delivery/http/handler"
	"github.com/leventa-backend/internal/delivery/http/middleware"
	pgRepo "github.com/leventa-backend/internal/repository/postgres"
	"github.com/leventa-backend/internal/routes"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/internal/storage"
	"github.com/leventa-backend/internal/usecase"
	"github.com/leventa-backend/pkg/logger"
	"github.com/leventa-backend/pkg/validator"

	gormPg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	seedFlag := flag.Bool("seed", false, "Run database seed")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.AppEnv, cfg.LogFile)
	log := logger.Get()

	// Connect to PostgreSQL
	db, err := gorm.Open(gormPg.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get database instance")
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Info().Msg("Connected to PostgreSQL")

	// Run seed if flag is set
	if *seedFlag {
		if cfg.AppEnv == "production" {
			log.Fatal().Msg("Refusing to seed demo accounts when APP_ENV=production")
		}
		if err := seeds.Run(db); err != nil {
			log.Fatal().Err(err).Msg("Seed failed")
		}
		log.Info().Msg("Seed completed successfully")
		return
	}

	// Initialize repositories
	userRepo := pgRepo.NewUserRepository(db)
	refreshTokenRepo := pgRepo.NewRefreshTokenRepository(db)
	propertyRepo := pgRepo.NewPropertyRepository(db)
	propertyImageRepo := pgRepo.NewPropertyImageRepository(db)
	propertyDocRepo := pgRepo.NewPropertyDocumentRepository(db)
	coListingRepo := pgRepo.NewCoListingRepository(db)
	transactor := pgRepo.NewTransactor(db)
	ownerRepo := pgRepo.NewOwnerRepository(db)
	dealRepo := pgRepo.NewDealRepository(db)
	commissionRepo := pgRepo.NewCommissionRepository(db)
	auditLogRepo := pgRepo.NewAuditLogRepository(db)
	masterDataRepo := pgRepo.NewMasterDataRepository(db)

	// Initialize services
	jwtService := service.NewJWTService(cfg.JWTSecret, cfg.JWTAccessExpiration, cfg.JWTRefreshExpiration)
	fileStorage := storage.NewLocalStorage(cfg.UploadPath, cfg.UploadMaxSize)
	documentStorage := storage.NewDocumentStorage(cfg.DocumentPath, cfg.UploadMaxSize)
	auditService := service.NewAuditService(auditLogRepo)

	// Initialize validator
	v := validator.New()

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, refreshTokenRepo, jwtService, usecase.AuthPolicy{
		MaxLoginAttempts:  cfg.MaxLoginAttempts,
		LockoutDuration:   time.Duration(cfg.LockoutMinutes) * time.Minute,
		MaxActiveSessions: cfg.MaxActiveSessions,
	})
	userUsecase := usecase.NewUserUsecase(userRepo, refreshTokenRepo, auditService)
	propertyUsecase := usecase.NewPropertyUsecase(propertyRepo, propertyImageRepo, ownerRepo, coListingRepo, masterDataRepo, propertyDocRepo, transactor, fileStorage, auditService)
	documentUsecase := usecase.NewDocumentUsecase(propertyRepo, coListingRepo, propertyDocRepo, transactor, documentStorage, auditService)
	ownerUsecase := usecase.NewOwnerUsecase(propertyRepo, coListingRepo, ownerRepo, auditService)
	coListingUsecase := usecase.NewCoListingUsecase(propertyRepo, coListingRepo, userRepo, transactor, auditService)
	commissionUsecase := usecase.NewCommissionUsecase(commissionRepo, auditService)
	dealUsecase := usecase.NewDealUsecase(propertyRepo, dealRepo, commissionRepo, transactor, auditService, cfg.CommissionDefaultPercentage)

	// Initialize handlers
	handlers := &routes.Handlers{
		Auth:       handler.NewAuthHandler(authUsecase, userUsecase, v),
		Property:   handler.NewPropertyHandler(propertyUsecase, v),
		Document:   handler.NewDocumentHandler(documentUsecase, v),
		User:       handler.NewUserHandler(userUsecase, v),
		Owner:      handler.NewOwnerHandler(ownerUsecase, v),
		CoListing:  handler.NewCoListingHandler(coListingUsecase, v),
		Deal:       handler.NewDealHandler(dealUsecase, v),
		Commission: handler.NewCommissionHandler(commissionUsecase, v),
		AuditLog:   handler.NewAuditLogHandler(auditLogRepo),
		MasterData: handler.NewMasterDataHandler(masterDataRepo, v),
	}

	// Create Fiber app
	fiberCfg := fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
		BodyLimit:    int(cfg.UploadMaxSize),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if proxies := splitCSV(cfg.TrustedProxies); len(proxies) > 0 {
		fiberCfg.TrustProxy = true
		fiberCfg.TrustProxyConfig = fiber.TrustProxyConfig{Proxies: proxies}
		fiberCfg.ProxyHeader = fiber.HeaderXForwardedFor
	}
	app := fiber.New(fiberCfg)

	// Global middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e any) {
			log.Error().
				Str("request_id", fmt.Sprint(c.Locals("request_id"))).
				Str("method", c.Method()).Str("path", c.Path()).
				Interface("panic", e).Bytes("stack", debug.Stack()).
				Msg("Panic recovered")
		},
	}))
	app.Use(middleware.RequestID())
	app.Use(middleware.SecurityHeaders(cfg.AppEnv == "production"))
	app.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	app.Use(middleware.Logger())

	// Register routes
	routes.Setup(app, handlers, jwtService, userRepo, cfg)

	// Graceful shutdown
	go func() {
		if err := app.Listen(cfg.AppHost + ":" + cfg.AppPort); err != nil {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	log.Info().Str("host", cfg.AppHost).Str("port", cfg.AppPort).Msg("Server started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}

	if err := sqlDB.Close(); err != nil {
		log.Error().Err(err).Msg("Database close error")
	}

	log.Info().Msg("Server stopped")
	logger.Close()
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
