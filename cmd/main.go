package main

import (
    "fmt"
    "log"
    "github.com/glitchdawg/Optimised-Delivery-Routes/internal/db"
)

func main() {
    err := db.InitDB()
    if err != nil {
        log.Fatalf("Failed to connect to DB: %v", err)
    }

    fmt.Println("DB connected successfully ✅")

}
