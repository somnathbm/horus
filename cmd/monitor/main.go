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
	"github.com/somnathbm/horus/internal/engine"
	"github.com/somnathbm/horus/internal/plugin"
	"github.com/somnathbm/horus/internal/registry"
	"github.com/somnathbm/horus/internal/target"
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
	dscvryRslt, dscvryErr := dscvryManager.Discover(ctx)
	if dscvryErr != nil {
		fmt.Printf("discovery: %v", dscvryErr)
	}

	// 5. instantiate engine with the registry
	reg := registry.New()
	reg.Register(target.TargetKubernetesJob, plugin.KubernetesJobPlugin{})
	// TODO: register other plugins if any

	// 6. instantiate engine and evaluate discovered targets
	eng := engine.New(reg)
	// iterate over discovered targets and evaluate them
	for _, t := range dscvryRslt.Targets {
		status, err := eng.Evaluate(ctx, t)
		if err != nil {
			fmt.Printf("target %v: %v\n", t.Name, err)
		} else {
			fmt.Printf("target %v: %v\n", t.Name, status)
		}
	}

}
