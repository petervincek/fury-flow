package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/petervincek/fury-flow/app"
	_ "github.com/petervincek/fury-flow/docs"
	"github.com/petervincek/fury-flow/internal/config"
	"github.com/petervincek/fury-flow/internal/logging"
	"go.uber.org/zap"
)

var logger = logging.GetLogger()

func main() {
	fmt.Printf("Fury flow\n")
	config.LoadEnvVariables(".env")
	app := app.NewApp()

	// start the application in separate go routine
	go func() {
		if err := app.StartApp(); err != nil {
			logger.Error("Error while starting the application", zap.Error(err))
		}
	}()

	// listen for interrupt signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM) // listen for interupt and termination
	<-sigs                                               // block until received
	// call the graceful shutdown
	app.ShutdownApp(30 * time.Second)
}
