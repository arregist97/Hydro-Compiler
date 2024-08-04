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

type state struct {
	//parenCount int
}

func parseTree(node *tokenizer.TokenTreeNode) (string, error) {
	var buffer string
	buffer = "global _start"
	buffer = buffer + "\n" + "_start:"
	state := state{}
	buffer, _, err := evalStmt(node, buffer, &state)
	return buffer, err
}

func evalStmt(node *tokenizer.TokenTreeNode, buffer string, state *state) (string, *state, error) {
	fmt.Println("Val: " + node.Val)
	if node.TokenType[0] != "Stmt" {
		return "", state, errors.New("Stmt expected. Recieved " + node.TokenType[0])
	}
	if node.Val == "EOF" {
		buffer = buffer + "\n" + "  mov    rax, 60"
		buffer = buffer + "\n" + "  mov    rdi, 0"
		buffer = buffer + "\n" + "  syscall"
		return buffer, nil
	}
	if len(node.TokenType) > 1 && node.TokenType[1] == "StmtTm" {
		return evalStmt(node.Right, buffer, state)
	}
	if node.Val == "exit" {
		buffer, state, err := evalExit(node.Left, buffer, state)
		if err != nil {
			return "", err
		}
		return evalStmt(node.Right, buffer, state)
	}
	if node.Val == "let" {
		return evalLet(node.Right, buffer, state)
	}
	return "", state, errors.New("Undefined Stmt: " + node.Val)
}

func evalExit(node *tokenizer.TokenTreeNode, buffer string, state *state) (string, *state, error) {
	if node.Val != "(" {
		return "", state, errors.New("Expected `(` after exit")
	}
	buffer, state, err := evalExpr(node, buffer, state, false)
	if err != nil {
		return "", state, err
	}
	buffer = buffer + "\n" + "  mov    rax, 60"
	buffer = buffer + "\n" + "  pop    rdi"
	buffer = buffer + "\n" + "  syscall"
	return evalTerminator(node.Right, buffer, state)
}

func evalLet(node *tokenizer.TokenTreeNode, buffer string, state *state) (string, *state, error) {
	if len(node.TokenType) > 2 && node.TokenType[2] != "ident" {
		log.Fatal("Improper declaration")
	}
	if node.Right.Val != "=" {
		log.Fatal("Expected '='")
	}
	//ToDo
	return "", nil
}

func evalExpr(node *tokenizer.TokenTreeNode, buffer string, state *state, paren bool) (string, *state, error) {
	if node.TokenType[0] != "Expr" {
		fmt.Println("Node val: " + node.Val)
		return "", state, errors.New("Expr expected, recieved " + node.TokenType[0])
	}
	if node.TokenType[1] == "Term" {
		return evalTerm(node, buffer, state, paren)
	} else if node.Val == "(" {
		return evalExpr(node.Left, buffer, state, true)
	} else {
		return "", state, errors.New("Invalid Expr: " + node.TokenType[1])
	}
}

func evalTerm(node *tokenizer.TokenTreeNode, buffer string, state *state, paren bool) (string, *state, error) {
	if node.TokenType[2] == "intLit" {
		buffer = buffer + "\n" + "  mov    rax, " + node.Val
		buffer = buffer + "\n" + "  push   rax"
	} else if node.TokenType[2] == "ident" {
		//ToDo
	} else {
		return "", state, errors.New("Invalid Term: " + node.TokenType[2])
	}
	if paren && (node.Right == nil || node.Right.Val != ")") {
		return "", state, errors.New("Expected ')'")
	}
	return buffer, state, nil
}

func evalTerminator(node *tokenizer.TokenTreeNode, buffer string, state *state) (string, *state, error) {
	if len(node.TokenType) > 1 && node.TokenType[1] == "StmtTm" {
		return evalStmt(node.Right, buffer, state)
	}
	return "", state, errors.New("Invalid Terminator: " + node.Val)
}
