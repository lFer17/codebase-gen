package agents

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// GO embed templates
//
//go:embed templates/*
var templatesFS embed.FS

type fileTask struct {
	Path    string
	Content string
}

type ProjectTemplate struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Language    string            `json:"language"`
	Prompt      string            `json:"prompt"`
	Files       map[string]string `json:"files"`
}

type PromptTemplate struct {
	Description string `json:"description"`
	Language    string `json:"language"`
	Template    string `json:"template"`
}

type Agent struct {
	openAi          *OpenAPI
	outputDir       string
	basePackage     string
	taskQueue       chan fileTask
	wg              sync.WaitGroup
	workerCount     int
	ctx             context.Context
	cancel          context.CancelFunc
	fileWriterMutex sync.Mutex
	filesWritten    map[string]bool
	selectedTmpl    string
	language        string
	templates       map[string]ProjectTemplate
	promptsTmpl     map[string]PromptTemplate
}

var (
	Languages = []string{"Go", "Python", "JavaScript", "java"}
)

func NewAgent(ctx context.Context,
	openAI *OpenAPI,
	outputDir string,
	basePackage string,
	templateName string,
	language string,
	workerCount int,
) (*Agent, error) {
	ctx, cancel := context.WithCancel(ctx)

	agent := &Agent{
		openAi:       openAI,
		outputDir:    outputDir,
		basePackage:  basePackage,
		taskQueue:    make(chan fileTask, 100),
		workerCount:  workerCount,
		ctx:          ctx,
		cancel:       cancel,
		filesWritten: make(map[string]bool),
		selectedTmpl: templateName,
		language:     language,
	}
	if err := agent.loadTemplates(); err != nil {
		return nil, err
	}

	agent.loadPromptTemplates()

	return agent, nil
}

func (a *Agent) Start() {
	log.Printf("Starting %d workers...\n", a.workerCount)
	for i := 0; i < a.workerCount; i++ {
		a.wg.Add(1)
		go a.worker(i)
	}
}

func (a *Agent) worker(id int) {
	defer a.wg.Done()
	log.Printf("Worker %d started\n", id)

	for {
		select {
		case task, ok := <-a.taskQueue:
			if !ok {
				log.Printf("Worker %d: Task channel closed, exiting\n", id)
				return
			}

			a.fileWriterMutex.Lock()
			if a.filesWritten[task.Path] {
				log.Printf("Worker %d: File %s already written, skipping\n", id, task.Path)
				a.fileWriterMutex.Unlock()
				continue
			}

			a.filesWritten[task.Path] = true
			a.fileWriterMutex.Unlock()

			err := a.writeFile(task)

			if err != nil {
				log.Printf("Worker %d: Error writing file %s: %v\n", id, task.Path, err)
			} else {
				log.Printf("Worker %d: Successfully wrote file %s\n", id, task.Path)
			}

		case <-a.ctx.Done():
			log.Printf("Worker %d: Context cancelled, exiting\n", id)
			return
		}
	}
}

func (a *Agent) writeFile(task fileTask) error {
	fullPath := filepath.Join(a.outputDir, task.Path)

	dir := filepath.Dir(fullPath)

	// if err := os.MkdirAll(dir, 0755); err != nil {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", fullPath, err)
	}

	err := os.WriteFile(fullPath, []byte(task.Content), 0644)

	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	log.Printf("Writing file to %s\n", fullPath)

	return nil
}

func (a *Agent) SendFileTask(path string, content string) {
	task := fileTask{
		Path:    path,
		Content: content,
	}

	go func() {
		a.taskQueue <- task
	}()
}

func (a *Agent) Stop() {
	log.Println("Stopping agent...")
	close(a.taskQueue)
	a.cancel()
	a.wg.Wait()
}

func (a *Agent) loadTemplates() error {

	a.templates = make(map[string]ProjectTemplate)

	loaded := 0

	log.Println("Loading templates from embedded filesystem...")
	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		return fmt.Errorf("reading template directory: %w", err)
	}

	for _, file := range entries {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := templatesFS.ReadFile(filepath.Join("templates", file.Name()))
		if err != nil {
			log.Printf("Warning: Could not read template file %s: %v", file.Name(), err)
			continue
		}

		var tmpl ProjectTemplate
		if err := json.Unmarshal(data, &tmpl); err != nil {
			log.Printf("Warning: Invalid template format in %s: %v", file.Name(), err)
			continue
		}

		a.templates[tmpl.Name] = tmpl
		log.Printf("Loaded template: %s - %s (%s)", tmpl.Name, tmpl.Description, tmpl.Language)
		loaded++
	}

	// load user custom templates
	userCustomTemplatePath := "./templates"
	if _, err := os.Stat(userCustomTemplatePath); !os.IsNotExist(err) {
		dirs, err := os.ReadDir(userCustomTemplatePath)
		if err != nil {
			fmt.Printf("reading custom template directory: %v", err)
		} else {
			for _, file := range dirs {
				if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
					continue
				}

				data, err := templatesFS.ReadFile(filepath.Join(userCustomTemplatePath, file.Name()))
				if err != nil {
					log.Printf("Warning: Could not read template file %s: %v", file.Name(), err)
					continue
				}

				var tmpl ProjectTemplate
				if err := json.Unmarshal(data, &tmpl); err != nil {
					log.Printf("Warning: Invalid template format in %s: %v", file.Name(), err)
					continue
				}

				if _, exists := a.templates[tmpl.Name]; exists {
					log.Printf("User template '%s' overrides embedded template with same name", tmpl.Name)
				}

				a.templates[tmpl.Name] = tmpl
				log.Printf("Loaded template: %s - %s (%s)", tmpl.Name, tmpl.Description, tmpl.Language)
				loaded++
			}
		}
	}

	if loaded == 0 {

		log.Println("No templates fund., adding default templates")

		for _, lang := range Languages {
			a.templates[lang+"-default"] = ProjectTemplate{
				Name:        lang + "-default",
				Description: "Default " + lang + " application",
				Language:    lang,
				Prompt:      "",
				Files:       make(map[string]string),
			}

			loaded++
		}

		a.templates["default"] = ProjectTemplate{
			Name:        "default",
			Description: "Default generic application",
			Language:    "default",
			Prompt:      "",
			Files:       make(map[string]string),
		}

		loaded++
	}

	log.Printf("Loaded %d templates\n", loaded)

	return nil
}

func (a *Agent) loadPromptTemplates() {
	a.promptsTmpl = make(map[string]PromptTemplate)

	for _, p := range defaultPrompts {
		a.promptsTmpl[p.Language] = p
	}

	customPromptPath := "./templates/prompts"

	if _, err := os.Stat(customPromptPath); !os.IsNotExist(err) {
		files, err := os.ReadDir(customPromptPath)
		if err == nil {

			for _, file := range files {
				if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
					continue
				}
				data, err := os.ReadFile(filepath.Join(customPromptPath, file.Name()))

				if err != nil {
					log.Printf("Warning: could not read prompt template file %s:%v", file.Name(), err)
					continue
				}

				var tmpl PromptTemplate

				if err := json.Unmarshal(data, &tmpl); err != nil {
					log.Printf("Warning: Invalid Prompt Template format in %s:%v", file.Name())
					continue
				}

				if _, exists := a.promptsTmpl[tmpl.Language]; exists {
					log.Printf("User prompt template '%s' overrides embedded template with same name", tmpl.Language)
				}

				a.promptsTmpl[tmpl.Language] = tmpl
			}

		}

	}

}
