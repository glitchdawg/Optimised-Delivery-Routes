package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/db"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/handlers"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	err = db.InitDB()
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/agents", handlers.GetAgents)
	r.GET("/agents/:id/orders", handlers.GetAgentOrders)
	r.GET("/agents/:id/payout", handlers.GetAgentPayout)
	r.POST("/allocate", handlers.TriggerAllocation)
	r.POST("/agents", handlers.AddAgent)
	r.POST("/warehouses", handlers.AddWarehouse)
	r.GET("/warehouses", handlers.GetWarehouses)
	r.POST("/orders", handlers.CreateOrder)
	r.GET("/orders", handlers.GetOrders)

	r.Run(":8080")
}
