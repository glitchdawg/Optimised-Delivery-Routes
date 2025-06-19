package models

import "time"

type Assignment struct {
    ID             int       `json:"id"`
    AgentID        int       `json:"agent_id"`
    OrderID        int       `json:"order_id"`
    AssignedOn     time.Time `json:"assigned_on"`
    DistanceKM     float64   `json:"distance_km"`
    EstimatedMins  int       `json:"estimated_time_minutes"`
}
