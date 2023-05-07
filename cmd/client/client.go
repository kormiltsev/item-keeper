package main

// import client "github.com/kormiltsev/item-keeper/internal/client"
import (
	"context"

	logger "github.com/kormiltsev/item-keeper/internal/logger"
	tui "github.com/kormiltsev/item-keeper/internal/tui"
	"go.uber.org/zap"
)

func main() {
	//redirect logger
	blog := logger.NewLog("./configs/logger.json")
	defer blog.Logger.Sync()
	undo := zap.RedirectStdLog(blog.Logger)
	defer undo()

	ctx := context.Background()
	tui.StartTui(ctx)
	// client.RunTestClient(":3333")
}
