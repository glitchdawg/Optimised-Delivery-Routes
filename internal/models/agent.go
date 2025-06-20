package models

import "time"

type Agent struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	WarehouseID *int      `json:"warehouse_id"`
	CheckedInAt time.Time `json:"checked_in_at"`
}
