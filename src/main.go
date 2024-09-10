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
	store := tokenizer.NewNodeStore()
	tree := tokenizer.BuildTokenTree(store, tokens)
	fmt.Println("\nToken Tree:")
	tree.PrintTokenTree()

	buffer, err := parseTree(tree)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buffer)

	fileName = filepath.Base(fileName)
	re := regexp.MustCompile(`\.[^.]+$`)
	baseName := re.ReplaceAllString(fileName, "")
	directory := "../build/"
	newFileName := baseName + ".asm"
	buildPath := directory + newFileName

	newFile, err := os.Create(buildPath)
	if err != nil {
		fmt.Println("failed to create new file: ", err)
	}
	defer newFile.Close()

	_, err = newFile.WriteString(buffer)
	if err != nil {
		fmt.Println("failed to write to new file: ", err)
	}

	oFileName := baseName + ".o"
	fmt.Println("nasm -felf64", newFileName)
	nasmCmd := exec.Command("nasm", "-felf64", newFileName)

	nasmCmd.Dir = "../build"

	nasmCmd.Stdout = os.Stdout
	nasmCmd.Stderr = os.Stderr

	// Run the nasm command
	err = nasmCmd.Run()
	if err != nil {
		log.Fatalf("nasm command execution failed: %v", err)
	}

	// Step 2: Run ld command
	fmt.Println("Running ld test.o -o test")

	// Create the ld command
	ldCmd := exec.Command("ld", oFileName, "-o", baseName)
	ldCmd.Dir = "../build"
	ldCmd.Stdout = os.Stdout
	ldCmd.Stderr = os.Stderr

	// Run the ld command
	err = ldCmd.Run()
	if err != nil {
		log.Fatalf("ld command execution failed: %v", err)
	}

	fmt.Println("Successfully assembled and linked the program.")

}

type state struct {
	stackPtr int
	context []map[string]int
	scopeI int
}

func (s *state) enterScope() {
	newScope := make(map[string]int)
	s.scopeI++
	s.context = append(s.context, newScope)
	fmt.Println("Enter new scope")
	fmt.Println(s.context)
}

func (s *state) exitScope() {
	s.context = s.context[:s.scopeI]
	s.scopeI--
	fmt.Println("Exit scope")
	fmt.Println(s.context)
}

func (s *state) decVar(val string) {
	scope := s.context[s.scopeI]
	scope[val] = s.stackPtr
}

func (s *state) getVar(val string) (int, error) {
	var scope map[string]int
	var stackLoc int
	var validIdent bool

	fmt.Println("Retrieving var value")
	fmt.Println(s.context)
	for i := s.scopeI; i >= 0; i-- {
		scope = s.context[i]
		stackLoc, validIdent = scope[val]
		if validIdent {
			break
		}
	}
	if !validIdent {
		return 0, errors.New("Undeclared ident " + val)
	}
	return stackLoc, nil
}

func newState() state {
	scope := make(map[string]int)
	context := make([]map[string]int, 1)
	context[0] = scope
	s := state{ stackPtr: 0, context: context, scopeI: 0 }
	fmt.Println("State create")
	fmt.Println(s.context)
	return s
}

func parseTree(node *tokenizer.TokenTreeNode) (string, error) {
	var buffer string
	buffer = "global _start"
	buffer = buffer + "\n" + "_start:"
	state := newState()
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
	} else if node.Val == "{" {
		state.enterScope()
	} else if node.Val == "}" {
		state.exitScope()
	} else {
		return "", errors.New("Undefined Stmt: " + node.Val)
	}
	return evalTerminator(node.Right, buffer, state) 
}

func evalExit(node *tokenizer.TokenTreeNode, buffer string, state *state) (string, error) {
	if node.Val != "(" {
		return "", errors.New("Expected `(` after exit")
	}
	buffer, err := evalExpr(node, buffer, state, false, 0)
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
	buf, err := evalExpr(node.Right.Left, buffer, state, false, 0)
	if err != nil {
		return "", err
	}
	buffer = buf

	state.decVar(node.Val)
	return buffer, nil
}

func evalExpr(node *tokenizer.TokenTreeNode, buffer string, state *state, paren bool, prec int) (string, error) {
	if node.TokenType[0] != "Expr" {
		fmt.Println("Node val: " + node.Val)
		return "", errors.New("Expr expected, recieved " + node.TokenType[0])
	}
	if node.Val == "(" {
		return evalExpr(node.Left, buffer, state, true, 0)
	}
	if node.TokenType[1] == "Term" {
		return evalTerm(node, buffer, state, paren)
	}
	if node.TokenType[1] == "StkVr" {
		return buffer, nil
	}
	if node.TokenType[1] == "ExprOp" {
		return evalBinExpr(node, buffer, state, paren, prec)
	}
	return "", errors.New("Invalid Expr: " + node.TokenType[1])
}

func evalBinExpr(node *tokenizer.TokenTreeNode, buffer string, state *state, paren bool, prec int) (string, error) {
	var err error
	buffer, err = evalExpr(node.Left, buffer, state, false, 0)
	if err != nil {
		return "", err
	}
	currNode := node
	for true {
		currPrec := 0
		op := currNode.Val
		if op == "*" || op == "/" {
			currPrec = 1
		}
		fmt.Println("Printing Tree ...")
		root := currNode
		for root.Root.TokenType[0] == "Expr" {
			root = root.Root
		}
		root.PrintTokenTree()
		if len(currNode.TokenType) <= 1 || currNode.TokenType[1] != "ExprOp" {
			break
		}
		if currPrec < prec {
			break
		}

		currPrec++

		buffer, err = evalExpr(currNode.Right, buffer, state, paren, currPrec)
		if err != nil {
			return "", err
		}
		if op == "+" {
			buffer = buffer + "\n" + "  pop    rbx"
			buffer = buffer + "\n" + "  pop    rax"
			buffer = buffer + "\n" + "  add    rax, rbx"
			buffer = buffer + "\n" + "  push   rax"
			state.stackPtr--

		} else if op == "*" {
			buffer = buffer + "\n" + "  pop    rbx"
			buffer = buffer + "\n" + "  pop    rax"
			buffer = buffer + "\n" + "  mul    rbx"
			buffer = buffer + "\n" + "  push   rax"
			state.stackPtr--
		} else if op == "-" {
			buffer = buffer + "\n" + "  pop    rbx"
			buffer = buffer + "\n" + "  pop    rax"
			buffer = buffer + "\n" + "  sub    rax, rbx"
			buffer = buffer + "\n" + "  push   rax"
			state.stackPtr--
		} else if op == "/" {
			buffer = buffer + "\n" + "  pop    rbx"
			buffer = buffer + "\n" + "  pop    rax"
			buffer = buffer + "\n" + "  div    rbx"
			buffer = buffer + "\n" + "  push   rax"
			state.stackPtr--
		}else {
			return "", errors.New("Invalid BinExpr: " + op)
		}

		err = tokenizer.ConsumeOperation(currNode)
		if err != nil {
			return "", err
		}

		currNode = currNode.Right
		if currNode != nil {
			fmt.Println("New currNode: ", currNode.Val)
		} else {
			fmt.Println("Null val for currNode")
		}

	}
	
	
	return buffer, nil
}

func evalTerm(node *tokenizer.TokenTreeNode, buffer string, state *state, paren bool) (string, error) {
	if node.TokenType[2] == "intLit" {
		buffer = buffer + "\n" + "  mov    rax, " + node.Val
		buffer = buffer + "\n" + "  push   rax"
		state.stackPtr++
	} else if node.TokenType[2] == "ident" {
		stackLoc, err := state.getVar(node.Val)
		if err != nil {
			return "", err
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
	if node.Val == "}" {
		state.exitScope()
		return evalTerminator(node.Right, buffer, state)
	}
	return "", errors.New("Invalid Terminator: " + node.Val)
}
