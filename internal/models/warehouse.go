package models

type Warehouse struct {
    ID   int     `json:"id"`
    Name string  `json:"name"`
    Lat  float64 `json:"lat"`
    Lon  float64 `json:"lon"`
}
