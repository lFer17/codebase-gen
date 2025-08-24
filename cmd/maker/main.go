package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/lFer17/codebase-maker/internal/agents"
)

func main() {

	openApikey := flag.String("openai-key", "", "OpenAI API key")
	outputDir := flag.String("output-dir", "./output", "Output directory for generated files")
	basePackage := flag.String("base-package", "github.com/user/app", "Base package for generated files")
	workerCount := flag.Int("worker-count", 4, "Number of concurrent workers")
	templateName := flag.String("template", "default", "Project to use")
	language := flag.String("language", "go", "programming language for the project")
	model := flag.String("model", "gpt-4o-mini", "OpenAI model to user")
	timeout := flag.Int("timeout", 120, "Time for OpenAI Api Calls")
	flag.Parse()

	if *openApikey == "" {
		err := godotenv.Load()
		*openApikey = os.Getenv("OPENAI_KEY")
		if *openApikey == "" && err != nil {
			fmt.Println("Please Provide OpenAi Api key using -openai-key flag or set OPENAI_KEY environment variable")
			os.Exit(1)
		}
	}

	ctx := context.Background()

	openAIClient := agents.NewOpenAI(ctx, *openApikey, *model, &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	})

	agent, err := agents.NewAgent(ctx,
		openAIClient,
		*outputDir,
		*basePackage,
		*templateName,
		*language,
		*workerCount)

	if err != nil {
		log.Fatal(err)

	}

	args := flag.Args()

	if len(args) == 0 {
		log.Printf("please pass arguments")
		os.Exit(1)
	}

	agent.Start()

	prompt := strings.Join(args, " ")

	if err = agent.GenerateCode(prompt); err != nil {
		log.Printf("error writing code: %v\n", err)
		agent.Stop()
		os.Exit(1)
	}

	time.Sleep(1 * time.Second)
	agent.Stop()
	fmt.Println("Finished writing project to", *outputDir)

}
