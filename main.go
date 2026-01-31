package main

import (
	"go-fiber-tablelink/config"
	"go-fiber-tablelink/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.ConnectDB()

	app := fiber.New()
	routes.IngredientRoutes(app)

	app.Listen(":8080")
}
