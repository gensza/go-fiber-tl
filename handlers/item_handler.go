package handlers

import (
	"strconv"
	"time"

	"go-fiber-tablelink/config"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetItems(c *fiber.Ctx) error {

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
		SELECT uuid, name, price, status
		FROM tm_item
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
		var price float64
		var status int
		var id, name string

		if err := rows.Scan(&id, &name, &price, &status); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		data = append(data, fiber.Map{
			"id":     id,
			"name":   name,
			"price":  price,
			"status": status,
		})
	}

	var total int
	config.DB.QueryRow(`
		SELECT COUNT(*) FROM tm_item WHERE deleted_at IS NULL
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

func CreateItem(c *fiber.Ctx) error {
	var req struct {
		Name   string  `json:"name"`
		Price  float64 `json:"price"`
		Status int     `json:"status"`
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
		FROM tm_item
		WHERE name = $1
		AND deleted_at IS NULL
	`, req.Name).Scan(&count)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if count > 0 {
		return c.Status(409).JSON(fiber.Map{
			"error": "item name already exists",
		})
	}

	id := uuid.New().String()
	// 📝 Insert data
	_, err = config.DB.Exec(`
	INSERT INTO tm_item (uuid, name, price, status, created_at)
	VALUES ($1, $2, $3, $4, $5)
`, id, req.Name, req.Price, req.Status, time.Now())

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "item created successfully",
	})
}

func UpdateItem(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "uuid is required",
		})
	}

	var req struct {
		Name   string  `json:"name"`
		Price  float64 `json:"price"`
		Status int     `json:"status"`
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
		FROM tm_item
		WHERE name = $1
		  AND uuid <> $2
		  AND deleted_at IS NULL
	`, req.Name, uuid).Scan(&count)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if count > 0 {
		return c.Status(409).JSON(fiber.Map{
			"error": "item name already exists",
		})
	}

	// 📝 Update data
	result, err := config.DB.Exec(`
		UPDATE tm_item
		SET name = $1,
		    price = $2,
		    status = $3,
		    updated_at = $4
		WHERE uuid = $5
		  AND deleted_at IS NULL
	`, req.Name, req.Price, req.Status, time.Now(), uuid)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "item not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "item updated successfully",
	})
}

func DeleteItem(c *fiber.Ctx) error {
	uuid := c.Params("uuid")

	// Guard: uuid wajib
	if uuid == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "uuid is required",
		})
	}

	// Soft delete
	result, err := config.DB.Exec(`
		UPDATE tm_item
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
			"error": "item not found or already deleted",
		})
	}

	return c.JSON(fiber.Map{
		"message": "item deleted successfully",
	})
}
