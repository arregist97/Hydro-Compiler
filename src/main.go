package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/arregist97/Hydro-Compiler/tokenizer"
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
	var empty []string
	test := tokenizer.recTokenize(content, empty)
	fmt.Println(test)
	re := regexp.MustCompile(`([\s()])`)

	content = re.ReplaceAllStringFunc(content, func(s string) string {
		if s == "(" {
			return s + " "
		}
		return " " + s + " "
	})

	parsedContent := strings.Fields(content)
	fmt.Println(parsedContent)

	return parsedContent
}

func recTokenize(content string, tokens []string) []string {
	var token string
	var updatedContent string
	var err error

	token, updatedContent, err = buildToken(content)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}
	fmt.Println(token + "/end")
	if len(token) > 0 {
		tokens = append(tokens, token)
	}
	if len(content) > 0 {
		return recTokenize(updatedContent, tokens)
	}
	return tokens
}

func buildToken(content string, iOpt ...uint8) (string, string, error) {
	var updatedToken string
	var updatedContent string
	var err error = nil
	var i uint8 = 0

	if len(iOpt) > 0 {
		i = iOpt[0]
	}

	if uint8(len(content)) <= i {
		return "", "", err
	}

	r, s := utf8.DecodeRuneInString(content[i:])
	if r == utf8.RuneError {
		return "", "", errors.New("Could not recognize token " + content[i:i+1]) 
	}

	size := uint8(s)
	peek, _ := utf8.DecodeRuneInString(content[i+size:])

	if r == ' ' && i == 0 {
		updatedToken, updatedContent, err = buildToken(content[size:])
	} else if isEndOfToken(r) || peek == utf8.RuneError || isEndOfToken(peek) {
		updatedToken = string(r)
		updatedContent = content[i + size:]
	} else {
		var token string
		token, updatedContent, err = buildToken(content, i + size)
		updatedToken = string(r) + token
	}
	return updatedToken, updatedContent, err
}

func isEndOfToken(a rune) bool {
	var endOfTokenRunes = [...]rune {'(', ')', ' ', '\n'}
	for _, b := range endOfTokenRunes {
		if b == a {
			return true
		}
	}
	return false
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
