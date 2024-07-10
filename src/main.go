package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Incorrect Usage. Expected:")
		fmt.Println("main.go <filename>")
		return
	}

        fileName := os.Args[1]
	
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	tokens := tokenize(string(content))

	buffer, err := parse(tokens)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	re := regexp.MustCompile(`\.[^.]+$`)
	baseName := re.ReplaceAllString(fileName, "")
	newFileName := baseName + ".asm"

	newFile, err := os.Create(newFileName)
	if err != nil {
		fmt.Println("failed to create new file: %w", err)
	}
	defer newFile.Close()

	_, err = newFile.WriteString(buffer)
	if err != nil {
		fmt.Println("failed to write to new file: %w", err)
	}
}

func tokenize(content string) []string {
	re := regexp.MustCompile(`([\s()])`)

	content = re.ReplaceAllStringFunc(content, func(s string) string {
		if s == "(" {
			return s + " "
		}
		return " " + s + " "
	})

	parsedContent := strings.Fields(content)

	return parsedContent
}

func parse(tokens []string) (string, error){
	var buffer string
	var token string
	buffer = "global _start"
	buffer = buffer + "\n" + "_start:"
	token, tokens = tokens[0], tokens[1:]
	if token == "exit(" {
		buffer = buffer + "\n" + "  mov    rax, 60"
		number := tokens[0]
		buffer = buffer + "\n" + "  mov    rdi, " + number
		buffer = buffer + "\n" + "  syscall"
		return buffer, nil
	}
	return "", errors.New("Could not parse" + token)
}
