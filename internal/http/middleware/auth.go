package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/DifuseHQ/dddns/internal/db"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func ValidateUUID(uuid string) (bool, error) {
	err := db.Database.QueryRow(`SELECT uuid FROM records WHERE uuid = $1`, uuid).Scan(&uuid)

	if err != nil {
		if err != sql.ErrNoRows {
			return false, fmt.Errorf("db connection failed")
		}
	}

	url := fmt.Sprintf("https://gin.difuse.io/harpoon/verify-identifier/%s", uuid)
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("error checking with remote db")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("invalid uuid")
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, fmt.Errorf("error decoding response from server")
	}

	if valid, ok := result["valid"].(bool); ok && valid {
		return true, nil
	}

	return false, nil
}

func UUIDCheckMiddleware(c *fiber.Ctx) error {
	uuid := c.Query("uuid")

	if uuid == "" || len(uuid) != 36 {
		return c.JSON(fiber.Map{
			"error": "Missing or invalid UUID",
		})
	} else {
		valid, err := ValidateUUID(uuid)

		if err != nil {
			return c.JSON(fiber.Map{
				"error": err,
			})
		}

		if !valid {
			return c.JSON(fiber.Map{
				"error": "Invalid UUID",
			})
		}
	}

	return c.Next()
}
