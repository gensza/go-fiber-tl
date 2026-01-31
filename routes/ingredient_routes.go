package routes

import (
	"go-fiber-tablelink/handlers"

	"github.com/gofiber/fiber/v2"
)

func IngredientRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/ingredients", handlers.GetIngredients)
	api.Post("/ingredients", handlers.CreateIngredient)
	api.Put("/ingredients/:uuid", handlers.UpdateIngredient)
	api.Delete("/ingredients/:uuid", handlers.DeleteIngredient)

	api.Get("/items", handlers.GetItems)
	api.Post("/items", handlers.CreateItem)
	api.Put("/items/:uuid", handlers.UpdateItem)
	api.Delete("/items/:uuid", handlers.DeleteItem)

	api.Delete("/items-ingredients", handlers.DeleteItemIngredients)

}
