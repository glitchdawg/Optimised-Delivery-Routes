package models

import "time"

type Payout struct {
    ID            int       `json:"id"`
    AgentID       int       `json:"agent_id"`
    Date          time.Time `json:"date"`
    TotalOrders   int       `json:"total_orders"`
    TotalDistance float64   `json:"total_distance"`
    TotalPay      float64   `json:"total_pay"`
}
