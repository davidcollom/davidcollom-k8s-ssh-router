package metrics

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var activeSessions = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "active_ssh_sessions",
	Help: "Number of active SSH sessions",
})

func init() {
	prometheus.MustRegister(activeSessions)
}

func StartMetricsServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func IncActiveSessions() {
	activeSessions.Inc()
}

func DecActiveSessions() {
	activeSessions.Dec()
}
