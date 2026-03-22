package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ----------- STRUCTS -----------

type Target struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	Port      int      `json:"port"`
	TimeoutMs int      `json:"timeout_ms"`
	Targets   []Target `json:"targets"`
}

type Result struct {
	Target Target
	Up     int
	Data   string
}

// ----------- GLOBAL -----------

var config Config
var configPath string

// ----------- MAIN -----------

func main() {
	flag.StringVar(&configPath, "config", "config.json", "Path to config file")
	flag.Parse()
	loadConfig(configPath)

	http.HandleFunc("/metrics", metricsHandler)

	fmt.Println("🚀 Aggregator running on port", config.Port)
	fmt.Println("Loaded targets:", len(config.Targets))
	fmt.Println("Config file:", configPath)

	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}

// ----------- CONFIG -----------

func loadConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Cannot open config file: %s", path))
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		panic("Invalid config.json")
	}
}

// ----------- METRICS -----------

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Result)

	for _, t := range config.Targets {
		go fetchMetrics(t, ch)
	}

	var results []Result
	for range config.Targets {
		results = append(results, <-ch)
	}

	output := ""

	// ---------- METADATA ----------
	output += "# HELP aggregator_target_up Health status of each target (1=up, 0=down)\n"
	output += "# TYPE aggregator_target_up gauge\n"

	allUp := true

	// ---------- TARGET METRICS ----------
	for _, res := range results {
		output += fmt.Sprintf("# TARGET %s (%s)\n", res.Target.Name, res.Target.URL)

		// metrics du target
		output += res.Data + "\n"

		// health metric
		if res.Up == 0 {
			allUp = false
		}

		output += fmt.Sprintf(
			"aggregator_target_up{name=\"%s\",url=\"%s\"} %d\n\n",
			res.Target.Name,
			res.Target.URL,
			res.Up,
		)
	}

	// ---------- GLOBAL HEALTH ----------
	output += "# HELP aggregator_up Global health status (1=ok, 0=degraded)\n"
	output += "# TYPE aggregator_up gauge\n"

	global := 1
	if !allUp {
		global = 0
	}

	output += fmt.Sprintf("aggregator_up %d\n\n", global)

	// ---------- COUNT ----------
	output += "# HELP aggregator_targets_total Number of targets\n"
	output += "# TYPE aggregator_targets_total gauge\n"
	output += fmt.Sprintf("aggregator_targets_total %d\n", len(config.Targets))

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte(output))
}

// ----------- CORE -----------

func fetchMetrics(target Target, ch chan<- Result) {
	client := http.Client{
		Timeout: time.Duration(config.TimeoutMs) * time.Millisecond,
	}

	resp, err := client.Get(target.URL)
	if err != nil {
		ch <- Result{Target: target, Up: 0, Data: "# ERROR " + err.Error()}
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ch <- Result{Target: target, Up: 1, Data: string(body)}
	} else {
		ch <- Result{
			Target: target,
			Up:     0,
			Data:   fmt.Sprintf("# ERROR status %d", resp.StatusCode),
		}
	}
}

func checkTarget(target Target) int {
	client := http.Client{
		Timeout: time.Duration(config.TimeoutMs) * time.Millisecond,
	}

	resp, err := client.Get(target.URL)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return 1
	}
	return 0
}
