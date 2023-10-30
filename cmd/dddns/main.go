package main

import (
	"fmt"
	"github.com/DifuseHQ/dddns/internal/db"
	"github.com/DifuseHQ/dddns/internal/dns"
	"github.com/DifuseHQ/dddns/internal/http/handler"
	"github.com/DifuseHQ/dddns/internal/http/middleware"
	"github.com/DifuseHQ/dddns/pkg/config"
	"github.com/DifuseHQ/dddns/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	cfg := config.InitConfig()

	logger.InitLogger(cfg.LogPath)
	logger.Log.Info("Starting DDDNS")

	db.InitDB(cfg.Domain)

	go dns.InitDNSServer(cfg.DNSAddr, cfg.DNSPort)
	logger.Log.Info("DNS server initialized")

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	checks := app.Group("/checks", cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	checks.Get("/is-domain-available/:domain", middleware.UUIDCheckMiddleware, handler.IsDomainAvailable(cfg))
	checks.Get("/is-domain-taken-by-someone/:domain", middleware.UUIDCheckMiddleware, handler.IsDomainTakenByElse(cfg))

	manageRecords := app.Group("/manage-record", cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	manageRecords.Post("/create-or-update", middleware.UUIDCheckMiddleware, handler.CreateRecord)
	manageRecords.Get("/delete", middleware.UUIDCheckMiddleware, handler.DeleteRecord)

	err := app.Listen(fmt.Sprintf("%s:%s", cfg.HTTPAddr, cfg.HTTPPort))

	if err != nil {
		logger.Log.Fatal("Failed to start HTTP server ", err)
	}
}
