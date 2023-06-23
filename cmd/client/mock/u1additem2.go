package main

import (
	"context"
	"log"
	"time"

	app "github.com/kormiltsev/item-keeper/internal/app"
	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
)

const (
	login    = "login1"
	password = "password1"
)

func main() {

	ctx := context.Background()

	// valid authorization
	err := app.AuthUser(ctx, login, password)
	if err != nil {
		log.Println("FAIL authorization user:", err)
		return
	}

	// invalid authorization
	// err = app.AuthUser(ctx, login, "wrong")
	// if err == nil {
	// 	log.Println(`FAIL create user: expect error "wrong password", but resieved nil`)
	// 	return
	// }

	// add item 2 to user 1
	err = app.AddNewItem(ctx, &appstorage.Item{
		Parameters:    []appstorage.Parameter{{Name: "Name", Value: "Name Of Item 2"}, {Name: "Color", Value: "Green or red or what ever"}, {Name: "Has file", Value: "2 (txt and jpeg)"}},
		UploadAddress: []string{"./data/sourceClient/test.txt", "./data/sourceClient/Jocker.jpeg"},
	})
	if err != nil {
		log.Println("FAIL add item 2:", err)
		return
	}

	time.Sleep(1 * time.Second)

}
