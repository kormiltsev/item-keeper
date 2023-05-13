package main

import (
	app "github.com/kormiltsev/item-keeper/internal/app"
)

func main() {
	// all storages are empty
	// reg new user
	app.RegUser("NewUserLogin", "NewUserPassword")
	// add items for user1
	app.AddNewItem()
	// search by parameters local on client's side
	app.SearchItemByParameters()

	// reg new user2
	app.RegUser("NewUser2Login", "NewUser2Password")
	// local data should empty for user1 now
	// add items for user2
	app.AddNewItem()

	// login user1
	app.AuthUser("NewUserLogin", "NewUserPassword")
	// app.UpdateDataFromServer() // AuthUser() run this if success

	// search by parameters local on client's side
	app.SearchItemByParameters()

	// login user2
	app.AuthUser("NewUser2Login", "NewUser2Password")
	// app.UpdateDataFromServer() // AuthUser() run this if success

	// search by parameters local on client's side
	app.SearchItemByParameters()
}
