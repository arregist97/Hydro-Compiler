package tokenizer

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"unicode/utf8"
)

type NodeStore struct {
	I int
	Block *nodeBlock
}

func (n *NodeStore) AddNode(tokenType []string, val string) {
	store := n.Block
	store.addNode(tokenType, val, n.I)
	n.I++
}

func (n *NodeStore) GetNode(index int) *TokenTreeNode {
	store := n.Block
	return store.getNode(index)
}

func (n *NodeStore) LinkNodes(j int, right bool, next *TokenTreeNode) {
	if next == nil {
		return
	}
	store := n.Block
	store.linkNodes(j, right, next)
}

func NewNodeStore() *NodeStore {
	return &NodeStore{
		Block: newNodeBlock(),
		I: 0,
	}
}

type nodeBlock struct {
	nodes *[]TokenTreeNode
	next *nodeBlock
}

func newNodeBlock() *nodeBlock {
	nodes := make([]TokenTreeNode, 100)
	return &nodeBlock{
		nodes: &nodes,
	}
}

func (n *nodeBlock) addNode(tokenType []string, val string, index int) {
	nodes := *n.nodes
	if index < cap(nodes) {
		nodes[index].TokenType = tokenType
		nodes[index].Val = val
	} else if n.next != nil {
		n.next.addNode(tokenType, val, index - cap(nodes))
	} else {
		newBlock := newNodeBlock()
		n.next = newBlock
		n.next.addNode(tokenType, val, index - cap(nodes))
	}
}

func (n *nodeBlock) getNode(index int) *TokenTreeNode {
	nodes := *n.nodes
	if index > cap(nodes) {
		if n.next == nil {
			log.Fatal("index overflow")
		}
		return n.next.getNode(index - cap(nodes))
	} else {
		return &nodes[index]
	}
}

func (n *nodeBlock) linkNodes(j int, right bool, next *TokenTreeNode) {
	nodes := *n.nodes
	if j > cap(nodes) {
		if n.next == nil {
			log.Fatal("index overflow")
		}
		n.next.linkNodes(j - cap(nodes), right, next)
	} else {
		fmt.Printf("Linking %s, %s to %s, %s\n", next.TokenType, next.Val, nodes[j].TokenType, nodes[j].Val)
		if right {
			fmt.Println("Right link")
			nodes[j].Right = next
		} else {
			fmt.Println("Left link")
			nodes[j].Left = next
		}
		next.Root = &nodes[j]
	}
}

type TokenTreeNode struct {
	Val string
	TokenType []string
	Left *TokenTreeNode
	Right *TokenTreeNode
	Root *TokenTreeNode
}

func (node *TokenTreeNode) PrintTokenTree () {
	fmt.Println("Type: ", node.TokenType)
	fmt.Println("Val: ", node.Val)
	if node.Left != nil {
		fmt.Println("{")
		node.Left.PrintTokenTree()
		fmt.Println("}")
	}
	if node.Right != nil {
		node.Right.PrintTokenTree()
	}
}

func RecTokenize(content string, tokens []string) []string {
	var token string
	var updatedContent string
	var err error

	token, updatedContent, err = buildToken(content)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}
	fmt.Println(token + "/end")
	if len(token) > 0 {
		tokens = append(tokens, token)
	}
	if len(content) > 0 {
		return RecTokenize(updatedContent, tokens)
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
		if s == 1 {
			return "", "", errors.New("Could not recognize token " + content[i:i+1]) 
		} else {
			return "", "", errors.New("Empty decode error")
		}
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
	var endOfTokenRunes = [...]rune {'(', ')', ' ', '\n', '+', '*', '-', '/'}
	for _, b := range endOfTokenRunes {
		if b == a {
			return true
		}
	}
	return false
}

func BuildTokenTree(store *NodeStore, tokens []string) *TokenTreeNode {
	if len(tokens) <= 0 {
		nodeI := store.I
		store.AddNode([]string {"Stmt", "StmtTm"}, "EOF")
		return store.GetNode(nodeI)
	}
	val := tokens[0]
	tokenType, err := validateToken(val)
	if err != nil {
		log.Fatal("Error building token tree: ", err)
	}
	var nodeI int = store.I
	store.AddNode(tokenType, val)
	fmt.Println("Printing Token")
	node := store.GetNode(nodeI)
	node.PrintTokenTree()
	if node.Val == "(" {
		fmt.Println("Entering Expr paren")
		expr := buildExpr(store, tokens[1:], true)
		offset := store.I - nodeI
		if isBinExpr(store, tokens[offset:]) {
			opI := store.I - 1
			opNode := store.GetNode(opI)
			rhs := constructRhs(store, tokens[offset + 1:], true)
			store.LinkNodes(nodeI, false, opNode)
			store.LinkNodes(opI, false, expr)
			store.LinkNodes(opI, true, rhs)
		} else {
			store.LinkNodes(nodeI, false, expr)
		}
	} else if stringInSlice(node.Val, []string{"exit", "="}) {
		fmt.Println("Entering Expr no paren")
		expr := buildExpr(store, tokens[1:], false)
		offset := store.I - nodeI
		if isBinExpr(store, tokens[offset:]) {
			opI := store.I - 1
			opNode := store.GetNode(opI)
			rhs := constructRhs(store, tokens[offset + 1:], false)
			store.LinkNodes(nodeI, false, opNode)
			store.LinkNodes(opI, false, expr)
			store.LinkNodes(opI, true, rhs)
		} else {
			store.LinkNodes(nodeI, false, expr)
		}
	}
	offset := store.I - nodeI
	tree := BuildTokenTree(store, tokens[offset:])
	store.LinkNodes(nodeI, true, tree)
	return node
}

func buildExpr(store *NodeStore, tokens []string, paren bool) *TokenTreeNode {
	if len(tokens) <= 0 {
		log.Fatal("Unexpected end of file.")
	}
	val := tokens[0]
	tokenType, err := validateToken(val)
	if err != nil {
		log.Fatal("Error building token tree: ", err)
	}
	if !paren && len(tokenType) > 1 && tokenType[1] == "StmtTm" {
		fmt.Println("Exiting Expr no paren")
		return nil
	}
	if len(tokenType) > 1 && tokenType[1] == "ExprOp" {
		fmt.Println("Exiting Expr detected ExprOp")
		return nil
	}
	nodeI := store.I
	store.AddNode(tokenType, val)
	fmt.Println("Printing Token")
	node := store.GetNode(nodeI)
	node.PrintTokenTree()

	if node.Val == "(" {
		fmt.Println("Entering Expr paren")
		expr := buildExpr(store, tokens[1:], true)
		offset := store.I - nodeI
		if isBinExpr(store, tokens[offset:]) {
			opI := store.I - 1
			opNode := store.GetNode(opI)
			rhs := constructRhs(store, tokens[offset + 1:], true)
			store.LinkNodes(nodeI, false, opNode)
			store.LinkNodes(opI, false, expr)
			store.LinkNodes(opI, true, rhs)
		} else {
			store.LinkNodes(nodeI, false, expr)
		}
	} else if node.Val == ")" && paren {
		fmt.Println("Exiting Expr paren")
		return node
	}
	offset := store.I - nodeI
	tree := buildExpr(store, tokens[offset:], paren)
	store.LinkNodes(nodeI, true, tree)
	return node
}

func isBinExpr(store *NodeStore, tokens []string) bool {
	if len(tokens) <= 0 {
		log.Fatal("Unexpected end of file.")
	}
	val := tokens[0]
	tokenType, err := validateToken(val)
	fmt.Println("Is BinExpr: " + val)
	fmt.Println(tokenType)
	if err != nil {
		log.Fatal("Error checking for binary expression: ", err)
	}
	if len(tokenType) > 1 && tokenType[1] == "ExprOp" {
		fmt.Println("True")
		store.AddNode(tokenType, val)
		return true
	}
	fmt.Println("False")
	return false

}

func constructRhs(store *NodeStore, tokens []string, paren bool) *TokenTreeNode{
	baseI := store.I
	fmt.Println("Entering expr of binexpr")
	fmt.Println("Paren: ", paren)
	expr := buildExpr(store, tokens, paren)
	offset := store.I - baseI
	if tokens[offset - 1] == ")" {
		return expr
	}
	if isBinExpr(store, tokens[offset:]) {
		opI := store.I - 1
		opNode := store.GetNode(opI)
		rhs := constructRhs(store, tokens[offset + 1:], paren)
		store.LinkNodes(opI, false, expr)
		store.LinkNodes(opI, true, rhs)
		return opNode
	}
	return expr
}

func validateToken(token string) ([]string, error) {
	var statements = []string {"exit", "let"}
	var expressionOperators = []string {"+", "*", "-", "/"}
	var paren = []string {"(", ")"}
	var statementTerminators = []string {"\n", ";"}
	var digitCheck = regexp.MustCompile(`^[0-9]+$`)
	var varCheck = regexp.MustCompile(`\b[_a-zA-Z][_a-zA-Z0-9]*\b`)

	if stringInSlice(token, expressionOperators) {
		return []string{"Expr", "ExprOp"}, nil
	}
	if token == "=" {
		return []string{"Stmt", "StmtOp"}, nil
	}
	if stringInSlice(token, statementTerminators) {
		return []string{"Stmt", "StmtTm"}, nil
	}
	if stringInSlice(token, statements) {
		return []string{"Stmt"}, nil
	}
	if stringInSlice(token, paren) {
		return []string{"Expr"}, nil
	}
	if digitCheck.MatchString(token) {
		return []string{"Expr", "Term", "intLit"}, nil
	}
	if varCheck.MatchString(token) {
		return []string{"Expr", "Term", "ident"}, nil
	}
	return []string{}, errors.New("Unable to identify token: `" + token + "`")
}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func ConsumeOperation(op *TokenTreeNode) error {
	if op == nil {
		return errors.New("ConsumeOperation: nil pointer")
	}
	if len(op.TokenType) <= 1 {
		if op.TokenType[1] != "ExprOp" {
			return errors.New("ConsumeOperation: Expected ExprOp, received " + op.TokenType[1])
		}
		return errors.New("ConsumeOperation: Expected ExprOp, received " + op.TokenType[0])
	}
	fmt.Println("Consume Op: ", op.Val)
	fmt.Println("Op Tree:")
	op.PrintTokenTree()
	rhs := op.Right
	if rhs == nil || rhs.TokenType[0] != "Expr" {
		return errors.New("ConsumeOperation: Misconstructed binary operation")
	}
	val := "Stack Placeholder"
	tokenType := []string {"Expr", "StkVr"}
	if rhs.TokenType[1] == "ExprOp" {
		rhs.Left.Val = val
		rhs.Left.TokenType = tokenType
	} else {
		rhs.Val = val
		rhs.TokenType = tokenType
	}
	root := op.Root
	fmt.Println("Op Root:", root)
	if root == nil {
		return errors.New("ConsumeOperation: nil root")
	}
	if len(root.TokenType) > 1 && root.TokenType[1] == "ExprOp" {
		root.Right = rhs
		rhs.Root = root
	} else {
		root.Left = rhs
		rhs.Root = root
	}
	return nil
}
