package k8sutils

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/supporttools/GoKubeBalancer/pkg/config"
)

// GetClusterID fetches the cluster ID for a given cluster name from Rancher.
func GetClusterID() (string, error) {
	log.Infof("Requesting cluster ID for cluster named '%s' from Rancher.", config.CFG.RancherCluster)

	url := fmt.Sprintf("%s/v3/clusters?name=%s", config.CFG.RancherAPI, config.CFG.RancherCluster)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Failed to create HTTP request for Rancher API: %v", err)
		return "", fmt.Errorf("create HTTP request: %w", err)
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(config.CFG.RancherKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+authHeader)

	// Create a client with a custom transport to ignore SSL certificate errors
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.CFG.InsecureSkipVerify,
			},
		},
		Timeout: 10 * time.Second, // 10 seconds timeout
	}

	log.Debugf("Sending request to Rancher API at URL: %s", url)
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Failed to send HTTP request to Rancher API: %v", err)
		return "", fmt.Errorf("send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Rancher API responded with status code %d", resp.StatusCode)
		return "", fmt.Errorf("get cluster ID, status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Errorf("Failed to decode JSON response from Rancher API: %v", err)
		return "", fmt.Errorf("decode JSON response: %w", err)
	}

	if len(result.Data) == 0 {
		log.Info("No cluster ID found for specified cluster name.")
		return "", fmt.Errorf("no cluster ID found for cluster name: %s", config.CFG.RancherCluster)
	}

	clusterID := result.Data[0].ID
	log.Infof("Successfully retrieved cluster ID: %s", clusterID)
	return clusterID, nil
}
