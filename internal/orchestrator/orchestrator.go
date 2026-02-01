package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Orchestrator struct {
	serverURL string
	username  string
	password  string
	client    *http.Client
	stopCh    chan struct{}
	config    Config
}

type Config struct {
	HeartbeatInterval time.Duration
	IdleTimeout       time.Duration
	PollInterval      time.Duration
}

func New(serverURL, username, password string) *Orchestrator {
	return &Orchestrator{
		serverURL: serverURL,
		username:  username,
		password:  password,
		client:    &http.Client{},
		stopCh:    make(chan struct{}),
		config: Config{
			HeartbeatInterval: 1 * time.Second,
			IdleTimeout:       10 * time.Second,
			PollInterval:      5 * time.Second,
		},
	}
}

func (o *Orchestrator) Start() error {
	log.Printf("Starting orchestrator, connecting to %s", o.serverURL)
	if err := o.loadConfig(); err != nil {
		log.Printf("Warning: failed to load config: %v", err)
	}
	go o.heartbeatLoop()
	go o.processLoop()
	log.Println("Orchestrator started")
	return nil
}

func (o *Orchestrator) Stop() {
	log.Println("Stopping orchestrator")
	close(o.stopCh)
}

func (o *Orchestrator) Wait() {
	<-o.stopCh
}

func (o *Orchestrator) heartbeatLoop() {
	ticker := time.NewTicker(o.config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-o.stopCh:
			return
		case <-ticker.C:
			if err := o.sendHeartbeat(); err != nil {
				log.Printf("Heartbeat failed: %v", err)
			}
		}
	}
}

func (o *Orchestrator) processLoop() {
	ticker := time.NewTicker(o.config.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-o.stopCh:
			return
		case <-ticker.C:
			if err := o.processTasks(); err != nil {
				log.Printf("Task processing failed: %v", err)
			}
		}
	}
}

func (o *Orchestrator) sendHeartbeat() error {
	body := map[string]string{"status": "active"}
	_, err := o.request("POST", "/api/v1/orchestrator/heartbeat", body)
	return err
}

func (o *Orchestrator) loadConfig() error {
	data, err := o.request("GET", "/api/v1/config", nil)
	if err != nil {
		return err
	}
	var configs []map[string]interface{}
	if err := json.Unmarshal(data, &configs); err != nil {
		return err
	}
	for _, cfg := range configs {
		key, _ := cfg["key"].(string)
		value, _ := cfg["value"].(string)
		switch key {
		case "orchestrator.heartbeat":
			if ms, err := time.ParseDuration(value + "ms"); err == nil {
				o.config.HeartbeatInterval = ms
			}
		case "orchestrator.idle":
			if ms, err := time.ParseDuration(value + "ms"); err == nil {
				o.config.IdleTimeout = ms
			}
		case "orchestrator.poll":
			if ms, err := time.ParseDuration(value + "ms"); err == nil {
				o.config.PollInterval = ms
			}
		}
	}
	log.Printf("Loaded config: heartbeat=%v, idle=%v, poll=%v",
		o.config.HeartbeatInterval, o.config.IdleTimeout, o.config.PollInterval)
	return nil
}

func (o *Orchestrator) processTasks() error {
	data, err := o.request("GET", "/api/v1/tasks?is_complete=false", nil)
	if err != nil {
		return err
	}
	var tasks []map[string]interface{}
	if err := json.Unmarshal(data, &tasks); err != nil {
		return err
	}
	log.Printf("Processing %d incomplete tasks", len(tasks))
	for _, task := range tasks {
		taskID, _ := task["id"].(string)
		status, _ := task["status"].(string)
		workerID, _ := task["worker_id"].(string)
		if workerID != "" && status == "active" {
			continue
		}
		if status == "idle" {
			log.Printf("Task %s is idle, checking for available workers", taskID)
		}
	}
	return nil
}

func (o *Orchestrator) request(method, path string, body interface{}) ([]byte, error) {
	url := o.serverURL + path
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(o.username, o.password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}
