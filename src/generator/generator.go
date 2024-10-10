package generator

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/arregist97/Hydro-Compiler/parser"
)

type state struct {
	stackPtr int
	context []map[string]int
	scopeI int
	labelI int
}

func (s *state) enterScope(node *parser.TokenTreeNode, buffer string) (string, error) {
	newScope := make(map[string]int)
	s.scopeI++
	s.context = append(s.context, newScope)
	s.decVar("{")
	fmt.Println("Enter new scope")
	fmt.Println(s.context)

	buf, err := evalTerminator(node.Left, buffer, s)
	if err != nil {
		return "", err
	}
	buffer = buf
	return buffer, nil
}

func (s *state) exitScope(buffer string) (string, error) {
	scopeStkPtr, err := s.getVar("{")
	if err != nil {
		return "", errors.New("No scope to exit")
	}
	s.context = s.context[:s.scopeI]
	s.scopeI--
	fmt.Println("Exit scope")
	fmt.Println(s.context)
	stackDiff := s.stackPtr - scopeStkPtr
	if stackDiff > 0 {
		buffer = buffer + "\n" + "  add    rsp, " + strconv.Itoa(stackDiff * 8)
		s.stackPtr = scopeStkPtr
	}
	return buffer, nil
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
	s := state{ stackPtr: 0, context: context, scopeI: 0, labelI: 0 }
	fmt.Println("State create")
	fmt.Println(s.context)
	return s
}

func Generate(node *parser.TokenTreeNode) (string, error) {
	var buffer string
	buffer = "global _start"
	buffer = buffer + "\n" + "_start:"
	state := newState()
	buffer, err := evalStmt(node, buffer, &state)
	return buffer, err
}

func evalStmt(node *parser.TokenTreeNode, buffer string, state *state) (string, error) {
	fmt.Println("Evaluating statement " + node.Val + "...")
	if node.TokenType[0] != "Stmt" {
		return "", errors.New("Stmt expected. Recieved " + node.TokenType[0])
	}
	if node.Val == "EOF" {
		fmt.Println("Test")
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
	} else if node.Val == "if" {
		buf, nd, err := evalIf(node, buffer, state)
		if err != nil {
			return "", err
		}
		buffer = buf
		node = nd
		fmt.Println("Exit if, node: " + node.Val)
	} else if node.Val == "{" {
		//enterScope needs to pass buffer and state back to evalStmt
		buf, err := state.enterScope(node, buffer)
		if err != nil {
			return "", err
		}
		buffer = buf
	} else if node.Val == "}" {
		buffer, err := state.exitScope(buffer)
		if err != nil {
			return "", err
		}
		return buffer, nil
	} else {
		return "", errors.New("Undefined Stmt: " + node.Val)
	}
	return evalTerminator(node.Right, buffer, state) 
}

func evalExit(node *parser.TokenTreeNode, buffer string, state *state) (string, error) {
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

func evalLet(node *parser.TokenTreeNode, buffer string, state *state) (string, error) {
	if len(node.TokenType) > 2 && node.TokenType[2] != "ident" {
		log.Fatal("Improper declaration")
	}
	if node.Right.Val != "=" {
		log.Fatal("Expected '='")
	}
	buf, err := evalExpr(node.Right.Left, buffer, state, false)
	if err != nil {
		return "", err
	}
	buffer = buf

	state.decVar(node.Val)
	return buffer, nil
}

func evalIf(node *parser.TokenTreeNode, buffer string, state *state) (string, *parser.TokenTreeNode, error) {
	buffer, err := evalExpr(node.Left, buffer, state, false)
	if err != nil {
		return "", nil, err
	}
	label := "label" + strconv.Itoa(state.labelI)
	state.labelI++
	buffer = buffer + "\n" + "  pop    rax"
	buffer = buffer + "\n" + "  test   rax, rax"
	buffer = buffer + "\n" + "  jz     " + label
	state.stackPtr--

	node = node.Right
	if node.Val != "{" {
		return "", nil, errors.New("Expected Scope.")
	}
	buffer, err = state.enterScope(node, buffer)
	if err != nil {
		return "", nil, err
	}

	node = node.Right
	if node.TokenType[0] != "ifPred"{
		buffer = buffer + "\n" + label + ":"
		return buffer, node, nil
	}
	
	endLabel := "label" + strconv.Itoa(state.labelI)
	state.labelI++
	for node.Val == "elif" {
		buffer = buffer + "\n" + "  jmp    " + endLabel
		buffer = buffer + "\n" + label + ":"
		
		buffer, err = evalExpr(node.Left, buffer, state, false)
		if err != nil {
			return "", nil, err
		}
		
		label = "label" + strconv.Itoa(state.labelI)
		state.labelI++
		buffer = buffer + "\n" + "  pop    rax"
		buffer = buffer + "\n" + "  test   rax, rax"
		buffer = buffer + "\n" + "  jz     " + label
		state.stackPtr--
		
		node = node.Right
		if node.Val != "{" {
			return "", nil, errors.New("Expected Scope.")
		}

		buffer, err = state.enterScope(node, buffer)
		if err != nil {
			return "", nil, err
		}

		node = node.Right
	}

	if node.Val == "else" {
		buffer = buffer + "\n" + "  jmp    " + endLabel
		buffer = buffer + "\n" + label + ":"
		node = node.Right
		if node.Val != "{" {
			return "", nil, errors.New("Expected Scope.")
		}
		buffer, err = state.enterScope(node, buffer)
		if err != nil {
			return "", nil, err
		}
	} else {
		buffer = buffer + "\n" + label + ":"
	}

	buffer = buffer + "\n" + endLabel + ":"

	return buffer, node, nil
}

func evalExpr(node *parser.TokenTreeNode, buffer string, state *state, paren bool) (string, error) {
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
	if node.TokenType[1] == "StkVr" {
		return buffer, nil
	}
	if node.TokenType[1] == "ExprOp" {
		return evalBinExpr(node, buffer, state, paren)
	}
	return "", errors.New("Invalid Expr: " + node.TokenType[1])
}

func evalBinExpr(node *parser.TokenTreeNode, buffer string, state *state, paren bool) (string, error) {
	var err error
	buffer, err = evalExpr(node.Left, buffer, state, false)
	if err != nil {
		return "", err
	}
	buffer, err = evalExpr(node.Right, buffer, state, paren)
	if err != nil {
		return "", err
	}
	buffer = buffer + "\n" + "  pop    rax"
	buffer = buffer + "\n" + "  pop    rbx"
	if node.Val == "+" {
		buffer = buffer + "\n" + "  add    rax, rbx"
	} else if node.Val == "*" {
		buffer = buffer + "\n" + "  mul    rbx"
	} else if node.Val == "-" {
		buffer = buffer + "\n" + "  sub    rax, rbx"
	} else if node.Val == "/" {
		buffer = buffer + "\n" + "  div    rbx"
	}else {
		return "", errors.New("Invalid BinExpr: " + node.Val)
	}
	buffer = buffer + "\n" + "  push   rax"
	state.stackPtr--
	
	return buffer, nil
}

func evalTerm(node *parser.TokenTreeNode, buffer string, state *state, paren bool) (string, error) {
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

func evalTerminator(node *parser.TokenTreeNode, buffer string, state *state) (string, error) {
	fmt.Println("Evaluating terminator: " + node.Val)
	if len(node.TokenType) > 1 && node.TokenType[1] == "StmtTm" {
		if node.Val == "EOF" {
			return evalStmt(node, buffer, state)
		}
		return evalStmt(node.Right, buffer, state)
	}
	if node.Val == "}" {
		buf, err := state.exitScope(buffer)
		if err != nil {
			return "", err
		}
		buffer = buf
		return evalTerminator(node.Right, buffer, state)
	}
	return "", errors.New("Invalid Terminator: " + node.Val)
}
