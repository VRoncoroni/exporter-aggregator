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
	http.HandleFunc("/health", healthHandler)

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

	for _, res := range results {
		output += fmt.Sprintf("# TARGET %s (%s)\n", res.Target.Name, res.Target.URL)
		output += res.Data + "\n"
		output += fmt.Sprintf("aggregator_target_up{name=\"%s\",url=\"%s\"} %d\n\n",
			res.Target.Name,
			res.Target.URL,
			res.Up,
		)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(output))
}

// ----------- HEALTH -----------

func healthHandler(w http.ResponseWriter, r *http.Request) {
	ch := make(chan Result)

	for _, t := range config.Targets {
		go func(target Target) {
			up := checkTarget(target)
			ch <- Result{Target: target, Up: up}
		}(t)
	}

	allUp := true
	output := "["

	for i := 0; i < len(config.Targets); i++ {
		res := <-ch

		status := "down"
		if res.Up == 1 {
			status = "up"
		} else {
			allUp = false
		}

		output += fmt.Sprintf(`{"name":"%s","url":"%s","status":"%s"}`,
			res.Target.Name,
			res.Target.URL,
			status,
		)

		if i < len(config.Targets)-1 {
			output += ","
		}
	}

	output += "]"

	code := 200
	globalStatus := "ok"
	if !allUp {
		code = 503
		globalStatus = "degraded"
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"status":"%s","targets":%s}`, globalStatus, output)))
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
