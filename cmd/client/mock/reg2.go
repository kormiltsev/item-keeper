package main

import (
	"context"
	"log"

	app "github.com/kormiltsev/item-keeper/internal/app"
)

const (
	login2    = "login2"
	password2 = "password2"
)

func main() {

	ctx := context.Background()

	// create new user
	err := app.RegUser(ctx, login2, password2)
	if err != nil {
		log.Println("FAIL create user:", err)
		return
	}

	// create user exists
	// err = app.RegUser(ctx, login2, "lll")
	// if err == nil {
	// 	log.Println(`FAIL create user: expect error "user exists", but resieved nil`)
	// 	return
	// }
}
