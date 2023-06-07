package main

import (
	"context"
	"log"
	"time"

	app "github.com/kormiltsev/item-keeper/internal/app"
)

func main() {

	ctx := context.Background()

	// valid authorization
	err := app.AuthUser(ctx, "login1", "password1")
	if err != nil {
		log.Println("FAIL authorization user:", err)
		return
	}

	time.Sleep(1 * time.Second)

}
