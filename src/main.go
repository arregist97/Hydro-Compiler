package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

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

	// tokens := tokenize(string(content))

	// buffer, err := parse(tokens)
	// if err != nil {
	// 	fmt.Println("Parse error:", err)
	// 	return
	// }

	var empty []string
	tokens := tokenizer.RecTokenize(string(content), empty)
	fmt.Println(tokens)
	tree := tokenizer.BuildTokenTree(tokens)
	tokenizer.PrintTokenTree(tree)

	buffer, err := parseTree(tree)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buffer)

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
	fmt.Println(parsedContent)

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

func parseTree(node *tokenizer.TokenTreeNode) (string, error) {
	var buffer string
	buffer = "global _start"
	buffer = buffer + "\n" + "_start:"
	return evalStmt(node, buffer)
}

func evalStmt(node *tokenizer.TokenTreeNode, buffer string) (string, error) {
	if node.TokenType[0] != "Stmt" {
		return "", errors.New("Stmt expected. Recieved " + node.TokenType[0])
	}
	if node.Val == "exit" {
		return evalExit(node.Right, buffer)
	}
	return "", errors.New("Undefined Stmt: " + node.Val)
}

func evalExit(node *tokenizer.TokenTreeNode, buffer string) (string, error) {
	if node.Val != "(" {
		return "", errors.New("Expected `(` after exit")
	}
	buffer = buffer + "\n" + "  mov    rax, 60"
	buffer, err := evalExpr(node.Right, buffer, "rdi")
	buffer = buffer + "\n" + "  syscall"
	return buffer, err
}

func evalExpr(node *tokenizer.TokenTreeNode, buffer string, register string) (string, error) {
	if node.TokenType[0] != "Expr" {
		return "", errors.New("Expr expected, recieved " + node.TokenType[0])
	}
	if node.TokenType[1] == "Term" {
		return evalTerm(node, buffer, register)
	}
	return "", errors.New("Invalid Expr: " + node.TokenType[1])
}

func evalTerm(node *tokenizer.TokenTreeNode, buffer string, register string) (string, error) {
	if node.TokenType[2] == "intLit" {
		buffer = buffer + "\n" + "  mov    " + register + ", " + node.Val
		return buffer, nil
	}
	return "", errors.New("Invalid Term: " + node.TokenType[2])
}
