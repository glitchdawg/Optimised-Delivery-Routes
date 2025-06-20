package models

import "time"

// type Order struct {
//     ID                   int
//     WarehouseID          int
//     Lat                  float64
//     Lon                  float64
//     DeliveryAddress      string
//     ScheduledFor         time.Time
//     Assigned             bool
//     DistanceKM           float64 // computed
//     EstimatedTimeMinutes int     // computed
//     AgentID              int     // assigned agent ID
// }
type Order struct {
    ID                   int       `json:"id"`
    WarehouseID          int       `json:"warehouse_id"`
    Lat                  float64   `json:"lat"`
    Lon                  float64   `json:"lon"`
    DeliveryAddress      string    `json:"delivery_address"`
    ScheduledFor         time.Time `json:"scheduled_for"`
    Assigned             bool      `json:"assigned"`
    DistanceKM           float64   `json:"distance_km"`
    EstimatedTimeMinutes int       `json:"estimated_time_minutes"`
    AgentID              *int       `json:"agent_id"`
}