package main

import (
	"fmt"
	"log"
	"os"

	"github.com/somnathbm/horus/internal/config"
)

func main() {
	// load app config
	appConfigPath := os.Getenv("APP_CONFIG_PATH")
	if appConfigPath == "" {
		appConfigPath = "configs/application.yaml"
	}
	appConfig, err := config.Load(appConfigPath)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println(appConfig)
}
