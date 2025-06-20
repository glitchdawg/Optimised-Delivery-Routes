package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/db"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/models"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/services"
)

func GetAgents(c *gin.Context) {
	warehouseID := c.Query("warehouse_id")

	var rows *sql.Rows
	var err error

	if warehouseID != "" {
		rows, err = db.DB.Query(`SELECT id, name, warehouse_id, checked_in_at FROM agents WHERE warehouse_id = $1`, warehouseID)
	} else {
		rows, err = db.DB.Query(`SELECT id, name, warehouse_id, checked_in_at FROM agents`)
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var agents []models.Agent
	for rows.Next() {
		var a models.Agent
		rows.Scan(&a.ID, &a.Name, &a.WarehouseID, &a.CheckedInAt)
		agents = append(agents, a)
	}

	c.JSON(200, agents)
}

func GetAgentOrders(c *gin.Context) {
	agentID := c.Param("id")

	rows, err := db.DB.Query(`
        SELECT o.id, o.delivery_address, o.lat, o.lon
        FROM orders o
        JOIN agent_assignments a ON o.id = a.order_id
        WHERE a.agent_id = $1 AND a.assigned_on = CURRENT_DATE
    `, agentID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		rows.Scan(&o.ID, &o.DeliveryAddress, &o.Lat, &o.Lon)
		orders = append(orders, o)
	}

	c.JSON(200, orders)
}

func GetAgentPayout(c *gin.Context) {
	agentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent ID"})
		return
	}

	payout, err := services.CalculatePayoutForAgent(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payout)
}

func AddAgent(c *gin.Context) {
	var input struct {
		Name        string  `json:"name"`
		WarehouseID int     `json:"warehouse_id"`
		// Latitude    float64 `json:"latitude" binding:"required"`
		// Longitude   float64 `json:"longitude" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec(`INSERT INTO agents (name, warehouse_id, checked_in_at) VALUES ($1, $2, CURRENT_TIMESTAMP)`,
		input.Name, input.WarehouseID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "agent added ✅"})
}

func GetWarehouses(c *gin.Context) {
	rows, err := db.DB.Query(`SELECT id, name, lat, lon FROM warehouses`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var warehouses []models.Warehouse
	for rows.Next() {
		var w models.Warehouse
		rows.Scan(&w.ID, &w.Name, &w.Lat, &w.Lon)
		warehouses = append(warehouses, w)
	}

	c.JSON(200, warehouses)
}

func AddWarehouse(c *gin.Context) {
	var input struct {
		Name string  `json:"name" binding:"required"`
		Lat  float64 `json:"lat" binding:"required"`
		Lon  float64 `json:"lon" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec(`INSERT INTO warehouses (name, lat, lon) VALUES ($1, $2, $3)`, input.Name, input.Lat, input.Lon)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "warehouse created ✅"})
}
