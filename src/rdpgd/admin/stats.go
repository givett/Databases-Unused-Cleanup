package admin

import (
	"encoding/json"
	"net/http"
)

// StatsHandler handle http request
type StatsHandler struct {
	stats Stats
}

// Stats interface used to aid in testing
type Stats interface {
	GetStats() interface{}
}

// AgentStats actual struct to get data
type AgentStats struct {
	QueueDepth      int `json:"task_queue_depth"`
	NumBoundDB      int `json:"num_bound_db"`
	NumFreeDB       int `json:"num_free_db"`
	NumReplSlots    int `json:"num_replication_slots"`
	NumDBBackupDisk int `json:"num_db_backup_files_on_disk"`
}

// MockStats used for testing mock data
type MockStats struct {
	Foo string
}

//GetStats return stats for agent
func (a *AgentStats) GetStats() interface{} {
	a.QueueDepth = getQueueDepth()
	a.NumBoundDB = getNumberOfBoundDatabases()
	a.NumFreeDB = getNumberOfFreeDatabases()
	a.NumReplSlots = getNumberOfReplicationSlots()
	a.NumDBBackupDisk = getNumberOfDatabaseBackupOnDisk()
	return a
}

//GetStats get mock data
func (m *MockStats) GetStats() interface{} {
	m.Foo = "123"
	return m
}

//NewStatsHandler create new stat handler
func NewStatsHandler(stats Stats) StatsHandler {
	return StatsHandler{
		stats: stats,
	}
}

//ServeHTTP serves http request
func (s *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	msg, err := json.Marshal(s.stats.GetStats())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(msg)
}
