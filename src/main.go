package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

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
	if node.Val == "EOF" {
		return buffer, nil
	}
	if len(node.TokenType) > 1 && node.TokenType[1] == "StmtTm" {
		return evalStmt(node.Right, buffer)
	}
	if node.Val == "exit" {
		return evalExit(node.Right, buffer)
	}
	if node.Val == "let" {
		return evalLet(node.Right, buffer)
	}
	return "", errors.New("Undefined Stmt: " + node.Val)
}

func evalExit(node *tokenizer.TokenTreeNode, buffer string) (string, error) {
	if node.Val != "(" {
		return "", errors.New("Expected `(` after exit")
	}
	buffer, err := evalExpr(node, buffer)
	buffer = buffer + "\n" + "  mov    rax, 60"
	buffer = buffer + "\n" + "  pop    rdi"
	buffer = buffer + "\n" + "  syscall"
	return buffer, err
}

func evalLet(node *tokenizer.TokenTreeNode, buffer string) (string, error) {
	if len(node.TokenType) > 2 && node.TokenType[2] != "ident" {
		log.Fatal("Improper declaration")
	}
	if node.Right.Val != "=" {
		log.Fatal("Expected '='")
	}
	//ToDo
	return "", nil
}

func evalExpr(node *tokenizer.TokenTreeNode, buffer string) (string, error) {
	if node.TokenType[0] != "Expr" {
		fmt.Println("Node val: " + node.Val)
		return "", errors.New("Expr expected, recieved " + node.TokenType[0])
	}
	if node.TokenType[1] == "Term" {
		return evalTerm(node, buffer)
	}
	if node.Val == "(" {
		buf, err := evalExpr(node.Left, buffer)
		if err != nil {
			return "", err
		}
		return evalTerminator(node.Right, buf)
	}
	return "", errors.New("Invalid Expr: " + node.TokenType[1])
}

func evalTerm(node *tokenizer.TokenTreeNode, buffer string) (string, error) {
	if node.TokenType[2] == "intLit" {
		buffer = buffer + "\n" + "  mov    rax, " + node.Val
		buffer = buffer + "\n" + "  push   rax"
		return buffer, nil
	}
	if node.TokenType[2] == "ident" {
	}
	return "", errors.New("Invalid Term: " + node.TokenType[2])
}

func evalTerminator(node *tokenizer.TokenTreeNode, buffer string) (string, error) {
	if node == nil || node.Val == ")" {
		return buffer, nil
	}
	if len(node.TokenType) > 1 && node.TokenType[1] == "StmtTm" {
		return evalStmt(node.Right, buffer)
	}
	return "", errors.New("Invalid Terminator: " + node.Val)
}
