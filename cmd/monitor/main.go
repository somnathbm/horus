package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/somnathbm/horus/internal/config"
	"github.com/somnathbm/horus/internal/discovery"
)

func main() {
	// 1. load app config
	appConfigPath := os.Getenv("APP_CONFIG_PATH")
	if appConfigPath == "" {
		appConfigPath = "configs/application.yaml"
	}
	appConfig, err := config.Load(appConfigPath)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// 2. start discovery
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	targets, err := discovery.Start(appConfig.Discovery, ctx)
	if err != nil {
		fmt.Printf("%v", err)
	}

	fmt.Println(targets)
}
