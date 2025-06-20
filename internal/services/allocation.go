package services

import (
	"math"
	"sort"

	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/db"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/models"
)

const (
	MaxDistanceKM    = 100.0
	MaxTimeMinutes   = 600
	AvgSpeedKMPerMin = 0.2 // 1 km in 5 min = 0.2 km/min
)

func AllocateOrdersForWarehouse(warehouse models.Warehouse) error {
	
	agents, err := fetchCheckedInAgents(warehouse.ID)
	if err != nil {
		return err
	}

	
	orders, err := fetchUnassignedOrders(warehouse.ID)
	if err != nil {
		return err
	}

	for i := range orders {
		dist := haversine(warehouse.Lat, warehouse.Lon, orders[i].Lat, orders[i].Lon)
		orders[i].DistanceKM = dist
		orders[i].EstimatedTimeMinutes = int(dist / AvgSpeedKMPerMin)
	}

	sortOrdersByDistance(&orders)

	currOrder := 0
	for _, agent := range agents {
		var usedKM float64
		var usedMinutes int
		var assigned []models.Order

		for currOrder < len(orders) {
			o := orders[currOrder]
			if o.Assigned {
				currOrder++
				continue
			}

			if usedKM+o.DistanceKM > MaxDistanceKM || usedMinutes+o.EstimatedTimeMinutes > MaxTimeMinutes {
				break
			}

			assigned = append(assigned, o)
			usedKM += o.DistanceKM
			usedMinutes += o.EstimatedTimeMinutes
			orders[currOrder].Assigned = true
			currOrder++
		}

		err := saveAssignments(agent.ID, assigned)
		if err != nil {
			return err
		}
	}

	return nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
func fetchCheckedInAgents(warehouseID int) ([]models.Agent, error) {
	rows, err := db.DB.Query(`
        SELECT id, name, warehouse_id, checked_in_at
        FROM agents
        WHERE warehouse_id = $1 AND checked_in_at::date = CURRENT_DATE
    `, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []models.Agent
	for rows.Next() {
		var a models.Agent
		if err := rows.Scan(&a.ID, &a.Name, &a.WarehouseID, &a.CheckedInAt); err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}
	return agents, nil
}
func fetchUnassignedOrders(warehouseID int) ([]models.Order, error) {
	rows, err := db.DB.Query(`
        SELECT id, warehouse_id, lat, lon, delivery_address, scheduled_for, assigned
        FROM orders
        WHERE warehouse_id = $1 AND scheduled_for = CURRENT_DATE AND assigned = false
    `, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.WarehouseID, &o.Lat, &o.Lon, &o.DeliveryAddress, &o.ScheduledFor, &o.Assigned); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func sortOrdersByDistance(orders *[]models.Order) {
	sort.Slice(*orders, func(i, j int) bool {
		return (*orders)[i].DistanceKM < (*orders)[j].DistanceKM
	})
}

func saveAssignments(agentID int, orders []models.Order) error {
    tx, err := db.DB.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare(`
        INSERT INTO agent_assignments (agent_id, order_id, assigned_on, distance_km, estimated_time_minutes)
        VALUES ($1, $2, CURRENT_DATE, $3, $4)
    `)
    if err != nil {
        tx.Rollback()
        return err
    }
    defer stmt.Close()

    updateStmt, err := tx.Prepare(`UPDATE orders SET assigned = true WHERE id = $1`)
    if err != nil {
        tx.Rollback()
        return err
    }
    defer updateStmt.Close()

    for _, order := range orders {
        _, err := stmt.Exec(agentID, order.ID, order.DistanceKM, order.EstimatedTimeMinutes)
        if err != nil {
            tx.Rollback()
            return err
        }

        _, err = updateStmt.Exec(order.ID)
        if err != nil {
            tx.Rollback()
            return err
        }
    }

    return tx.Commit()
}
