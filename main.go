package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

// build is the GitHub commit this binary was built from, shown in the page's
// credit line. Set with: go build -ldflags "-X main.build=$(git rev-parse --short HEAD)"
var build = "dev"

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, struct{ Build string }{build})
}

func latestMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	m, err := GetLatestMetric()
	if err != nil {
		http.Error(w, `{"error": "no data"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(m)
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	hoursStr := r.URL.Query().Get("hours")
	hours := 1
	if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
		hours = h
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()
	metrics, err := GetHistoricalMetrics(since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if metrics == nil {
		metrics = []Metric{} // Return empty array instead of null
	}
	json.NewEncoder(w).Encode(metrics)
}

func main() {
	// Initialize DB
	InitDB("metrics.db")

	// Start background metric collection every 5 seconds
	StartMonitoring(5 * time.Second)

	// Setup HTTP server
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/latest", latestMetricHandler)
	http.HandleFunc("/api/history", historyHandler)

	log.Println("Starting piMonitor on port 8092...")
	if err := http.ListenAndServe(":8092", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
