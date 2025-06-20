package handlers

import (
    "github.com/glitchdawg/Optimised-Delivery-Routes/internal/services"
    "github.com/glitchdawg/Optimised-Delivery-Routes/internal/db"
    "github.com/glitchdawg/Optimised-Delivery-Routes/internal/models"
    "github.com/gin-gonic/gin"
    "net/http"
)

func TriggerAllocation(c *gin.Context) {
    rows, err := db.DB.Query(`SELECT id, name, lat, lon FROM warehouses`)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    var warehouses []models.Warehouse
    for rows.Next() {
        var w models.Warehouse
        if err := rows.Scan(&w.ID, &w.Name, &w.Lat, &w.Lon); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        warehouses = append(warehouses, w)
    }

    for _, w := range warehouses {
        err := services.AllocateOrdersForWarehouse(w)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "warehouse_id": w.ID,
                "error":        err.Error(),
            })
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Allocation complete ✅"})
}
