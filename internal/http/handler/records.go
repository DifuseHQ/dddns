package handler

import (
	"github.com/DifuseHQ/dddns/internal/db"
	"github.com/DifuseHQ/dddns/internal/db/model"
	"github.com/gofiber/fiber/v2"
)

func CreateRecord(c *fiber.Ctx) error {
	uuid := c.Query("uuid")

	type RequestBody struct {
		Domain string `json:"domain"`
		IPv4   string `json:"ipv4"`
		IPv6   string `json:"ipv6"`
	}

	var body RequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON sent by client"})
	}

	record := &model.Record{
		UUID:       uuid,
		Domain:     body.Domain,
		ARecord:    body.IPv4,
		AAAARecord: body.IPv6,
	}

	success, err := db.InsertOrUpdateRecord(db.Database, record, record.Domain)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert or update record"})
	}

	if success {
		return c.JSON(fiber.Map{
			"message": "Record successfully created or updated",
			"record":  record,
		})
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Record update failed"})
	}
}

func DeleteRecord(c *fiber.Ctx) error {
	uuid := c.Query("uuid")

	success, err := db.DeleteRecord(db.Database, uuid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete record"})
	}

	if success {
		return c.JSON(fiber.Map{
			"message": "Record successfully deleted",
		})
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Record deletion failed"})
	}
}
