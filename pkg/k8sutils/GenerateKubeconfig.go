package k8sutils

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/supporttools/GoKubeBalancer/pkg/config"
)

// GenerateKubeconfig creates a kubeconfig for a specified cluster and returns it as a string.
func GenerateKubeconfig(ctx context.Context, clusterID string) (string, error) {
	logrus.Info("Generating kubeconfig...")

	url := fmt.Sprintf("%s/v3/clusters/%s?action=generateKubeconfig", config.CFG.RancherAPI, clusterID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		logrus.Errorf("Failed to create HTTP request: %v", err)
		return "", fmt.Errorf("create HTTP request: %w", err)
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(config.CFG.RancherKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+authHeader)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	var response *http.Response
	for retries := 0; retries < 5; retries++ {
		response, err = client.Do(req)
		if err != nil {
			logrus.Errorf("Failed to send HTTP request (attempt %d): %v", retries+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if response.StatusCode == http.StatusOK {
			break
		}
		logrus.Errorf("Failed to generate kubeconfig, status code: %d", response.StatusCode)
		time.Sleep(2 * time.Second)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("generate kubeconfig, status code: %d", response.StatusCode)
	}
	defer response.Body.Close()

	var respBody struct {
		Config string `json:"config"`
	}
	if err := json.NewDecoder(response.Body).Decode(&respBody); err != nil {
		logrus.Errorf("Failed to decode JSON response: %v", err)
		return "", fmt.Errorf("decode JSON response: %w", err)
	}

	logrus.Info("Kubeconfig data retrieved successfully.")
	return respBody.Config, nil
}
