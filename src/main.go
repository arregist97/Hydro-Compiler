package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Check if a command line argument is provided
	if len(os.Args) != 2 {
		fmt.Println("Incorrect Usage. Expected:")
		fmt.Println("main.go <filename>")
		return
	}

	// Get the command line argument
        fileName := os.Args[1]

	// Read the contents of the file
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

        // Parse the contents of the file
	tokens := tokenize(string(content))

	// Print the parsed content
	for _, str := range tokens {
		fmt.Println(str)
	}
}

// parse function to split content into strings separated by ' ', '(', ')', or newline characters
func tokenize(content string) []string {
	// Create a regular expression to match '(', ')' and newline
	re := regexp.MustCompile(`([\s()])`)

	// Replace matches with spaces around them
	content = re.ReplaceAllStringFunc(content, func(s string) string {
		if s == " " || s == "\n" {
			return " "
		}
		return " " + s + " "
	})

	// Split the content by spaces
	parsedContent := strings.Fields(content)

	return parsedContent
}
