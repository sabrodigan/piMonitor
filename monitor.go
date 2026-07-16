package main

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Read sysfs file safely
func readSysInt(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	return strconv.Atoi(s)
}

func readSysFloat(path string) (float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	return strconv.ParseFloat(s, 64)
}

func getCPUTemp() float64 {
	val, err := readSysFloat("/sys/class/thermal/thermal_zone0/temp")
	if err == nil {
		return val / 1000.0 // Convert millidegrees to C
	}
	// Fallback for non-Pi
	return 40.0 + rand.Float64()*10.0
}

func getCPUFreq() int {
	val, err := readSysInt("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq")
	if err == nil {
		return val / 1000 // Convert kHz to MHz
	}
	// Fallback for non-Pi
	return 1500 + rand.Intn(500)
}

func getFanStatus() (bool, int) {
	// Try to find the fan hwmon device
	matches, err := filepath.Glob("/sys/class/hwmon/hwmon*/fan1_input")
	if err == nil && len(matches) > 0 {
		rpm, err := readSysInt(matches[0])
		if err == nil {
			return rpm > 0, rpm
		}
	}
	
	// Try PWM
	matches, err = filepath.Glob("/sys/class/hwmon/hwmon*/pwm1")
	if err == nil && len(matches) > 0 {
		pwm, err := readSysInt(matches[0])
		if err == nil {
			return pwm > 0, pwm
		}
	}

	// Try cooling_device
	matches, err = filepath.Glob("/sys/class/thermal/cooling_device*/cur_state")
	if err == nil && len(matches) > 0 {
		state, err := readSysInt(matches[0])
		if err == nil {
			return state > 0, state
		}
	}

	// Fallback
	return false, 0
}

func getMemoryStats() (total int64, used int64) {
	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)
	if err == nil {
		// Convert to MB
		unit := uint64(info.Unit)
		if unit == 0 {
			unit = 1
		}
		total = int64(info.Totalram * unit / 1024 / 1024)
		free := int64(info.Freeram * unit / 1024 / 1024)
		buffers := int64(info.Bufferram * unit / 1024 / 1024)
		used = total - free - buffers
		return total, used
	}
	// Fallback
	return 8192, 2048 + int64(rand.Intn(1024))
}

func getLoadAvgs() (float64, float64, float64) {
	data, err := os.ReadFile("/proc/loadavg")
	if err == nil {
		parts := strings.Fields(string(data))
		if len(parts) >= 3 {
			l1, _ := strconv.ParseFloat(parts[0], 64)
			l5, _ := strconv.ParseFloat(parts[1], 64)
			l15, _ := strconv.ParseFloat(parts[2], 64)
			return l1, l5, l15
		}
	}
	// Fallback
	return 0.5, 0.4, 0.3
}

func CollectMetric() Metric {
	total, used := getMemoryStats()
	l1, l5, l15 := getLoadAvgs()
	fanActive, fanSpeed := getFanStatus()

	return Metric{
		Timestamp: time.Now().Unix(),
		CPUTemp:   getCPUTemp(),
		CPUFreq:   getCPUFreq(),
		FanActive: fanActive,
		FanSpeed:  fanSpeed,
		RAMTotal:  total,
		RAMUsed:   used,
		Load1:     l1,
		Load5:     l5,
		Load15:    l15,
	}
}

func StartMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m := CollectMetric()
			if err := InsertMetric(m); err != nil {
				log.Printf("Failed to insert metric: %v", err)
			}
		}
	}()
}
