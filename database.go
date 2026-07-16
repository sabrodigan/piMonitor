package main

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type Metric struct {
	Timestamp int64   `json:"timestamp"`
	CPUTemp   float64 `json:"cpuTemp"`
	CPUFreq   int     `json:"cpuFreq"`
	FanActive bool    `json:"fanActive"`
	FanSpeed  int     `json:"fanSpeed"`
	RAMTotal  int64   `json:"ramTotal"`
	RAMUsed   int64   `json:"ramUsed"`
	Load1     float64 `json:"load1"`
	Load5     float64 `json:"load5"`
	Load15    float64 `json:"load15"`
}

func InitDB(filepath string) {
	var err error
	db, err = sql.Open("sqlite", filepath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER,
		cpu_temp REAL,
		cpu_freq INTEGER,
		fan_active BOOLEAN,
		fan_speed INTEGER,
		ram_total INTEGER,
		ram_used INTEGER,
		load1 REAL,
		load5 REAL,
		load15 REAL
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON metrics(timestamp);
	`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

func InsertMetric(m Metric) error {
	query := `
		INSERT INTO metrics (timestamp, cpu_temp, cpu_freq, fan_active, fan_speed, ram_total, ram_used, load1, load5, load15)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, m.Timestamp, m.CPUTemp, m.CPUFreq, m.FanActive, m.FanSpeed, m.RAMTotal, m.RAMUsed, m.Load1, m.Load5, m.Load15)
	return err
}

func GetLatestMetric() (Metric, error) {
	var m Metric
	query := `
		SELECT timestamp, cpu_temp, cpu_freq, fan_active, fan_speed, ram_total, ram_used, load1, load5, load15
		FROM metrics
		ORDER BY timestamp DESC
		LIMIT 1
	`
	err := db.QueryRow(query).Scan(&m.Timestamp, &m.CPUTemp, &m.CPUFreq, &m.FanActive, &m.FanSpeed, &m.RAMTotal, &m.RAMUsed, &m.Load1, &m.Load5, &m.Load15)
	return m, err
}

// GetHistoricalMetrics retrieves metrics for charting, e.g., the last N hours
func GetHistoricalMetrics(since int64) ([]Metric, error) {
	query := `
		SELECT timestamp, cpu_temp, cpu_freq, fan_active, fan_speed, ram_total, ram_used, load1, load5, load15
		FROM metrics
		WHERE timestamp >= ?
		ORDER BY timestamp ASC
	`
	rows, err := db.Query(query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var m Metric
		if err := rows.Scan(&m.Timestamp, &m.CPUTemp, &m.CPUFreq, &m.FanActive, &m.FanSpeed, &m.RAMTotal, &m.RAMUsed, &m.Load1, &m.Load5, &m.Load15); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// CleanupOldMetrics removes records older than retentionDays to prevent infinite DB growth
func CleanupOldMetrics(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Unix()
	_, err := db.Exec("DELETE FROM metrics WHERE timestamp < ?", cutoff)
	return err
}
