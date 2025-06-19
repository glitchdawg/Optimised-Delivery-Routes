package models

import "time"

type Order struct {
    ID                   int
    WarehouseID          int
    Lat                  float64
    Lon                  float64
    DeliveryAddress      string
    ScheduledFor         time.Time
    Assigned             bool
    DistanceKM           float64 // computed
    EstimatedTimeMinutes int     // computed
}
