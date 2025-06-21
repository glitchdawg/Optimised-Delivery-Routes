package services

import (
	"math"
	"sort"
	"time"

	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/db"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/models"
)

const (
	MaxDistanceKM    = 100.0
	MaxTimeMinutes   = 600
	AvgSpeedKMPerMin = 0.2 // 1 km in 5 min = 0.2 km/min
)

type AgentCapacity struct {
	Agent       models.Agent
	UsedKM      float64
	UsedMinutes int
	Orders      []models.Order
}

func OptimizeRoute(warehouse models.Warehouse, orders []models.Order) []models.Order {
	if len(orders) == 0 {
		return orders
	}
	visited := make([]bool, len(orders))
	route := make([]models.Order, 0, len(orders))
	currLat, currLon := warehouse.Lat, warehouse.Lon

	for range orders {
		minDist := math.MaxFloat64
		nextIdx := -1
		for i, o := range orders {
			if visited[i] {
				continue
			}
			d := haversine(currLat, currLon, o.Lat, o.Lon)
			if d < minDist {
				minDist = d
				nextIdx = i
			}
		}
		if nextIdx == -1 {
			break
		}
		visited[nextIdx] = true
		route = append(route, orders[nextIdx])
		currLat, currLon = orders[nextIdx].Lat, orders[nextIdx].Lon
	}
	return route
}

func calculateRouteMetrics(warehouse models.Warehouse, orders []models.Order) (float64, int) {
	if len(orders) == 0 {
		return 0, 0
	}

	totalDistance := 0.0
	currLat, currLon := warehouse.Lat, warehouse.Lon

	totalDistance += haversine(currLat, currLon, orders[0].Lat, orders[0].Lon)
	
	for i := 1; i < len(orders); i++ {
		totalDistance += haversine(orders[i-1].Lat, orders[i-1].Lon, orders[i].Lat, orders[i].Lon)
	}
	
	if len(orders) > 0 {
		lastOrder := orders[len(orders)-1]
		totalDistance += haversine(lastOrder.Lat, lastOrder.Lon, warehouse.Lat, warehouse.Lon)
	}

	totalTime := int(totalDistance / AvgSpeedKMPerMin)
	return totalDistance, totalTime
}

func AllocateOrdersForWarehouse(warehouse models.Warehouse) error {
	agents, err := fetchCheckedInAgents(warehouse.ID)
	if err != nil {
		return err
	}

	if len(agents) == 0 {
		return nil 
	}

	orders, err := fetchUnassignedOrders(warehouse.ID)
	if err != nil {
		return err
	}

	if len(orders) == 0 {
		return nil 
	}

	
	agentCapacities := make([]AgentCapacity, len(agents))
	for i, agent := range agents {
		agentCapacities[i] = AgentCapacity{
			Agent:       agent,
			UsedKM:      0,
			UsedMinutes: 0,
			Orders:      make([]models.Order, 0),
		}
	}

	for i := range orders {
		dist := haversine(warehouse.Lat, warehouse.Lon, orders[i].Lat, orders[i].Lon)
		orders[i].DistanceKM = dist
		orders[i].EstimatedTimeMinutes = int(dist / AvgSpeedKMPerMin)
	}

	sortOrdersByDistance(&orders)

	unassignedOrders := make([]int, 0)
	orderIndex := 0
	
	for orderIndex < len(orders) {
		assignedInThisRound := false
		
		for agentIdx := range agentCapacities {
			if orderIndex >= len(orders) {
				break
			}
			
			order := orders[orderIndex]
			if order.Assigned {
				orderIndex++
				continue
			}
			
			testOrders := append(agentCapacities[agentIdx].Orders, order)
			totalDist, totalTime := calculateRouteMetrics(warehouse, testOrders)
			
			if totalDist <= MaxDistanceKM && totalTime <= MaxTimeMinutes {
				agentCapacities[agentIdx].Orders = testOrders
				agentCapacities[agentIdx].UsedKM = totalDist
				agentCapacities[agentIdx].UsedMinutes = totalTime
				orders[orderIndex].Assigned = true
				assignedInThisRound = true
				orderIndex++
			} else {
				
				continue
			}
		}
		
		if !assignedInThisRound {
			if orderIndex < len(orders) && !orders[orderIndex].Assigned {
				orderIndex++
			}
		}
	}

	for _, agentCap := range agentCapacities {
		if len(agentCap.Orders) > 0 {
			optimizedOrders := OptimizeRoute(warehouse, agentCap.Orders)
			err := saveAssignments(agentCap.Agent.ID, optimizedOrders)
			if err != nil {
				return err
			}
		}
	}

	for i := range orders {
		if !orders[i].Assigned {
			unassignedOrders = append(unassignedOrders, orders[i].ID)
		}
	}

	if len(unassignedOrders) > 0 {
		err := postponeOrdersToNextDay(unassignedOrders)
		if err != nil {
			return err
		}
	}

	return nil
}

//Balanced load distribution
func AllocateOrdersBalanced(warehouse models.Warehouse) error {
	agents, err := fetchCheckedInAgents(warehouse.ID)
	if err != nil {
		return err
	}

	if len(agents) == 0 {
		return nil
	}

	orders, err := fetchUnassignedOrders(warehouse.ID)
	if err != nil {
		return err
	}

	if len(orders) == 0 {
		return nil
	}

	for i := range orders {
		dist := haversine(warehouse.Lat, warehouse.Lon, orders[i].Lat, orders[i].Lon)
		orders[i].DistanceKM = dist
		orders[i].EstimatedTimeMinutes = int(dist / AvgSpeedKMPerMin)
	}

	sortOrdersByDistance(&orders)

	agentCapacities := make([]AgentCapacity, len(agents))
	for i, agent := range agents {
		agentCapacities[i] = AgentCapacity{
			Agent:       agent,
			UsedKM:      0,
			UsedMinutes: 0,
			Orders:      make([]models.Order, 0),
		}
	}

	unassignedOrders := make([]int, 0)
	for _, order := range orders {
		assigned := false
		
		sort.Slice(agentCapacities, func(i, j int) bool {
			return len(agentCapacities[i].Orders) < len(agentCapacities[j].Orders)
		})
		
		for agentIdx := range agentCapacities {
			testOrders := append(agentCapacities[agentIdx].Orders, order)
			totalDist, totalTime := calculateRouteMetrics(warehouse, testOrders)
			
			if totalDist <= MaxDistanceKM && totalTime <= MaxTimeMinutes {
				agentCapacities[agentIdx].Orders = testOrders
				agentCapacities[agentIdx].UsedKM = totalDist
				agentCapacities[agentIdx].UsedMinutes = totalTime
				assigned = true
				break
			}
		}
		
		if !assigned {
			unassignedOrders = append(unassignedOrders, order.ID)
		}
	}

	for _, agentCap := range agentCapacities {
		if len(agentCap.Orders) > 0 {
			optimizedOrders := OptimizeRoute(warehouse, agentCap.Orders)
			err := saveAssignments(agentCap.Agent.ID, optimizedOrders)
			if err != nil {
				return err
			}
		}
	}

	if len(unassignedOrders) > 0 {
		err := postponeOrdersToNextDay(unassignedOrders)
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
        INSERT INTO agent_assignments (agent_id, order_id, assigned_on, distance_km, estimated_time_minutes, sequence_number)
        VALUES ($1, $2, CURRENT_DATE, $3, $4, $5)
    `)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	updateStmt, err := tx.Prepare(`UPDATE orders SET assigned = true, agent_id = $2 WHERE id = $1`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer updateStmt.Close()

	for seq, order := range orders {
		_, err := stmt.Exec(agentID, order.ID, order.DistanceKM, order.EstimatedTimeMinutes, seq+1)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = updateStmt.Exec(order.ID, agentID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func postponeOrdersToNextDay(orderIDs []int) error {
	if len(orderIDs) == 0 {
		return nil
	}
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE orders SET scheduled_for = $1 WHERE id = $2`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	nextDay := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	for _, id := range orderIDs {
		_, err := stmt.Exec(nextDay, id)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}