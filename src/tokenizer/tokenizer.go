package tokenizer

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

type parseTreeNode struct {
	val string
	left *parseTreeNode
	right *parseTreeNode
}

func PrintParseTree (node *parseTreeNode) {
	fmt.Println(node.val)
	if(node.right != nil){
		PrintParseTree(node.right)
	}
}

func RecTokenize(content string, tokens []string) []string {
	var token string
	var updatedContent string
	var err error

	token, updatedContent, err = buildToken(content)
	if err != nil {
		fmt.Println("Error reading file:", err)
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
		return "", "", errors.New("Could not recognize token " + content[i:i+1]) 
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

func BuildParseTree(tokens []string) (*parseTreeNode) {
	if len(tokens) <= 0 {
		return nil
	}
	node := parseTreeNode{val: tokens[0]}
	tree := BuildParseTree(tokens[1:])
	if tree != nil {
		node.right = tree
	}
	return &node
}
