package main

import (
	"context"

	ui "github.com/kormiltsev/item-keeper/internal/commandui"
)

// import client "github.com/kormiltsev/item-keeper/internal/client"

// logger "github.com/kormiltsev/item-keeper/internal/logger"
// "go.uber.org/zap"

func main() {
	//redirect logger
	// blog := logger.NewLog("./configs/logger.json")
	// defer blog.Logger.Sync()
	// undo := zap.RedirectStdLog(blog.Logger)
	// defer undo()

	// chan close signal
	var chclose = make(chan struct{})

	// start app

	// start ui
	ui.StartTui(context.Background(), chclose)

	<-chclose
}
