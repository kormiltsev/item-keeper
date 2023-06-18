package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	ui "github.com/kormiltsev/item-keeper/internal/commandui"
	logger "github.com/kormiltsev/item-keeper/internal/logger"
	"go.uber.org/zap"
)

// import client "github.com/kormiltsev/item-keeper/internal/client"

// logger "github.com/kormiltsev/item-keeper/internal/logger"
// "go.uber.org/zap"

func main() {
	//redirect logger
	blog := logger.NewLog("./configs/loggerClient.json")
	defer blog.Logger.Sync()
	undo := zap.RedirectStdLog(blog.Logger)
	defer undo()

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// chan close signal
	var chclose = make(chan struct{})

	// start app
	// start ui
	ui.StartTui(context.Background(), chclose, stopSignal)

	<-chclose
}
