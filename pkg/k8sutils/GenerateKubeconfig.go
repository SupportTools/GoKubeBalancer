package k8sutils

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/supporttools/GoKubeBalancer/pkg/config"
)

// GenerateKubeconfig creates a kubeconfig for a specified cluster and returns it as a string.
func GenerateKubeconfig(ctx context.Context, clusterID string) (string, error) {
	logger.Info("Generating kubeconfig...")

	// Construct the request URL for generating kubeconfig.
	url := fmt.Sprintf("%s/v3/clusters/%s?action=generateKubeconfig", config.CFG.RancherAPI, clusterID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		logger.Errorf("Failed to create HTTP request: %v", err)
		return "", fmt.Errorf("create HTTP request: %w", err)
	}

	// Encode the authentication credentials and set the request headers.
	authHeader := base64.StdEncoding.EncodeToString([]byte(config.CFG.RancherKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+authHeader)

	// Configure the client to ignore certificate validation
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second, // 10 seconds timeout
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to send HTTP request: %v", err)
		return "", fmt.Errorf("send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		logger.Errorf("Failed to generate kubeconfig, status code: %d", resp.StatusCode)
		return "", fmt.Errorf("generate kubeconfig, status code: %d", resp.StatusCode)
	}

	// Decode the response body to extract the kubeconfig data.
	var response struct {
		Config string `json:"config"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Errorf("Failed to decode JSON response: %v", err)
		return "", fmt.Errorf("decode JSON response: %w", err)
	}

	logger.Info("Kubeconfig data retrieved successfully.")
	return response.Config, nil
}
