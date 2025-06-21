package services

import (
	"time"

	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/db"
	"github.com/glitchdawg/Optimised-Delivery-Routes/internal/models"
)

func CalculatePayoutForAgent(agentID int) (*models.Payout, error) {
    row := db.DB.QueryRow(`
        SELECT COUNT(*), COALESCE(SUM(distance_km), 0)
        FROM agent_assignments
        WHERE agent_id = $1 AND assigned_on = CURRENT_DATE
    `, agentID)

    var totalOrders int
    var totalDistance float64
    if err := row.Scan(&totalOrders, &totalDistance); err != nil {
        return nil, err
    }

    var pay float64
    switch {
    case totalOrders >= 50:
        pay = float64(totalOrders) * 42
    case totalOrders >= 25:
        pay = float64(totalOrders) * 35
    case totalOrders > 0:
        pay = 500
    default:
        pay = 0
    }

    payout := &models.Payout{
        AgentID:       agentID,
        Date:          time.Now(),
        TotalOrders:   totalOrders,
        TotalDistance: totalDistance,
        TotalPay:      pay,
    }

    return payout, nil
}