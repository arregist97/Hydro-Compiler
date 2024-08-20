package tokenizer

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"unicode/utf8"
)

type TokenTreeBlock struct {
	Block *[]TokenTreeNode
	I int
	Next *TokenTreeBlock
}

func NewTokenTreeBlock() *TokenTreeBlock {
	block := make([]TokenTreeNode, 100)
	return &TokenTreeBlock{
		Block: &block,
		I: 0,
	}
}

func AddTokenTreeNode(treeBlock *TokenTreeBlock, tokenType []string, val string) {
	block := *treeBlock.Block
	if treeBlock.I < cap(block) {
		block[treeBlock.I].TokenType = tokenType
		block[treeBlock.I].Val = val
		treeBlock.I++
	} else if treeBlock.Next != nil {
		AddTokenTreeNode(treeBlock.Next, tokenType, val)
	} else {
		newTree := NewTokenTreeBlock()
		treeBlock.Next = newTree
		AddTokenTreeNode(newTree, tokenType, val)
	}
}

func LinkTokenTreeNode(treeBlock *TokenTreeBlock, j int, right bool, next *TokenTreeNode) {
	if next == nil {
		return
	}
	block := *treeBlock.Block
	if j > cap(block) {
		if treeBlock.Next == nil {
			log.Fatal("block overflow")
		}
		LinkTokenTreeNode(treeBlock.Next, j - cap(block), right, next)
	} else {
		fmt.Printf("Linking %s, %s to %s, %s\n", next.TokenType, next.Val, block[j].TokenType, block[j].Val)
		if right {
			fmt.Println("Right link")
			block[j].Right = next
		} else {
			fmt.Println("Left link")
			block[j].Left = next
		}
	}
}

type TokenTreeNode struct {
	Val string
	TokenType []string
	Left *TokenTreeNode
	Right *TokenTreeNode
}

func PrintTokenTree (node *TokenTreeNode) {
	fmt.Println("Type: ", node.TokenType)
	fmt.Println("Val: ", node.Val)
	if node.Left != nil {
		fmt.Println("{")
		PrintTokenTree(node.Left)
		fmt.Println("}")
	}
	if node.Right != nil {
		PrintTokenTree(node.Right)
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
	var endOfTokenRunes = [...]rune {'(', ')', ' ', '\n'}
	for _, b := range endOfTokenRunes {
		if b == a {
			return true
		}
	}
	return false
}

func BuildTokenTree(block *TokenTreeBlock, tokens []string) *TokenTreeNode {
	blockNodes := *block.Block
	if len(tokens) <= 0 {
		nodeI := block.I
		AddTokenTreeNode(block, []string {"Stmt", "StmtTm"}, "EOF")
		return &blockNodes[nodeI]
	}
	val := tokens[0]
	tokenType, err := validateToken(val)
	if err != nil {
		log.Fatal("Error building token tree: ", err)
	}
	var nodeI int = block.I
	AddTokenTreeNode(block, tokenType, val)
	fmt.Println("Printing Token")
	PrintTokenTree(&blockNodes[nodeI])
	if blockNodes[nodeI].Val == "(" {
		fmt.Println("Entering Expr paren")
		expr := buildExpr(block, tokens[1:], true)
		LinkTokenTreeNode(block, nodeI, false, expr)
	} else if stringInSlice(blockNodes[nodeI].Val, []string{"exit", "="}) {
		fmt.Println("Entering Expr no paren")
		expr := buildExpr(block, tokens[1:], false)
		LinkTokenTreeNode(block, nodeI, false, expr)
	}
	offset := block.I - nodeI
	tree := BuildTokenTree(block, tokens[offset:])
	LinkTokenTreeNode(block, nodeI, true, tree)
	return &blockNodes[nodeI]
}

func buildExpr(block *TokenTreeBlock, tokens []string, paren bool) *TokenTreeNode {
	if len(tokens) <= 0 {
		log.Fatal("Unexpected end of file.")
	}
	val := tokens[0]
	tokenType, err := validateToken(val)
	if err != nil {
		log.Fatal("Error building token tree: ", err)
	}
	if !paren && len(tokenType) > 1 && tokenType[1] == "StmtTm" {
		fmt.Print("Exiting Expr no paren")
		return nil
	}
	nodeI := block.I
	AddTokenTreeNode(block, tokenType, val)
	fmt.Println("Printing Token")
	blockNodes := *block.Block
	PrintTokenTree(&blockNodes[nodeI])

	if blockNodes[nodeI].Val == "(" {
		fmt.Println("Entering Expr paren")
		expr := buildExpr(block, tokens[1:], true)
		LinkTokenTreeNode(block, nodeI, false, expr)
	} else if blockNodes[nodeI].Val == ")" && paren {
		fmt.Println("Exiting Expr paren")
		return &blockNodes[nodeI]
	}
	offset := block.I - nodeI
	tree := buildExpr(block, tokens[offset:], paren)
	LinkTokenTreeNode(block, nodeI, true, tree)
	return &blockNodes[nodeI]
}

func validateToken(token string) ([]string, error) {
	var statements = []string {"exit", "let"}
	var expressionOperators = []string {"(", ")"}
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
