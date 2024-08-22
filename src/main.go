package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

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
	block := tokenizer.NewTokenTreeBlock()
	tree := tokenizer.BuildTokenTree(block, tokens)
	fmt.Println("\nToken Tree:")
	tokenizer.PrintTokenTree(tree)

	buffer, err := parseTree(tree)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buffer)

	fileName = filepath.Base(fileName)
	re := regexp.MustCompile(`\.[^.]+$`)
	baseName := re.ReplaceAllString(fileName, "")
	directory := "../build/"
	newFileName := directory + baseName + ".asm"

	newFile, err := os.Create(newFileName)
	if err != nil {
		fmt.Println("failed to create new file: %w", err)
	}
	defer newFile.Close()

	_, err = newFile.WriteString(buffer)
	if err != nil {
		fmt.Println("failed to write to new file: %w", err)
	}

	// Example Unix command: cat the file
	fmt.Println("nasm -felf64", newFileName)
	cmd := exec.Command("nasm -felf64", newFileName)

	// Attach the command's output to the current process's standard output
	cmd.Stdout = os.Stdout

	// Run the command
	err = cmd.Run()
	if err != nil {
		log.Fatal("command execution failed: %w", err)
	}
}

type state struct {
	stackPtr int
	lookupTbl map[string]int
}

func parseTree(node *tokenizer.TokenTreeNode) (string, error) {
	var buffer string
	buffer = "global _start"
	buffer = buffer + "\n" + "_start:"
	table := make(map[string]int)
	state := state{ stackPtr: 0, lookupTbl: table }
	buffer, err := evalStmt(node, buffer, &state)
	return buffer, err
}

func evalStmt(node *tokenizer.TokenTreeNode, buffer string, state *state) (string, error) {
	fmt.Println("Evaluating statement " + node.Val + "...")
	if node.TokenType[0] != "Stmt" {
		return "", errors.New("Stmt expected. Recieved " + node.TokenType[0])
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
		buf, err := evalExit(node.Left, buffer, state)
		if err != nil {
			return "", err
		}
		buffer = buf
		
	} else if node.Val == "let" {
		buf, err := evalLet(node.Right, buffer, state)
		if err != nil {
			return "", err
		}
		buffer = buf
		
		node = node.Right.Right
	} else {
		return "", errors.New("Undefined Stmt: " + node.Val)
	}
	return evalTerminator(node.Right, buffer, state) 
}

func evalExit(node *tokenizer.TokenTreeNode, buffer string, state *state) (string, error) {
	if node.Val != "(" {
		return "", errors.New("Expected `(` after exit")
	}
	buffer, err := evalExpr(node, buffer, state, false)
	if err != nil {
		return "", err
	}
	buffer = buffer + "\n" + "  mov    rax, 60"
	buffer = buffer + "\n" + "  pop    rdi"
	buffer = buffer + "\n" + "  syscall"
	state.stackPtr--
	return buffer, nil
}

func evalLet(node *tokenizer.TokenTreeNode, buffer string, state *state) (string, error) {
	if len(node.TokenType) > 2 && node.TokenType[2] != "ident" {
		log.Fatal("Improper declaration")
	}
	if node.Right.Val != "=" {
		log.Fatal("Expected '='")
	}
	buf, err := evalExpr(node.Right.Left, buffer, state, false)
	if err != nil {
		return "", nil
	}
	buffer = buf

	state.lookupTbl[node.Val] = state.stackPtr
	return buffer, nil
}

func evalExpr(node *tokenizer.TokenTreeNode, buffer string, state *state, paren bool) (string, error) {
	if node.TokenType[0] != "Expr" {
		fmt.Println("Node val: " + node.Val)
		return "", errors.New("Expr expected, recieved " + node.TokenType[0])
	}
	if node.Val == "(" {
		return evalExpr(node.Left, buffer, state, true)
	}
	if node.TokenType[1] == "Term" {
		return evalTerm(node, buffer, state, paren)
	}
	return "", errors.New("Invalid Expr: " + node.TokenType[1])
}

func evalTerm(node *tokenizer.TokenTreeNode, buffer string, state *state, paren bool) (string, error) {
	if node.TokenType[2] == "intLit" {
		buffer = buffer + "\n" + "  mov    rax, " + node.Val
		buffer = buffer + "\n" + "  push   rax"
		state.stackPtr++
	} else if node.TokenType[2] == "ident" {
		stackLoc, validIdent := state.lookupTbl[node.Val]
		if !validIdent {
			return "", errors.New("Undeclared ident " + node.Val)
		}
		fmt.Println("Stack Pointer", state.stackPtr, "var location", stackLoc)
		stackOffset := (state.stackPtr - stackLoc) * 8
		fmt.Println(stackOffset)
		buffer = buffer + "\n" + "  push   QWORD [rsp + " + strconv.Itoa(stackOffset) + "]"
		state.stackPtr++
	} else {
		return "", errors.New("Invalid Term: " + node.TokenType[2])
	}
	if paren && (node.Right == nil || node.Right.Val != ")") {
		return "", errors.New("Expected ')'")
	}
	return buffer, nil
}

func evalTerminator(node *tokenizer.TokenTreeNode, buffer string, state *state) (string, error) {
	if len(node.TokenType) > 1 && node.TokenType[1] == "StmtTm" {
		if node.Val == "EOF" {
			evalStmt(node, buffer, state)
		}
		return evalStmt(node.Right, buffer, state)
	}
	return "", errors.New("Invalid Terminator: " + node.Val)
}
