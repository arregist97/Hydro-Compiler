package tokenizer

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"unicode/utf8"
)

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

func BuildTokenTree(tokens []string) *TokenTreeNode {
	if len(tokens) <= 0 {
		return &TokenTreeNode{Val: "EOF", TokenType: []string {"Stmt", "StmtTm"}}
	}
	val := tokens[0]
	tokenType, err := validateToken(val)
	if err != nil {
		log.Fatal("Error building token tree: ", err)
	}
	node := TokenTreeNode{Val: val, TokenType: tokenType}
	var offset int = 0
	if node.Val == "(" {
		expr, dist := buildExpr(tokens[1:], offset)
		node.Left = expr
		offset = dist
	}
	tree := BuildTokenTree(tokens[1+offset:])
	node.Right = tree
	return &node
}

func buildExpr(tokens []string, dist int) (*TokenTreeNode, int) {
	if len(tokens) <= 0 {
		log.Fatal("Unexpected end of file.")
	}
	val := tokens[0]
	tokenType, err := validateToken(val)
	if err != nil {
		log.Fatal("Error building token tree: ", err)
	}
	node := TokenTreeNode{Val: val, TokenType: tokenType}
	dist++
	var offset int = 0
	if node.Val == "(" {
		expr, exprDist := buildExpr(tokens[1:], offset)
		node.Left = expr
		offset = exprDist
		dist = dist + exprDist
	} else if node.Val == ")" {
		return &node, dist
	}
	tree, dist := buildExpr(tokens[1+offset:], dist)
	node.Right = tree
	return &node, dist
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
