package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Worker struct {
	serverURL   string
	username    string
	password    string
	client      *http.Client
	stopCh      chan struct{}
	config      Config
	currentTask string
}

type Config struct {
	HeartbeatInterval time.Duration
	IdleTimeout       time.Duration
	PollInterval      time.Duration
}

type WorkResponse struct {
	Task interface{} `json:"task,omitempty"`
	Role interface{} `json:"role,omitempty"`
}

func New(serverURL, username, password string) *Worker {
	return &Worker{
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

func (w *Worker) Start() error {
	log.Printf("Starting worker %s, connecting to %s", w.username, w.serverURL)
	if err := w.loadConfig(); err != nil {
		log.Printf("Warning: failed to load config: %v", err)
	}
	go w.heartbeatLoop()
	go w.workLoop()
	log.Println("Worker started")
	return nil
}

func (w *Worker) Stop() {
	log.Println("Stopping worker")
	close(w.stopCh)
}

func (w *Worker) Wait() {
	<-w.stopCh
}

func (w *Worker) heartbeatLoop() {
	ticker := time.NewTicker(w.config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.sendHeartbeat(); err != nil {
				log.Printf("Heartbeat failed: %v", err)
			}
		}
	}
}

func (w *Worker) workLoop() {
	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.requestAndProcessWork(); err != nil {
				log.Printf("Work processing failed: %v", err)
			}
		}
	}
}

func (w *Worker) sendHeartbeat() error {
	body := map[string]interface{}{"status": w.getStatus()}
	if w.currentTask != "" {
		body["task_id"] = w.currentTask
	}
	_, err := w.request("POST", "/api/v1/workers/heartbeat", body)
	return err
}

func (w *Worker) getStatus() string {
	if w.currentTask != "" {
		return "working"
	}
	return "idle"
}

func (w *Worker) loadConfig() error {
	data, err := w.request("GET", "/api/v1/config", nil)
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
		case "worker.heartbeat":
			if ms, err := time.ParseDuration(value + "ms"); err == nil {
				w.config.HeartbeatInterval = ms
			}
		case "worker.idle":
			if ms, err := time.ParseDuration(value + "ms"); err == nil {
				w.config.IdleTimeout = ms
			}
		case "worker.poll":
			if ms, err := time.ParseDuration(value + "ms"); err == nil {
				w.config.PollInterval = ms
			}
		}
	}
	log.Printf("Loaded config: heartbeat=%v, idle=%v, poll=%v",
		w.config.HeartbeatInterval, w.config.IdleTimeout, w.config.PollInterval)
	return nil
}

func (w *Worker) requestAndProcessWork() error {
	data, err := w.request("POST", "/api/v1/workers/request", nil)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	var workResp WorkResponse
	if err := json.Unmarshal(data, &workResp); err != nil {
		return err
	}
	if workResp.Task != nil {
		taskMap, ok := workResp.Task.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid task format")
		}
		taskID, _ := taskMap["id"].(string)
		title, _ := taskMap["title"].(string)
		log.Printf("Received work: Task %s - %s", taskID, title)
		w.currentTask = taskID
		if err := w.processTask(taskMap, workResp.Role); err != nil {
			log.Printf("Failed to process task: %v", err)
			w.currentTask = ""
			return err
		}
		log.Printf("Completed task %s", taskID)
		w.currentTask = ""
	}
	return nil
}

func (w *Worker) processTask(task map[string]interface{}, role interface{}) error {
	taskID, _ := task["id"].(string)
	log.Printf("Processing task %s", taskID)
	time.Sleep(2 * time.Second)
	_, err := w.request("POST", fmt.Sprintf("/api/v1/tasks/%s/complete", taskID), nil)
	return err
}

func (w *Worker) request(method, path string, body interface{}) ([]byte, error) {
	url := w.serverURL + path
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
	req.SetBasicAuth(w.username, w.password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := w.client.Do(req)
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
