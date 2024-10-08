package parser

import(
	"errors"
	"fmt"
	"log"
	"regexp"
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

func BuildTokenTree(store *NodeStore, tokens []string, inScope bool) *TokenTreeNode {
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
	if node.Val == "{" {
		fmt.Println("Entering Scope")
		scope := BuildTokenTree(store, tokens[1:], true)
		store.LinkNodes(nodeI, false, scope)
	} else if node.Val == "}" {
		if !inScope {
			log.Fatal("Out of scope")
		}
		return node
	} else if node.Val == "(" {
		fmt.Println("Entering Expr paren")
		expr, _ := constructExpr(store, tokens[1:], true, 0)
		store.LinkNodes(nodeI, false, expr)
	} else if stringInSlice(node.Val, []string{"exit", "=", "if", "elif"}) {
		fmt.Println("Entering Expr no paren")
		expr, _ := constructExpr(store, tokens[1:], false, 0)
		store.LinkNodes(nodeI, false, expr)
	}
	offset := store.I - nodeI
	tree := BuildTokenTree(store, tokens[offset:], inScope)
	store.LinkNodes(nodeI, true, tree)
	return node
}

func constructAtom(store *NodeStore, tokens []string, paren bool) *TokenTreeNode {
	if len(tokens) <= 0 {
		log.Fatal("Unexpected end of file.")
	}
	val := tokens[0]
	tokenType, err := validateToken(val)
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
	nodeI := store.I
	store.AddNode(tokenType, val)
	fmt.Println("Printing Token")
	node := store.GetNode(nodeI)
	node.PrintTokenTree()

	if node.Val == "(" {
		fmt.Println("Entering Expr paren")
		expr, _ := constructExpr(store, tokens[1:], true, 0)
		store.LinkNodes(nodeI, false, expr)
	} else if node.Val == ")" && paren {
		fmt.Println("Exiting Expr paren")
		return node
	}
	offset := store.I - nodeI
	tree := constructAtom(store, tokens[offset:], paren)
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
		return true
	}
	fmt.Println("False")
	return false

}

func constructExpr(store *NodeStore, tokens []string, paren bool, minPrec int) (*TokenTreeNode, []string){
	baseI := store.I
	fmt.Println("Entering expr of binexpr")
	fmt.Println("Paren: ", paren)
	expr := constructAtom(store, tokens, paren)
	offset := store.I - baseI
	if tokens[offset - 1] == ")" {
		return expr, tokens[offset:]
	}
	tokens = tokens[offset:]
	firstI := true
	for true {
		if !isBinExpr(store, tokens) {
			break
		}
		val := tokens[0]
		tokenType, _ := validateToken(val)
		currPrec := 0
		if val == "*" || val == "/" {
			currPrec = 1
		}
		if currPrec < minPrec {
			break
		}
		opI := store.I
		store.AddNode(tokenType, val)
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

func validateToken(token string) ([]string, error) {
	var statements = []string {"exit", "let", "if"}
	var ifPreds = []string {"elif", "else"}
	var expressionOperators = []string {"+", "*", "-", "/"}
	var paren = []string {"(", ")"}
	var statementTerminators = []string {"\n", ";"}
	var digitCheck = regexp.MustCompile(`^[0-9]+$`)
	var varCheck = regexp.MustCompile(`\b[_a-zA-Z][_a-zA-Z0-9]*\b`)

	if stringInSlice(token, expressionOperators) {
		return []string{"Expr", "ExprOp"}, nil
	}
	if stringInSlice(token, ifPreds) {
		return []string{"ifPred"}, nil
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
	if token == "{" {
		return []string{"Stmt", "Scope"}, nil
	}
	if token == "}" {
		return []string{"Stmt", "ScopeTm"}, nil
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
