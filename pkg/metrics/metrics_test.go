package metrics

import (
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestStartMetricsServer(t *testing.T) {
	go StartMetricsServer(2113)

	req, err := http.NewRequest("GET", "http://127.0.0.1:2113/metrics", nil)
	assert.NoError(t, err, "Failed to create request")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "Failed to get response")

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
}

func TestActiveSessions(t *testing.T) {
	IncActiveSessions()
	IncActiveSessions()
	gatherers := prometheus.Gatherers{prometheus.DefaultGatherer}
	metrics, err := gatherers.Gather()
	assert.NoError(t, err, "Expected no error gathering metrics")

	var found bool
	for _, m := range metrics {
		if m.GetName() == "active_ssh_sessions" {
			assert.Equal(t, 2.0, *m.Metric[0].Gauge.Value, "Expected 2 active sessions")
			found = true
			break
		}
	}
	assert.True(t, found, "Expected active_ssh_sessions metric to be found")

	DecActiveSessions()
	metrics, err = gatherers.Gather()
	assert.NoError(t, err, "Expected no error gathering metrics")

	found = false
	for _, m := range metrics {
		if m.GetName() == "active_ssh_sessions" {
			assert.Equal(t, 1.0, *m.Metric[0].Gauge.Value, "Expected 1 active session")
			found = true
			break
		}
	}
	assert.True(t, found, "Expected active_ssh_sessions metric to be found")
}
