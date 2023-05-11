package main

import app "github.com/kormiltsev/item-keeper/internal/app"

func main() {
	// app.Encode()

	app.RegUser("NewUserLogin", "NewUserPassword")
	app.AddNewItem()
	app.SearchItemByParameters()
}
