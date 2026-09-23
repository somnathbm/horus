package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/somnathbm/horus/internal/config"
	"github.com/somnathbm/horus/internal/discovery"
	"github.com/somnathbm/horus/internal/discovery/local"
)

func main() {
	// 1. load app config
	appConfigPath := os.Getenv("APP_CONFIG_PATH")
	if appConfigPath == "" {
		appConfigPath = "configs/application.yaml"
	}
	appConfig, configErr := config.Load(appConfigPath)
	if configErr != nil {
		log.Fatalf("Error: %v", configErr)
	}

	// 2. create parent context
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 3. construct config -> discoverer
	var discoverers []discovery.Discoverer
	discoveryConfig := appConfig.Discovery

	// local config
	for _, localConfig := range discoveryConfig.Local {
		discoverers = append(discoverers, local.New(localConfig))
	}

	// other config if any (AWS, kubernetes etc.)

	// 4. instantiate manager
	dscvryManager := discovery.New(discoverers...)
	dscvryResult, dscvryErr := dscvryManager.Discover(ctx)
	if dscvryErr != nil {
		fmt.Printf("discovery: %v", dscvryErr)
	}
	fmt.Println("@@@@@")
	fmt.Println(dscvryResult)

}
