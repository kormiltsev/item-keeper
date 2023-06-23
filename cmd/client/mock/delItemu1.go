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

	var i int64 = 5
	arr := make([]int64, 0, 10)
	for i < 6 {
		arr = append(arr, i)
		log.Println("try to delete this files = ", arr)
		list, err := app.DeleteItems(ctx, arr)
		if err != nil {
			log.Println("FAIL: delete item:", err)
			return
		}

		if len(list) != 0 {
			log.Println("this item ids was not found:", list)
		}
		i++
	}

	time.Sleep(1 * time.Second)

}
