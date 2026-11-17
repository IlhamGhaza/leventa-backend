package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/leventa-backend/docs"
	"github.com/leventa-backend/internal/config"
	"github.com/leventa-backend/internal/delivery/http/handler"
	"github.com/leventa-backend/internal/delivery/http/middleware"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/service"
)

// Handlers holds all handler instances.
type Handlers struct {
	Auth       *handler.AuthHandler
	Property   *handler.PropertyHandler
	Document   *handler.DocumentHandler
	User       *handler.UserHandler
	Owner      *handler.OwnerHandler
	CoListing  *handler.CoListingHandler
	Deal       *handler.DealHandler
	Commission *handler.CommissionHandler
	AuditLog   *handler.AuditLogHandler
	MasterData *handler.MasterDataHandler
}

// Setup registers all routes.
func Setup(app *fiber.App, h *Handlers, jwtService *service.JWTService, userLookup middleware.UserLookup, cfg *config.Config) {
	// Static files (uploads)
	app.Use("/uploads", static.New(cfg.UploadPath))

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// OpenAPI spec and Swagger UI.
	app.Get("/openapi.yaml", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/yaml")
		return c.Send(docs.OpenAPI)
	})
	app.Get("/swagger", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentSecurityPolicy,
			"default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline' https://unpkg.com; img-src 'self' data:; connect-src 'self'")
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(swaggerHTML)
	})

	// Every API request is rate limited per IP; login/refresh and uploads have
	// stricter limits on top.
	api := app.Group("/api/v1", middleware.RateLimiter(cfg.APIRateLimitMax, cfg.APIRateLimitWindow, "api"))
	uploadLimiter := middleware.RateLimiter(cfg.UploadRateLimitMax, cfg.UploadRateLimitWindow, "upload")

	// Rate limiter for auth endpoints
	authRateLimiter := middleware.RateLimiter(cfg.RateLimitMax, cfg.RateLimitWindow, "auth")

	// --- Auth Routes ---
	auth := api.Group("/auth")
	auth.Post("/login", authRateLimiter, h.Auth.Login)
	auth.Post("/refresh", authRateLimiter, h.Auth.Refresh)
	auth.Post("/logout", middleware.Auth(jwtService, userLookup), h.Auth.Logout)
	auth.Get("/me", middleware.Auth(jwtService, userLookup), h.Auth.Me)
	// Self-service: own WhatsApp number only (the public Marketing contact).
	auth.Patch("/me", middleware.Auth(jwtService, userLookup), h.Auth.UpdateMe)

	// --- Property Routes ---
	// Route middleware only rejects unauthenticated/obviously wrong roles;
	// ownership and visibility are enforced in the use cases.
	authMW := middleware.Auth(jwtService, userLookup)
	properties := api.Group("/properties")
	properties.Get("/", middleware.OptionalAuth(jwtService, userLookup), h.Property.List)
	properties.Get("/:id", middleware.OptionalAuth(jwtService, userLookup), h.Property.GetByID)
	properties.Post("/", authMW, middleware.RequireRole(entity.RoleSales), h.Property.Create)
	properties.Patch("/:id", authMW, middleware.RequireRole(entity.RoleSales), h.Property.Update)
	properties.Put("/:id", authMW, middleware.RequireRole(entity.RoleSales), h.Property.Update)
	properties.Delete("/:id", authMW, middleware.RequireRole(entity.RoleSales), h.Property.Delete)
	properties.Put("/:id/status", authMW, middleware.RequireRole(entity.RoleSales), h.Property.UpdateStatus)
	properties.Post("/:id/images", uploadLimiter, authMW, middleware.RequireRole(entity.RoleSales), h.Property.UploadImage)
	properties.Delete("/:id/images/:imageId", authMW, middleware.RequireRole(entity.RoleSales), h.Property.DeleteImage)
	properties.Put("/:id/images/:imageId/thumbnail", authMW, middleware.RequireRole(entity.RoleSales), h.Property.SetThumbnail)

	// Ownership documents: private files, primary/co-listing marketing only (D-029).
	properties.Get("/:id/documents", authMW, middleware.RequireRole(entity.RoleSales), h.Document.List)
	properties.Post("/:id/documents", uploadLimiter, authMW, middleware.RequireRole(entity.RoleSales), h.Document.Upload)
	properties.Get("/:id/documents/:docId/download", authMW, middleware.RequireRole(entity.RoleSales), h.Document.Download)
	properties.Delete("/:id/documents/:docId", authMW, middleware.RequireRole(entity.RoleSales), h.Document.Delete)

	// Owner PII: reachable only via the property, primary marketing only.
	properties.Get("/:id/owner", authMW, middleware.RequireRole(entity.RoleSales), h.Owner.Get)
	properties.Patch("/:id/owner", authMW, middleware.RequireRole(entity.RoleSales), h.Owner.Update)

	// Co-listing management (admin writes; admin/primary/co-listing read).
	properties.Get("/:id/co-listings", authMW, middleware.RequireRole(entity.RoleAdmin, entity.RoleSales), h.CoListing.List)
	properties.Post("/:id/co-listings", authMW, middleware.RequireRole(entity.RoleAdmin), h.CoListing.Add)
	properties.Post("/:id/transfer-ownership", authMW, middleware.RequireRole(entity.RoleAdmin), h.CoListing.Transfer)
	properties.Put("/:id/co-listings/:marketingId", authMW, middleware.RequireRole(entity.RoleAdmin), h.CoListing.Replace)
	properties.Delete("/:id/co-listings/:marketingId", authMW, middleware.RequireRole(entity.RoleAdmin), h.CoListing.Remove)

	// --- Deal Routes ---
	// Admin records the outcome on a property; Marketing only reads own deals.
	properties.Post("/:id/deal", authMW, middleware.RequireRole(entity.RoleAdmin), h.Deal.RecordDeal)
	properties.Post("/:id/cancel", authMW, middleware.RequireRole(entity.RoleAdmin), h.Deal.Cancel)
	deals := api.Group("/deals", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleSales, entity.RoleAdmin))
	deals.Get("/", h.Deal.List)
	deals.Get("/:id", h.Deal.GetByID)

	// --- Commission Routes ---
	commissions := api.Group("/commissions", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleSales, entity.RoleAdmin))
	commissions.Get("/", h.Commission.List)
	commissions.Get("/:id", h.Commission.GetByID)
	commissions.Patch("/:id", middleware.RequireRole(entity.RoleAdmin), h.Commission.UpdatePercentage)
	commissions.Patch("/:id/status", middleware.RequireRole(entity.RoleAdmin), h.Commission.UpdateStatus)

	// --- User Routes (Admin only) ---
	users := api.Group("/users", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin))
	users.Get("/", h.User.List)
	users.Get("/:id", h.User.GetByID)
	users.Post("/", h.User.Create)
	users.Put("/:id", h.User.Update)
	users.Patch("/:id/status", h.User.UpdateStatus)
	users.Patch("/:id/role", h.User.UpdateRole)
	users.Delete("/:id", h.User.Delete)

	// --- Audit Log Routes ---
	auditLogs := api.Group("/audit-logs", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin))
	auditLogs.Get("/", h.AuditLog.List)

	// --- Master Data Routes ---
	masterData := api.Group("/master-data")
	masterData.Get("/provinces", h.MasterData.ListProvinces)
	masterData.Get("/cities", h.MasterData.ListCities)
	masterData.Get("/districts", h.MasterData.ListDistricts)
	masterData.Post("/districts", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.CreateDistrict)
	masterData.Put("/districts/:id", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.UpdateDistrict)
	masterData.Delete("/districts/:id", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.DeleteDistrict)
	masterData.Get("/facilities", h.MasterData.ListFacilities)
	masterData.Post("/facilities", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.CreateFacility)
	masterData.Put("/facilities/:id", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.UpdateFacility)
	masterData.Delete("/facilities/:id", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.DeleteFacility)
	masterData.Post("/provinces", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.CreateProvince)
	masterData.Put("/provinces/:id", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.UpdateProvince)
	masterData.Delete("/provinces/:id", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.DeleteProvince)
	masterData.Put("/cities/:id", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.UpdateCity)
	masterData.Delete("/cities/:id", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.DeleteCity)
	masterData.Post("/cities", middleware.Auth(jwtService, userLookup), middleware.RequireRole(entity.RoleAdmin), h.MasterData.CreateCity)
}

const swaggerHTML = `<!doctype html>
<html><head><meta charset="utf-8"><title>PMS API</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css"></head>
<body><div id="ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({url: "/openapi.yaml", dom_id: "#ui"});</script>
</body></html>`
