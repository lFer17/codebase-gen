package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lFer17/codebase-maker/internal/agents"
)

type Server struct {
	agent      *agents.Agent
	upgrader   websocket.Upgrader
	openAIkey  string
	outputBase string
}

type WebSocketClient struct {
	conn       *websocket.Conn
	writeMutex sync.Mutex
}

func NewWebSocketClient(conn *websocket.Conn) *WebSocketClient {
	return &WebSocketClient{
		conn: conn,
	}
}

func (c *WebSocketClient) WriteJSON(v interface{}) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	return c.conn.WriteJSON(v)
}

type ProjectRequest struct {
	Prompt      string `json:"prompt"`
	Language    string `json:"language"`
	Template    string `json:"template"`
	BasePackage string `json:"basePackage"`
	WorkerCount int    `json:"workerCount"`
	Model       string `json:"model"`
	ProjectName string `json:"projectName"`
}
type ProgressEvent struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	File       string `json:"file,omitempty"`
	Error      string `json:"error,omitempty"`
	ZipURL     string `json:"zipUrl,omitempty"`
	ProjectDir string `json:"projectDir,omitempty"`
}

func NewServer(openAIKey, outputBase string) *Server {
	if err := os.MkdirAll(outputBase, 0755); err != nil {
		log.Printf("Failed to create output base directory:%v", err)
	}

	return &Server{
		openAIkey:  openAIKey,
		outputBase: outputBase,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
		return
	}

	defer conn.Close()

	wsClient := NewWebSocketClient(conn)

	var req ProjectRequest

	err = conn.ReadJSON(&req)

	if err != nil {
		sendEvent(wsClient, ProgressEvent{
			Type:  "error",
			Error: "Invalid request: " + err.Error(),
		})
		return
	}
	projectName := req.ProjectName

	if projectName == "" {
		projectName = fmt.Sprintf("%s-project", req.Language)
	}

	sessionID := uuid.New().String()

	sessionDir := filepath.Join(s.outputBase, sessionID)

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		sendEvent(wsClient, ProgressEvent{
			Type:  "error",
			Error: "Failed to create a session directory: " + err.Error(),
		})
		return
	}

	projectDir := filepath.Join(sessionDir, projectName)

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		sendEvent(wsClient, ProgressEvent{
			Type:  "Error",
			Error: "Failed to create project directory: " + err.Error(),
		})
		return
	}

	ctx := context.Background()
	// Consider use streaming function from OpenAi
	httpClient := http.Client{
		Timeout: 1000 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			ResponseHeaderTimeout: 1000 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			DisableCompression:    false,
			ExpectContinueTimeout: 5 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   1000 * time.Second,
				KeepAlive: 1000 * time.Second,
			}).DialContext,
		},
	}

	client := agents.NewOpenAI(ctx, s.openAIkey, req.Model, &httpClient)

	progressCallBack := func(eventType, message, file string) {
		sendEvent(wsClient, ProgressEvent{
			Type:       eventType,
			Message:    message,
			File:       file,
			ProjectDir: projectName,
		})
	}

	agent, err := agents.NewAgentWithCallback(
		ctx, client, projectDir, req.BasePackage,
		req.Template, req.Language, req.WorkerCount,
		progressCallBack,
	)

	if err != nil {
		sendEvent(wsClient, ProgressEvent{
			Type:  "error",
			Error: "Failed to initialize agent: " + err.Error(),
		})
		return
	}

	agent.Start()

	sendEvent(wsClient, ProgressEvent{
		Type:    "start",
		Message: "Starting code generation",
	})

	if err := agent.GenerateCode(req.Prompt); err != nil {
		sendEvent(wsClient, ProgressEvent{
			Type:  "error",
			Error: "Code generation failed: " + err.Error(),
		})
		agent.Stop()
		return
	}

	time.Sleep(1 * time.Second)
	agent.Stop()

	zipName := fmt.Sprintf("%s.zip", projectName)
	zipPath := filepath.Join(sessionDir, zipName)

	sendEvent(wsClient, ProgressEvent{
		Type:  "file",
		Error: "Generating Zip file: " + zipName,
	})

	if err := createZip(projectDir, zipPath); err != nil {
		sendEvent(wsClient, ProgressEvent{
			Type:  "error",
			Error: "Failed to create zip file: " + err.Error(),
		})
	}

	zipURL := "/download/" + sessionID

	sendEvent(wsClient, ProgressEvent{
		Type:    "complete",
		Message: "Code generation completed!",
		ZipURL:  zipURL,
	})
}

func sendEvent(client *WebSocketClient, event ProgressEvent) {
	err := client.WriteJSON(event)

	if err != nil {
		log.Printf("Error writing data to connection: %v \n", err)
	}

}
