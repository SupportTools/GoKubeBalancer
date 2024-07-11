package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// AppConfig structure for environment-based configurations.
type AppConfig struct {
	Debug              bool          `json:"debug"`
	MetricsPort        int           `json:"metricsPort"`
	InsecureSkipVerify bool          `json:"insecureSkipVerify"`
	FrontendHttpPort   int           `json:"frontendHttpPort"`
	FrontendHttpsPort  int           `json:"frontendHttpsPort"`
	BackendHttpPort    int           `json:"backendHttpPort"`
	BackendHttpsPort   int           `json:"backendHttpsPort"`
	NodeSelector       string        `json:"nodeSelector"`
	NewNodeThreshold   time.Duration `json:"newNodeThreshold"`
	RescanInterval     time.Duration `json:"rescanInterval"`
	RancherAPI         string        `json:"rancherAPI"`
	RancherKey         string        `json:"rancherKey"`
	RancherCluster     string        `json:"rancherCluster"`
}

var CFG AppConfig

// LoadConfiguration loads configuration from environment variables.
func LoadConfiguration() error {
	CFG.Debug = parseEnvBool("DEBUG", false)                                                   // Assuming false as the default value
	CFG.MetricsPort = parseEnvInt("METRICS_PORT", 9099)                                        // Assuming 9099 as the default port
	CFG.InsecureSkipVerify = parseEnvBool("INSECURE_SKIP_VERIFY", false)                       // Assuming false as the default value
	CFG.FrontendHttpPort = parseEnvInt("FRONTEND_HTTP_PORT", 80)                               // Assuming 80 as the default port
	CFG.FrontendHttpsPort = parseEnvInt("FRONTEND_HTTPS_PORT", 443)                            // Assuming 443 as the default port
	CFG.BackendHttpPort = parseEnvInt("BACKEND_HTTP_PORT", 80)                                 // Assuming 80 as the default port
	CFG.BackendHttpsPort = parseEnvInt("BACKEND_HTTPS_PORT", 443)                              // Assuming 443 as the default port
	CFG.NodeSelector = getEnvOrDefault("NODE_SELECTOR", "node-role.kubernetes.io/worker=true") // Node Selector for selecting backend members
	CFG.RancherAPI = getEnvOrDefault("RANCHER_API", "https://rancher.example.com")             // Rancher API URL
	CFG.RancherKey = getEnvOrDefault("RANCHER_KEY", "")                                        // Rancher API Key access:secret
	CFG.RancherCluster = getEnvOrDefault("RANCHER_CLUSTER", "local")                           // Rancher cluster name
	CFG.NewNodeThreshold = time.Duration(parseEnvInt("NEW_NODE_THRESHOLD", 15)) * time.Minute  // Assuming 60 minutes as the default threshold, this gives the node time to warm up before being considered healthy
	CFG.RescanInterval = time.Duration(parseEnvInt("RESCAN_INTERVAL", 5)) * time.Second        // Time interval for rescanning the backend members

	// Validate the configuration
	if err := ValidateConfiguration(&CFG); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
		os.Exit(1)
	}
	log.Printf("Configuration validated")
	return nil
}

func getBackendMembers() []string {
	backendMembers := os.Getenv("BACKEND_MEMBERS")
	if backendMembers == "" {
		log.Println("No backend members configured")
		return nil
	}
	return []string{backendMembers}
	// TODO: Parse backend members from a file or a list of IPs
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Error parsing %s as int: %v. Using default value: %d", key, err, defaultValue)
		return defaultValue
	}
	return intValue
}

func parseEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Error parsing %s as bool: %v. Using default value: %t", key, err, defaultValue)
		return defaultValue
	}
	return boolValue
}

func validatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port number %d; must be between 1 and 65535", port)
	}
	return nil
}

func validateNonEmpty(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", field)
	}
	return nil
}

func ValidateConfiguration(cfg *AppConfig) error {
	if err := validatePort(cfg.MetricsPort); err != nil {
		return err
	}
	if err := validateNonEmpty("nodeSelector", cfg.NodeSelector); err != nil {
		return err
	}
	if err := validateNonEmpty("rancherAPI", cfg.RancherAPI); err != nil {
		return err
	}
	if err := validateNonEmpty("rancherKey", cfg.RancherKey); err != nil {
		return err
	}
	if err := validateNonEmpty("rancherCluster", cfg.RancherCluster); err != nil {
		return err
	}
	if err := validateNonEmpty("frontendHttpPort", strconv.Itoa(cfg.FrontendHttpPort)); err != nil {
		return err
	}
	if err := validateNonEmpty("frontendHttpsPort", strconv.Itoa(cfg.FrontendHttpsPort)); err != nil {
		return err
	}
	if err := validateNonEmpty("backendHttpPort", strconv.Itoa(cfg.BackendHttpPort)); err != nil {
		return err
	}
	if err := validateNonEmpty("backendHttpsPort", strconv.Itoa(cfg.BackendHttpsPort)); err != nil {
		return err
	}
	return nil
}
