package handlers

import (
	"go-fiber-tablelink/config"

	"github.com/gofiber/fiber/v2"
)

func DeleteItemIngredients(c *fiber.Ctx) error {
	var req struct {
		UuidItem       string `json:"uuid_item"`
		UuidIngredient string `json:"uuid_ingredient"`
	}

	// Parse JSON body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Guard: uuid wajib
	if req.UuidItem == "" || req.UuidIngredient == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "uuid is required",
		})
	}

	// Soft delete
	result, err := config.DB.Exec(`
		DELETE FROM tm_item_ingredient
		WHERE uuid_item = $1
		AND uuid_ingredient = $2
	`, req.UuidItem, req.UuidIngredient)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "item or ingredient not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "item ingredient deleted successfully",
	})
}
