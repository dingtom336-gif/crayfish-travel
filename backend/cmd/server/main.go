package main

import (
	"context"
	"log"

	_ "github.com/xiaozhang/crayfish-travel/backend/docs"
	"github.com/xiaozhang/crayfish-travel/backend/internal/bidding"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/config"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/database"
	redisclient "github.com/xiaozhang/crayfish-travel/backend/internal/common/redis"
	"github.com/xiaozhang/crayfish-travel/backend/internal/identity"
	"github.com/xiaozhang/crayfish-travel/backend/internal/lock"
	"github.com/xiaozhang/crayfish-travel/backend/internal/aiparser"
	"github.com/xiaozhang/crayfish-travel/backend/internal/order"
	"github.com/xiaozhang/crayfish-travel/backend/internal/payment"
	"github.com/xiaozhang/crayfish-travel/backend/internal/riskcontrol"
	"github.com/xiaozhang/crayfish-travel/backend/internal/router"
)

// @title           Crayfish Travel API
// @version         1.0
// @description     Reverse-matching travel bidding platform API.
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Database
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Identity module
	encryptor, err := identity.NewEncryptor(cfg.Security.AESKey)
	if err != nil {
		log.Fatalf("failed to create encryptor: %v", err)
	}
	identityHandler := identity.NewHandler(db, encryptor)

	// NLP module
	lunarSvc, err := aiparser.NewLunarService(2026, 2035)
	if err != nil {
		log.Fatalf("failed to init lunar service: %v", err)
	}
	arkClient := aiparser.NewARKClient(cfg.ARK.APIKey, cfg.ARK.BaseURL, cfg.ARK.Model, cfg.ARK.Temperature, cfg.ARK.MaxTokens, cfg.ARK.Timeout)
	heuristicParser := aiparser.NewHeuristicParser()
	fallbackParser := aiparser.NewFallbackParser(arkClient, heuristicParser)
	dateValidator := aiparser.NewDateValidator(lunarSvc)
	nlpHandler := aiparser.NewHandler(db, fallbackParser, dateValidator)

	// Bidding module (FlyAI real data with Chinese mock fallback)
	supplier := bidding.NewFlyAISupplier()
	biddingHandler := bidding.NewHandler(db, supplier)

	// Redis
	rdb, err := redisclient.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	// Risk Control module
	fundPool := riskcontrol.NewFundPool(db)
	antifraud := riskcontrol.NewAntifraudChecker(rdb)
	hedgeCalculator := riskcontrol.NewHedgeCalculator(db)
	refundProcessor := riskcontrol.NewRefundProcessor(db, fundPool, hedgeCalculator, antifraud)
	riskcontrolHandler := riskcontrol.NewHandler(fundPool, antifraud)

	// Lock module
	orchestrator := lock.NewOrchestrator(db, rdb, fundPool)
	lockHandler := lock.NewHandler(orchestrator)

	// Start lock expiry watcher
	lockTimer := lock.NewLockTimer(orchestrator, rdb)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go lockTimer.StartExpiryWatcher(ctx)

	// Payment module
	alipayClient := payment.NewMockAlipayClient()
	callbackProcessor := payment.NewCallbackProcessor(db, rdb, alipayClient)
	paymentHandler := payment.NewHandler(db, alipayClient, callbackProcessor)

	// Order module
	smsNotifier := order.NewMockSmsNotifier()
	orderService := order.NewOrderService(db, encryptor)
	orderHandler := order.NewHandler(orderService, smsNotifier, refundProcessor)

	// Wire callback -> auto order creation
	callbackProcessor.SetOrderCreator(orderService)

	// Router
	handlers := &router.Handlers{
		Identity:    identityHandler,
		NLP:         nlpHandler,
		Bidding:     biddingHandler,
		Lock:        lockHandler,
		Payment:     paymentHandler,
		Order:       orderHandler,
		RiskControl: riskcontrolHandler,
	}

	deps := router.Dependencies{DB: db, Redis: rdb}
	r := router.Setup(cfg.Server.Mode, handlers, cfg.AllowedOrigins, cfg.AdminToken, deps)

	log.Printf("Crayfish Travel server starting on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
