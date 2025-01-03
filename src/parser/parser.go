package parser

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/arregist97/Hydro-Compiler/tokenizer"
)

type NodeStore struct {
	I     int
	Block *nodeBlock
}

func (n *NodeStore) AddNode(token *tokenizer.Token, tokenType []string) {
	store := n.Block
	store.addNode(token, tokenType, n.I)
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
		I:     0,
	}
}

type nodeBlock struct {
	nodes *[]TokenTreeNode
	next  *nodeBlock
}

func newNodeBlock() *nodeBlock {
	nodes := make([]TokenTreeNode, 100)
	return &nodeBlock{
		nodes: &nodes,
	}
}

func (n *nodeBlock) addNode(token *tokenizer.Token, tokenType []string, index int) {
	nodes := *n.nodes
	if index < cap(nodes) {
		nodes[index].TokenType = tokenType
		nodes[index].Token = token
	} else if n.next != nil {
		n.next.addNode(token, tokenType, index-cap(nodes))
	} else {
		newBlock := newNodeBlock()
		n.next = newBlock
		n.next.addNode(token, tokenType, index-cap(nodes))
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
		n.next.linkNodes(j-cap(nodes), right, next)
	} else {
		fmt.Printf("Linking %s, %s to %s, %s\n", next.TokenType, next.Token.Val, nodes[j].TokenType, nodes[j].Token.Val)
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
	Token     *tokenizer.Token
	TokenType []string
	Left      *TokenTreeNode
	Right     *TokenTreeNode
	Root      *TokenTreeNode
}

func (node *TokenTreeNode) PrintTokenTree() {
	fmt.Println("Type: ", node.TokenType)
	fmt.Println("Val: ")
	node.Token.Print()
	if node.Left != nil {
		fmt.Println("{")
		node.Left.PrintTokenTree()
		fmt.Println("}")
	}
	if node.Right != nil {
		node.Right.PrintTokenTree()
	}
}

func BuildTokenTree(store *NodeStore, tokens []*tokenizer.Token, inScope bool) *TokenTreeNode {
	if len(tokens) <= 0 {
		return nil
	}
	token := tokens[0]
	tokenType, err := validateToken(token)
	if err != nil {
		log.Fatal("Error building token tree: ", err)
	}
	var nodeI int = store.I
	store.AddNode(token, tokenType)
	fmt.Println("Printing Token")
	node := store.GetNode(nodeI)
	node.PrintTokenTree()
	if node.Token.Val == "{" {
		fmt.Println("Entering Scope")
		scope := BuildTokenTree(store, tokens[1:], true)
		store.LinkNodes(nodeI, false, scope)
	} else if node.Token.Val == "}" {
		if !inScope {
			log.Fatal("Out of scope")
		}
		return node
	} else if node.Token.Val == "(" {
		fmt.Println("Entering Expr paren")
		expr, _ := constructExpr(store, tokens[1:], true, 0)
		store.LinkNodes(nodeI, false, expr)
	} else if stringInSlice(node.Token.Val, []string{"exit", "=", "if", "elif"}) {
		fmt.Println("Entering Expr no paren")
		expr, _ := constructExpr(store, tokens[1:], false, 0)
		store.LinkNodes(nodeI, false, expr)
	}
	offset := store.I - nodeI
	tokens = tokens[offset:]
	if stringInSlice(node.Token.Val, []string{"if", "elif", "else"}) {
		tokens = skipNewLine(tokens)
	}
	tree := BuildTokenTree(store, tokens, inScope)
	store.LinkNodes(nodeI, true, tree)
	return node
}

func skipNewLine(tokens []*tokenizer.Token) []*tokenizer.Token {
	fmt.Println("Checking if newline: '" + tokens[0].Val + "'")
	if len(tokens) <= 0 || tokens[0].Val != "\n" {
		return tokens
	}

	return tokens[1:]

}

func constructAtom(store *NodeStore, tokens []*tokenizer.Token, paren bool) *TokenTreeNode {
	if len(tokens) <= 0 {
		log.Fatal("Unexpected end of file.")
	}
	token := tokens[0]
	tokenType, err := validateToken(token)
	if err != nil {
		log.Fatal("Error building token tree: ", err)
	}
	if !paren && tokenType[0] != "Expr" {
		fmt.Println("Exiting Expr no paren")
		return nil
	}
	if len(tokenType) > 1 && tokenType[1] == "ExprOp" {
		fmt.Println("Exiting Expr detected ExprOp")
		return nil
	}
	if paren && token.Val == "\n" {
		fmt.Println("Skipping newline inside paren Expr")
		return constructAtom(store, tokens[1:], paren)
	}
	nodeI := store.I
	store.AddNode(token, tokenType)
	fmt.Println("Printing Token")
	node := store.GetNode(nodeI)
	node.PrintTokenTree()

	if node.Token.Val == "(" {
		fmt.Println("Entering Expr paren")
		expr, _ := constructExpr(store, tokens[1:], true, 0)
		store.LinkNodes(nodeI, false, expr)
	} else if node.Token.Val == ")" && paren {
		fmt.Println("Exiting Expr paren")
		return node
	}
	offset := store.I - nodeI
	tree := constructAtom(store, tokens[offset:], paren)
	store.LinkNodes(nodeI, true, tree)
	return node
}

func isBinExpr(tokens []*tokenizer.Token) bool {
	if len(tokens) <= 0 {
		log.Fatal("Unexpected end of file.")
	}
	token := tokens[0]
	tokenType, err := validateToken(token)
	fmt.Println("Is BinExpr: " + token.Val)
	fmt.Println(tokenType)
	if err != nil {
		log.Fatal("Error checking for binary expression: ", err)
	}
	if len(tokenType) > 1 && tokenType[1] == "ExprOp" {
		fmt.Println("True")
		return true
	}
	fmt.Println("False")
	return false

}

func constructExpr(store *NodeStore, tokens []*tokenizer.Token, paren bool, minPrec int) (*TokenTreeNode, []*tokenizer.Token) {
	baseI := store.I
	fmt.Println("Entering expr")
	fmt.Println("Paren: ", paren)
	expr := constructAtom(store, tokens, paren)
	offset := store.I - baseI
	if tokens[offset-1].Val == ")" {
		return expr, tokens[offset:]
	}
	tokens = tokens[offset:]
	firstI := true
	for {
		if !isBinExpr(tokens) {
			break
		}
		token := tokens[0]
		tokenType, _ := validateToken(token)
		currPrec := 0
		if token.Val == "*" || token.Val == "/" {
			currPrec = 1
		}
		if currPrec < minPrec {
			break
		}
		opI := store.I
		store.AddNode(token, tokenType)
		opNode := store.GetNode(opI)
		currPrec = currPrec + 1
		tokens = tokens[1:]
		expr2, updatedTokens := constructExpr(store, tokens, paren, currPrec)
		tokens = updatedTokens
		if firstI {
			store.LinkNodes(opI, false, expr)
			store.LinkNodes(opI, true, expr2)
		} else {
			store.LinkNodes(opI, true, expr)
			store.LinkNodes(opI, false, expr2)
		}
		expr = opNode
		firstI = false
	}
	return expr, tokens
}

func validateToken(token *tokenizer.Token) ([]string, error) {
	var statements = []string{"exit", "let", "if"}
	var ifPreds = []string{"elif", "else"}
	var expressionOperators = []string{"+", "*", "-", "/"}
	var paren = []string{"(", ")"}
	var statementTerminators = []string{"\n", ";", "EOF"}
	var digitCheck = regexp.MustCompile(`^[0-9]+$`)
	var varCheck = regexp.MustCompile(`\b[_a-zA-Z][_a-zA-Z0-9]*\b`)

	token.Print()

	if stringInSlice(token.Val, expressionOperators) {
		return []string{"Expr", "ExprOp"}, nil
	}
	if stringInSlice(token.Val, ifPreds) {
		return []string{"ifPred"}, nil
	}
	if token.Val == "=" {
		return []string{"Stmt", "StmtOp"}, nil
	}
	if stringInSlice(token.Val, statementTerminators) {
		return []string{"Stmt", "StmtTm"}, nil
	}
	if stringInSlice(token.Val, statements) {
		return []string{"Stmt"}, nil
	}
	if stringInSlice(token.Val, paren) {
		return []string{"Expr"}, nil
	}
	if token.Val == "{" {
		return []string{"Stmt", "Scope"}, nil
	}
	if token.Val == "}" {
		return []string{"Stmt", "ScopeTm"}, nil
	}
	if digitCheck.MatchString(token.Val) {
		return []string{"Expr", "Term", "intLit"}, nil
	}
	if varCheck.MatchString(token.Val) {
		return []string{"Expr", "Term", "ident"}, nil
	}
	return []string{}, errors.New("Unable to identify token: `" + token.Val + "`")
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
