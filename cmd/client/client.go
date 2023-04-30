package main

import client "github.com/kormiltsev/item-keeper/internal/client"

func main() {
	client.RunTestClient(":3333")
}
