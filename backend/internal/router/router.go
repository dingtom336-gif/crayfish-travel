package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/xiaozhang/crayfish-travel/backend/internal/bidding"
	"github.com/xiaozhang/crayfish-travel/backend/internal/common/middleware"
	"github.com/xiaozhang/crayfish-travel/backend/internal/identity"
	"github.com/xiaozhang/crayfish-travel/backend/internal/lock"
	"github.com/xiaozhang/crayfish-travel/backend/internal/nlp"
	"github.com/xiaozhang/crayfish-travel/backend/internal/order"
	"github.com/xiaozhang/crayfish-travel/backend/internal/payment"
	"github.com/xiaozhang/crayfish-travel/backend/internal/riskcontrol"
)

// Handlers holds all module handlers for dependency injection.
type Handlers struct {
	Identity    *identity.Handler
	NLP         *nlp.Handler
	Bidding     *bidding.Handler
	Lock        *lock.Handler
	Payment     *payment.Handler
	Order       *order.Handler
	RiskControl *riskcontrol.Handler
}

// Setup configures all routes and middleware.
func Setup(mode string, h *Handlers, allowedOrigins []string, adminToken string) *gin.Engine {
	gin.SetMode(mode)
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(allowedOrigins))
	r.Use(middleware.TraceID())
	r.Use(middleware.Metrics())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"trace_id": middleware.GetTraceID(c),
		})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Prometheus metrics
	r.GET("/metrics", middleware.MetricsHandler())

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Sessions (lightweight, no PII)
		v1.POST("/sessions", h.Identity.CreateSession)
		v1.GET("/sessions/:session_id", h.Identity.GetSession)

		// Identity (real handlers)
		identityGroup := v1.Group("/identity")
		{
			identityGroup.POST("", h.Identity.Create)
			identityGroup.DELETE("/:session_id", h.Identity.Delete)
		}

		// NLP (real handlers)
		nlpGroup := v1.Group("/nlp")
		{
			nlpGroup.POST("/parse", h.NLP.Parse)
			nlpGroup.POST("/confirm", h.NLP.Confirm)
		}

		// Bidding (real handlers)
		biddingGroup := v1.Group("/bidding")
		{
			biddingGroup.POST("/start", h.Bidding.Start)
			biddingGroup.GET("/:session_id/packages", h.Bidding.GetPackages)
		}

		// Lock (real handlers)
		lockGroup := v1.Group("/lock")
		{
			lockGroup.POST("/acquire", h.Lock.Acquire)
			lockGroup.GET("/:session_id/status", h.Lock.Status)
		}

		// Payment (real handlers)
		paymentGroup := v1.Group("/payment")
		{
			paymentGroup.POST("/create", h.Payment.Create)
			paymentGroup.POST("/callback", h.Payment.Callback)
		}

		// Orders (real handlers)
		ordersGroup := v1.Group("/orders")
		{
			ordersGroup.GET("", h.Order.List)
			ordersGroup.GET("/:id", h.Order.Get)
			ordersGroup.POST("/:id/refund", h.Order.RequestRefund)
		}

		// Risk Control (real handlers)
		riskcontrolGroup := v1.Group("/riskcontrol")
		{
			riskcontrolGroup.GET("/pool/status", h.RiskControl.PoolStatus)
		}

		// Admin (dev only, requires token)
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminAuth(adminToken))
		{
			admin.POST("/fund-pool/seed", h.RiskControl.Seed)
		}
	}

	return r
}

// placeholder returns a handler that echoes the endpoint name with trace ID.
func placeholder(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"endpoint": name,
			"status":   "stub",
			"trace_id": middleware.GetTraceID(c),
		})
	}
}
