package health

import (
	"encoding/json"
	"net/http"

	"github.com/supporttools/GoKubeBalancer/pkg/logging"
)

// VersionInfo represents the structure of version information.
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
}

var logger = logging.SetupLogging()

// Variables to be set by the linker during the build process
var (
	Version   = "unknown"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

var versionInfo = VersionInfo{
	Version:   Version,
	GitCommit: GitCommit,
	BuildTime: BuildTime,
}

func HealthzHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func ReadyzHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func VersionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(versionInfo)
	})
}
