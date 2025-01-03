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
	context  []map[string]int
	scopeI   int
	labelI   int
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
		return "", errors.New("no scope to exit")
	}
	s.context = s.context[:s.scopeI]
	s.scopeI--
	fmt.Println("Exit scope")
	fmt.Println(s.context)
	stackDiff := s.stackPtr - scopeStkPtr
	if stackDiff > 0 {
		buffer = buffer + "\n" + "  add    rsp, " + strconv.Itoa(stackDiff*8)
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
		return 0, errors.New("undeclared ident " + val)
	}
	return stackLoc, nil
}

func newState() state {
	scope := make(map[string]int)
	context := make([]map[string]int, 1)
	context[0] = scope
	s := state{stackPtr: 0, context: context, scopeI: 0, labelI: 0}
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
	fmt.Println("Evaluating statement " + node.Token.Val + "...")
	if node.TokenType[0] != "Stmt" {
		return "", errors.New("statement expected, recieved " + node.TokenType[0])
	}
	if node.Token.Val == "EOF" {
		fmt.Println("Test")
		buffer = buffer + "\n" + "  mov    rax, 60"
		buffer = buffer + "\n" + "  mov    rdi, 0"
		buffer = buffer + "\n" + "  syscall"
		return buffer, nil
	}
	if len(node.TokenType) > 1 && node.TokenType[1] == "StmtTm" {
		return evalStmt(node.Right, buffer, state)
	}
	if node.Token.Val == "exit" {
		buf, err := evalExit(node.Left, buffer, state)
		if err != nil {
			return "", err
		}
		buffer = buf

	} else if node.Token.Val == "let" {
		buf, err := evalLet(node.Right, buffer, state)
		if err != nil {
			return "", err
		}
		buffer = buf

		node = node.Right.Right
	} else if node.Token.Val == "if" {
		buf, nd, err := evalIf(node, buffer, state)
		if err != nil {
			return "", err
		}
		buffer = buf
		node = nd
		fmt.Println("Exiting if, node: " + node.Token.Val)
	} else if node.Token.Val == "{" {
		buf, err := state.enterScope(node, buffer)
		if err != nil {
			return "", err
		}
		buffer = buf
	} else if node.Token.Val == "}" {
		buffer, err := state.exitScope(buffer)
		if err != nil {
			return "", err
		}
		return buffer, nil
	} else {
		return "", errors.New("undefined Stmt: " + node.Token.Val)
	}
	return evalTerminator(node.Right, buffer, state)
}

func evalExit(node *parser.TokenTreeNode, buffer string, state *state) (string, error) {
	if node.Token.Val != "(" {
		return "", errors.New("expected `(` after exit")
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
	if node.Right.Token.Val != "=" {
		log.Fatal("Expected '='")
	}
	buf, err := evalExpr(node.Right.Left, buffer, state, false)
	if err != nil {
		return "", err
	}
	buffer = buf

	state.decVar(node.Token.Val)
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
	if node.Token.Val != "{" {
		return "", nil, errors.New("expected scope")
	}
	buffer, err = state.enterScope(node, buffer)
	if err != nil {
		return "", nil, err
	}

	next := node.Right
	if next.TokenType[0] != "ifPred" {
		buffer = buffer + "\n" + label + ":"
		return buffer, node, nil
	}

	endLabel := "label" + strconv.Itoa(state.labelI)
	state.labelI++
	for next.Token.Val == "elif" {
		node = next
		next = node.Right

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

		if next.Token.Val != "{" {
			return "", nil, errors.New("expected scope")
		}
		node = next
		next = node.Right

		buffer, err = state.enterScope(node, buffer)
		if err != nil {
			return "", nil, err
		}
	}

	if next.Token.Val == "else" {
		node = next
		next = node.Right

		buffer = buffer + "\n" + "  jmp    " + endLabel
		buffer = buffer + "\n" + label + ":"

		if next.Token.Val != "{" {
			return "", nil, errors.New("expected scope")
		}
		node = next
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
		fmt.Println("Node val: " + node.Token.Val)
		return "", errors.New("expression expected, recieved " + node.TokenType[0])
	}
	if node.Token.Val == "(" {
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
	return "", errors.New("invalid expression: " + node.TokenType[1])
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
	if node.Token.Val == "+" {
		buffer = buffer + "\n" + "  add    rax, rbx"
		buffer = buffer + "\n" + "  push   rax"
	} else if node.Token.Val == "*" {
		buffer = buffer + "\n" + "  mul    rbx"
		buffer = buffer + "\n" + "  push   rax"
	} else if node.Token.Val == "-" {
		buffer = buffer + "\n" + "  sub    rbx, rax"
		buffer = buffer + "\n" + "  push   rbx"
	} else if node.Token.Val == "/" {
		buffer = buffer + "\n" + "  div    rbx"
		buffer = buffer + "\n" + "  push   rax"
	} else {
		return "", errors.New("invalid binary expression: " + node.Token.Val)
	}
	state.stackPtr--

	return buffer, nil
}

func evalTerm(node *parser.TokenTreeNode, buffer string, state *state, paren bool) (string, error) {
	if node.TokenType[2] == "intLit" {
		buffer = buffer + "\n" + "  mov    rax, " + node.Token.Val
		buffer = buffer + "\n" + "  push   rax"
		state.stackPtr++
	} else if node.TokenType[2] == "ident" {
		stackLoc, err := state.getVar(node.Token.Val)
		if err != nil {
			return "", err
		}

		fmt.Println("Stack Pointer", state.stackPtr, "var location", stackLoc)
		stackOffset := (state.stackPtr - stackLoc) * 8
		fmt.Println(stackOffset)
		buffer = buffer + "\n" + "  push   QWORD [rsp + " + strconv.Itoa(stackOffset) + "]"
		state.stackPtr++
	} else {
		return "", errors.New("invalid term: " + node.TokenType[2])
	}
	if paren && (node.Right == nil || node.Right.Token.Val != ")") {
		return "", errors.New("expected ')'")
	}
	return buffer, nil
}

func evalTerminator(node *parser.TokenTreeNode, buffer string, state *state) (string, error) {
	fmt.Println("Evaluating terminator: " + node.Token.Val)
	if len(node.TokenType) > 1 && node.TokenType[1] == "StmtTm" {
		if node.Token.Val == "EOF" {
			return evalStmt(node, buffer, state)
		}
		return evalStmt(node.Right, buffer, state)
	}
	if node.Token.Val == "}" {
		buf, err := state.exitScope(buffer)
		if err != nil {
			return "", err
		}
		buffer = buf
		return evalTerminator(node.Right, buffer, state)
	}
	return "", errors.New("invalid terminator: " + node.Token.Val)
}
