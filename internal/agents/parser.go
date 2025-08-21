package agents

import (
	"log"
	"regexp"
	"strings"
)

func (a *Agent) ParserCode(content string) error {

	codeBlockRegex := regexp.MustCompile(`(?s)---FILE_PATH: (.+?)\n(.*?)---END_FILE`)

	matches := codeBlockRegex.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		log.Printf("Could not finde FILE_PATH in codeblock")

		return nil
	}

	for _, match := range matches {
		if len(match) < 3 {
			log.Printf("Invalid match found: %v", match)
			continue
		}

		filePath := strings.TrimSpace(match[1])
		code := strings.TrimSpace(match[2])

		code = regexp.MustCompile("^```[a-zA-Z0-9]*\n").ReplaceAllString(code, "")
		code = regexp.MustCompile("\n```$").ReplaceAllString(code, "")

		a.taskQueue <- fileTask{
			Path:    filePath,
			Content: code,
		}

	}

	return nil
}
