package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/lFer17/codebase-maker/internal/agents/server"
)

func main() {
	openApikey := flag.String("openai-key", "", "OpenAI API key")
	outputDir := flag.String("output-dir", "./output", "Output directory for generated files")
	port := flag.String("port", "3000", "Server port")

	flag.Parse()

	if *openApikey == "" {
		err := godotenv.Load()
		*openApikey = os.Getenv("OPENAI_KEY")
		if *openApikey == "" && err != nil {
			fmt.Println("Please Provide OpenAi Api key using -openai-key flag or set OPENAI_KEY environment variable")
			os.Exit(1)
		}
	}

	srv := server.NewServer(*openApikey, *outputDir)

	http.Handle("/", http.FileServer(http.Dir("web/static")))

	http.HandleFunc("/api/generate", srv.HandleGenerate)
	http.HandleFunc("/download/", srv.HandleDownload)

	log.Printf("Server starting on http://localhost:%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))

}
