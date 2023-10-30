package handler

import (
	"fmt"
	"github.com/DifuseHQ/dddns/internal/db"
	"github.com/DifuseHQ/dddns/internal/utils"
	"github.com/DifuseHQ/dddns/pkg/config"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func IsDomainAvailable(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		domain := c.Params("domain")

		if domain == "" {
			return c.JSON(fiber.Map{
				"error": fmt.Errorf("Missing domain").Error(),
			})
		}

		validate := validator.New()
		err := validate.Var(domain, "required,fqdn")

		if err != nil {
			return c.JSON(fiber.Map{
				"error": fmt.Errorf("Invalid domain").Error(),
			})
		}

		if !utils.DomainEndsWith(domain, cfg.Domain) {
			return c.JSON(fiber.Map{
				"error": fmt.Errorf("Invalid domain, needs to be a subdomain of %s", cfg.Domain).Error(),
			})
		}

		var count int

		query := `SELECT COUNT(*) FROM records WHERE domain = ?`

		err = db.Database.QueryRow(query, domain).Scan(&count)

		if err != nil {
			return c.JSON(fiber.Map{
				"error": fmt.Errorf("Domain availablitity check failed").Error(),
			})
		}

		return c.JSON(fiber.Map{
			"available": count == 0,
		})
	}
}

func IsDomainTakenByElse(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		domain := c.Params("domain")
		uuid := c.Query("uuid")

		if domain == "" {
			return c.JSON(fiber.Map{
				"error": fmt.Errorf("Missing domain").Error(),
			})
		}

		validate := validator.New()
		err := validate.Var(domain, "required,fqdn")

		if err != nil {
			return c.JSON(fiber.Map{
				"error": fmt.Errorf("Invalid domain").Error(),
			})
		}

		if !utils.DomainEndsWith(domain, cfg.Domain) {
			return c.JSON(fiber.Map{
				"error": fmt.Errorf("Invalid domain, needs to be a subdomain of %s", cfg.Domain).Error(),
			})
		}

		var count int

		query := `SELECT COUNT(*) FROM records WHERE domain = ? AND uuid != ?`

		err = db.Database.QueryRow(query, domain, uuid).Scan(&count)

		if err != nil {
			return c.JSON(fiber.Map{
				"error": fmt.Errorf("Domain availablitity check failed").Error(),
			})
		}

		return c.JSON(fiber.Map{
			"available": count == 0,
		})
	}
}
