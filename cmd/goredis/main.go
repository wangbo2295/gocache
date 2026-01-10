package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wangbo/gocache/config"
	"github.com/wangbo/gocache/database"
	"github.com/wangbo/gocache/logger"
	"github.com/wangbo/gocache/persistence/aof"
	"github.com/wangbo/gocache/server"
)

var (
	configFile = flag.String("c", "", "Configuration file path")
)

func main() {
	flag.Parse()

	// Load configuration
	if *configFile != "" {
		if err := config.Load(*configFile); err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			os.Exit(1)
		}
	}

	// Initialize logger
	logger.SetLevel(config.Config.LogLevel)
	if config.Config.LogFile != "" {
		if err := logger.SetFile(config.Config.LogFile); err != nil {
			fmt.Printf("Failed to set log file: %v\n", err)
			os.Exit(1)
		}
	}

	logger.Info("Starting GoCache server...")
	logger.Info("Version: 1.0.0-MVP")
	logger.Info("Binding to %s:%d", config.Config.Bind, config.Config.Port)

	// Create database
	db := database.MakeDB()

	// Create AOF handler if enabled
	var aofHandler *aof.AOFHandler
	var err error

	if config.Config.AppendOnly {
		logger.Info("AOF persistence enabled: %s", config.Config.AppendFilename)
		aofHandler, err = aof.MakeAOFHandler(config.Config.AppendFilename, db)
		if err != nil {
			logger.Error("Failed to initialize AOF: %v", err)
			os.Exit(1)
		}
		defer aofHandler.Close()
	}

	// Create handler
	var handler *server.Handler
	if aofHandler != nil {
		handler = server.MakeHandlerWithAOF(db, aofHandler)
	} else {
		handler = server.MakeHandler(db)
	}

	// Create and start server
	srv := server.MakeServer(config.Config, handler)

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down server...")
		srv.Stop()
		if aofHandler != nil {
			aofHandler.Close()
		}
		logger.Close()
		os.Exit(0)
	}()

	// Start server
	if err := srv.Start(); err != nil {
		logger.Error("Server error: %v", err)
		os.Exit(1)
	}
}
