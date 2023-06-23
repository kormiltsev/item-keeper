package main

import (
	"context"
	"log"

	app "github.com/kormiltsev/item-keeper/internal/app"
)

const (
	login    = "login1"
	password = "password1"
)

func main() {

	ctx := context.Background()

	// create new user
	err := app.RegUser(ctx, login, password)
	if err != nil {
		log.Println("FAIL create user:", err)
		return
	}

	// create user exists
	// err = app.RegUser(ctx, login, "lll")
	// if err == nil {
	// 	log.Println(`FAIL create user: expect error "user exists", but resieved nil`)
	// 	return
	// }
}
