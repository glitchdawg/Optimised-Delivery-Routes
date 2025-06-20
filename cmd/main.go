package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/db"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/handlers"
)

func main() {
	err := db.InitDB()
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	r := gin.Default()

	// API Routes
	r.GET("/agents", handlers.GetAgents)
	r.GET("/agents/:id/orders", handlers.GetAgentOrders)
	r.GET("/agents/:id/payout", handlers.GetAgentPayout)
	r.POST("/allocate", handlers.TriggerAllocation)

	r.Run(":8080")
}
