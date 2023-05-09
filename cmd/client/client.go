package main

// import client "github.com/kormiltsev/item-keeper/internal/client"
import (
	"context"
	"log"

	app "github.com/kormiltsev/item-keeper/internal/app"
	// logger "github.com/kormiltsev/item-keeper/internal/logger"
	// "go.uber.org/zap"
)

func main() {
	//redirect logger
	// blog := logger.NewLog("./configs/logger.json")
	// defer blog.Logger.Sync()
	// undo := zap.RedirectStdLog(blog.Logger)
	// defer undo()

	ctx := context.Background()
	err := app.RegUser(ctx, "Login1", "Password1")
	if err != nil {
		log.Println("FAIL create user")
		return
	}
	log.Println("done")
	// previous client
	// tui.StartTui(ctx)
	// client.RunTestClient(":3333")
}
