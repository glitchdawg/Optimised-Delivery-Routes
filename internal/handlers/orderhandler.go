package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/db"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/models"
)

func GetOrders(c *gin.Context) {
	rows, err := db.DB.Query(`SELECT id, warehouse_id, lat, lon, delivery_address, assigned, agent_id FROM orders`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.WarehouseID, &o.Lat, &o.Lon, &o.DeliveryAddress, &o.Assigned, &o.AgentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, orders)

}
func CreateOrder(c *gin.Context) {
	var o models.Order
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(
		`INSERT INTO orders (warehouse_id, delivery_address, lat, lon, assigned, agent_id) VALUES ($1, $2, $3, $4, false, NULL)`,
		o.WarehouseID, o.DeliveryAddress, o.Lat, o.Lon,
	)
	if err != nil {
		// Add this line:
		fmt.Println("CreateOrder error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order created"})
}
