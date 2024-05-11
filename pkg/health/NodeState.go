package health

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

var nodeStates sync.Map // A thread-safe map to store node states

// NodeState holds the recovery state of a node
type NodeState struct {
	NodeName      string                        `json:"nodeName"`
	OverallStatus string                        `json:"overallStatus"`
	Timestamp     string                        `json:"timestamp"`
	RecoverySteps map[string]RecoveryStepDetail `json:"recoverySteps"`
}

// RecoveryStepDetail holds details for each recovery step
type RecoveryStepDetail struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// RegisterNodeState updates the state of a node in the nodeStates map
func RegisterNodeState(nodeName, step, status string, overallStatus string) {
	now := time.Now().Format(time.RFC3339)
	newStepDetail := RecoveryStepDetail{
		Status:    status,
		Timestamp: now,
	}

	existingValue, _ := nodeStates.LoadOrStore(nodeName, NodeState{
		NodeName:      nodeName,
		OverallStatus: overallStatus,
		Timestamp:     now,
		RecoverySteps: map[string]RecoveryStepDetail{step: newStepDetail},
	})

	if nodeState, ok := existingValue.(NodeState); ok {
		if nodeState.RecoverySteps == nil {
			nodeState.RecoverySteps = make(map[string]RecoveryStepDetail)
		}
		nodeState.RecoverySteps[step] = newStepDetail
		nodeState.Timestamp = now // update the timestamp to the latest update
		nodeState.OverallStatus = overallStatus
		nodeStates.Store(nodeName, nodeState)
	}
}

// NodeStatesHandler returns the current state of all nodes as JSON
func NodeStatesHandler(w http.ResponseWriter, r *http.Request) {
	var allStates []NodeState
	nodeStates.Range(func(_, value interface{}) bool {
		if state, ok := value.(NodeState); ok {
			allStates = append(allStates, state)
		}
		return true
	})

	// Sort states by nodeName
	sort.Slice(allStates, func(i, j int) bool {
		return allStates[i].NodeName < allStates[j].NodeName
	})

	jsonData, err := json.Marshal(allStates)
	if err != nil {
		http.Error(w, "Failed to encode states", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// GetNodeState retrieves the complete state for a given node.
func GetNodeState(nodeName string) (NodeState, bool) {
	value, exists := nodeStates.Load(nodeName)
	if !exists {
		return NodeState{}, false // Return empty if no state is found
	}

	if state, ok := value.(NodeState); ok {
		return state, true
	}

	return NodeState{}, false // Return empty if conversion fails
}
