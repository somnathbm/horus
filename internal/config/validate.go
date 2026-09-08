package config

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var (
	ErrNoDiscoverySources = errors.New("no discovery sources found")
	ErrNoSourceName       = errors.New("source name is required")
	ErrInvalidFilePath    = errors.New("valid absolute file path is required")
	ErrInvalidClusterURL  = errors.New("valid URL is required")
	ErrInvalidAccountID   = errors.New("valid 12-digit accountID is required")
	ErrInvalidAWSRegion   = errors.New("valid aws region is required")
)

var reAccountID = regexp.MustCompile(`^\d{12}$`)
var allowedSchemes = []string{"http", "https"}

// Validate semantics for the config
func validateConfig(appConfig AppConfig) (AppConfig, error) {
	// if no discovery sources are configured
	if len(appConfig.Discovery.Local) == 0 && len(appConfig.Discovery.Kubernetes) == 0 && len(appConfig.Discovery.AWS) == 0 {
		return AppConfig{}, fmt.Errorf("app.discovery: %w", ErrNoDiscoverySources)
	}

	sources := appConfig.Discovery

	for i, local := range sources.Local {
		sources.Local[i].Name = strings.TrimSpace(local.Name)
		sources.Local[i].Path = strings.TrimSpace(local.Path)
		if sources.Local[i].Name == "" {
			return AppConfig{}, fmt.Errorf("discovery.local[%v]: %w", i, ErrNoSourceName)
		}
		if sources.Local[i].Path == "" || !filepath.IsAbs(sources.Local[i].Path) {
			return AppConfig{}, fmt.Errorf("discovery.local[%v]: %w", i, ErrInvalidFilePath)
		}
	}

	// kubernetes
	for i, k8s := range sources.Kubernetes {
		sources.Kubernetes[i].Name = strings.TrimSpace(k8s.Name)
		sources.Kubernetes[i].ClusterURL = strings.TrimSpace(k8s.ClusterURL)
		urlRes, err := url.Parse(sources.Kubernetes[i].ClusterURL)

		if err != nil {
			return AppConfig{}, fmt.Errorf("discovery.kubernetes[%v]: %w", i, ErrInvalidClusterURL)
		}
		if sources.Kubernetes[i].Name == "" {
			return AppConfig{}, fmt.Errorf("discovery.kubernetes[%v]: %w", i, ErrNoSourceName)
		}
		if sources.Kubernetes[i].ClusterURL == "" || urlRes.Scheme == "" || !slices.Contains(allowedSchemes, urlRes.Scheme) || urlRes.Host == "" {
			return AppConfig{}, fmt.Errorf("discovery.kubernetes[%v]: %w", i, ErrInvalidClusterURL)
		}
	}

	// AWS
	for i, aws := range sources.AWS {
		sources.AWS[i].Name = strings.TrimSpace(aws.Name)
		sources.AWS[i].AccountID = strings.TrimSpace(aws.AccountID)
		sources.AWS[i].Region = strings.TrimSpace(aws.Region)

		if sources.AWS[i].Name == "" {
			return AppConfig{}, fmt.Errorf("discovery.aws[%v]: %w", i, ErrNoSourceName)
		}
		if sources.AWS[i].AccountID == "" || len(sources.AWS[i].AccountID) != 12 || !reAccountID.MatchString(sources.AWS[i].AccountID) {
			return AppConfig{}, fmt.Errorf("discovery.aws[%v]: %w", i, ErrInvalidAccountID)
		}
		if sources.AWS[i].Region == "" {
			return AppConfig{}, fmt.Errorf("discovery.aws[%v]: %w", i, ErrInvalidAWSRegion)
		}
	}

	return appConfig, nil
}
