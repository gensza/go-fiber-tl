package handlers

import (
	"strconv"
	"time"

	"go-fiber-tablelink/config"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetIngredients(c *fiber.Ctx) error {

	limit := c.Query("limit")
	page := c.Query("page")
	offset := c.Query("offset")

	pageInt := 1
	maxLimitInt := 1000
	limitInt := 10
	pageOffset := 0

	if limit != "" {
		limitInt, _ = strconv.Atoi(limit)
	}
	if limit != "" && limitInt > maxLimitInt {
		limitInt = maxLimitInt
	}
	if offset != "" {
		pageOffset, _ = strconv.Atoi(offset)
	}
	if page != "" {
		pageInt, _ = strconv.Atoi(page)
		pageOffset = (limitInt * pageInt) - limitInt
	}

	rows, err := config.DB.Query(`
		SELECT uuid, name, cause_alergy, type, status
		FROM tm_ingredient
		WHERE deleted_at IS NULL
		ORDER BY uuid DESC
		LIMIT $1 OFFSET $2
	`, limitInt, pageOffset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var data []fiber.Map
	for rows.Next() {
		var itype, status int
		var id, name string
		var causeAlergy bool

		if err := rows.Scan(&id, &name, &causeAlergy, &itype, &status); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		data = append(data, fiber.Map{
			"id":           id,
			"name":         name,
			"cause_alergy": causeAlergy,
			"type":         itype,
			"status":       status,
		})
	}

	var total int
	config.DB.QueryRow(`
		SELECT COUNT(*) FROM tm_ingredient WHERE deleted_at IS NULL
	`).Scan(&total)

	return c.JSON(fiber.Map{
		"data": data,
		"meta": fiber.Map{
			"page":  pageInt,
			"limit": limitInt,
			"total": total,
		},
	})
}

func CreateIngredient(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		CauseAlergy bool   `json:"cause_alergy"`
		Type        int    `json:"type"`
		Status      int    `json:"status"`
	}

	// Parse JSON body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Basic validation
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	// 🔍 Check unique name (exclude soft deleted)
	var count int
	err := config.DB.QueryRow(`
		SELECT COUNT(*)
		FROM tm_ingredient
		WHERE name = $1
		AND deleted_at IS NULL
	`, req.Name).Scan(&count)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if count > 0 {
		return c.Status(409).JSON(fiber.Map{
			"error": "ingredient name already exists",
		})
	}

	id := uuid.New().String()
	// 📝 Insert data
	_, err = config.DB.Exec(`
	INSERT INTO tm_ingredient (uuid, name, cause_alergy, type, status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
`, id, req.Name, req.CauseAlergy, req.Type, req.Status, time.Now())

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "ingredient created successfully",
	})
}

func UpdateIngredient(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "uuid is required",
		})
	}

	var req struct {
		Name        string `json:"name"`
		CauseAlergy bool   `json:"cause_alergy"`
		Type        int    `json:"type"`
		Status      int    `json:"status"`
	}

	// Parse body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Basic validation
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	// 🔍 Check unique name (exclude current uuid + soft deleted)
	var count int
	err := config.DB.QueryRow(`
		SELECT COUNT(*)
		FROM tm_ingredient
		WHERE name = $1
		  AND uuid <> $2
		  AND deleted_at IS NULL
	`, req.Name, uuid).Scan(&count)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if count > 0 {
		return c.Status(409).JSON(fiber.Map{
			"error": "ingredient name already exists",
		})
	}

	// 📝 Update data
	result, err := config.DB.Exec(`
		UPDATE tm_ingredient
		SET name = $1,
		    cause_alergy = $2,
		    type = $3,
		    status = $4,
		    updated_at = $5
		WHERE uuid = $6
		  AND deleted_at IS NULL
	`, req.Name, req.CauseAlergy, req.Type, req.Status, time.Now(), uuid)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "ingredient not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "ingredient updated successfully",
	})
}

func DeleteIngredient(c *fiber.Ctx) error {
	uuid := c.Params("uuid")

	// Guard: uuid wajib
	if uuid == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "uuid is required",
		})
	}

	// Soft delete
	result, err := config.DB.Exec(`
		UPDATE tm_ingredient
		SET deleted_at = NOW()
		WHERE uuid = $1
		  AND deleted_at IS NULL
	`, uuid)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "ingredient not found or already deleted",
		})
	}

	return c.JSON(fiber.Map{
		"message": "ingredient deleted successfully",
	})
}
