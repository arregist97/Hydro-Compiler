package tokenizer

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"unicode/utf8"
)

type Token struct {
	Val    string
	Line   int
	Column int
}

func (token *Token) Print() {
	fmt.Println("Val: " + token.Val)
	fmt.Println("Line: " + strconv.Itoa(token.Line))
	fmt.Println("Col: " + strconv.Itoa(token.Column))

}

func Tokenize(content string, tokens []*Token) []*Token {
	return recTokenize(content, 1, 1, tokens)
}

func recTokenize(content string, line int, col int, tokens []*Token) []*Token {
	var tokenVal string
	var skippedLines int
	var tokenCol int
	var tokenSize int
	var updatedContent string
	var err error

	tokenVal, skippedLines, tokenCol, tokenSize, updatedContent, err = buildToken(content, col, 0)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	fmt.Println(tokenVal + "/end")
	if len(tokenVal) > 0 {
		token := Token{
			Val:    tokenVal,
			Line:   line,
			Column: tokenCol,
		}
		fmt.Println("Token Created!")
		token.Print()
		tokens = append(tokens, &token)
	}
	if skippedLines <= 0 && tokenVal != "\n" {
		col = col + tokenSize
	} else if tokenVal == "\n" {
		line++
		col = 1
	} else {
		line = line + skippedLines
		col = tokenSize
	}

	if len(content) > 0 {
		return recTokenize(updatedContent, line, col, tokens)
	}

	token := Token{
		Val:    "EOF",
		Line:   line,
		Column: col,
	}
	fmt.Println("Token Created!")
	token.Print()
	tokens = append(tokens, &token)

	return tokens
}

func buildToken(content string, col int, i int) (string, int, int, int, string, error) {
	var updatedToken string
	var updatedContent string
	var updatedSize int
	var colPlace int
	var tokenSize int
	var tokenCol = col
	var skippedLines = 0
	var err error = nil

	if len(content) <= i {
		return "", 0, 0, 0, "", err
	}

	r, size := utf8.DecodeRuneInString(content[i:])
	if r == utf8.RuneError {
		if size == 1 {
			return "", 0, 0, size, "", errors.New("could not recognize token " + content[i:i+1])
		} else {
			return "", 0, 0, size, "", errors.New("empty decode error")
		}
	}

	peek, _ := utf8.DecodeRuneInString(content[i+size:])
	if peek == utf8.RuneError {
		updatedToken = string(r)
		updatedContent = content[i+size:]
		return updatedToken, skippedLines, col, size, updatedContent, err
	}

	if r == ' ' && i == 0 {
		updatedToken, skippedLines, tokenCol, tokenSize, updatedContent, err = buildToken(content[size:], col+size, 0)
		updatedSize = tokenSize + size
	} else if r == '/' && (peek == '/' || peek == '*') {
		lnCmt := false
		if peek == '/' {
			lnCmt = true
		}
		updatedToken = ""
		updatedContent, skippedLines, colPlace, err = skipComment(content, i+2, lnCmt)
		updatedSize = colPlace + size
	} else if r == '\n' {
		updatedToken = "\n"
		updatedContent, skippedLines, colPlace, err = skipBlankSpace(content, i+1)
		updatedSize = colPlace + size
	} else if isEndOfToken(r) || isEndOfToken(peek) {
		updatedToken = string(r)
		updatedContent = content[i+size:]
		updatedSize = size
	} else {
		var token string
		token, skippedLines, tokenCol, tokenSize, updatedContent, err = buildToken(content, col, i+size)
		updatedToken = string(r) + token
		updatedSize = tokenSize + size
	}
	return updatedToken, skippedLines, tokenCol, updatedSize, updatedContent, err
}

func skipBlankSpace(content string, i int) (string, int, int, error) {
	var skippedLines = 0
	var columnPlace = 1
	var updatedContent string
	var err error = nil

	r, size := utf8.DecodeRuneInString(content[i:])
	if r == utf8.RuneError {
		if size == 1 {
			return "", 0, 0, errors.New("could not recognize token " + content[i:i+1])
		} else {
			return "", 0, 0, errors.New("empty decode error")
		}
	}

	if r == ' ' {
		updatedContent, skippedLines, columnPlace, err = skipBlankSpace(content, i+1)
		return updatedContent, skippedLines, columnPlace + size, err
	}
	if r == '\n' {
		updatedContent, skippedLines, _, err = skipBlankSpace(content, i+1)
		columnPlace = 1
		return updatedContent, skippedLines + 1, columnPlace, err
	}

	return content[i:], skippedLines, columnPlace, nil

}

func skipComment(content string, i int, lineComment bool) (string, int, int, error) {
	var skippedLines = 0
	var columnPlace int
	var updatedContent string
	var err error = nil

	r, size := utf8.DecodeRuneInString(content[i:])
	if r == utf8.RuneError {
		if size == 1 {
			return "", 0, size, errors.New("could not recognize token " + content[i:i+1])
		} else {
			return "", 0, size, errors.New("empty decode error")
		}
	}

	if lineComment && r == '\n' {
		updatedContent := content[i:]
		return updatedContent, skippedLines, size, nil
	}

	if r != '*' || lineComment {
		updatedContent, skippedLines, columnPlace, err = skipComment(content, i+1, lineComment)
		if r == '\n' {
			skippedLines++
			columnPlace = 1
			return updatedContent, skippedLines, columnPlace, err
		}
		return updatedContent, skippedLines, columnPlace + size, err
	}

	peek, s := utf8.DecodeRuneInString(content[i+size:])
	size = size + s
	if peek == utf8.RuneError {
		if s == 1 {
			return "", 0, size, errors.New("could not recognize token " + content[i:i+1])
		} else {
			return "", 0, size, errors.New("empty decode error")
		}
	}

	if peek == '/' {
		updatedContent := content[i+size:]
		return updatedContent, skippedLines, size, nil
	}

	updatedContent, skippedLines, columnPlace, err = skipComment(content, i+1, lineComment)
	return updatedContent, skippedLines, columnPlace + size, err

}

func isEndOfToken(a rune) bool {
	var endOfTokenRunes = [...]rune{'(', ')', '{', '}', ' ', '\n', '=', '+', '*', '-', '/'}

	for _, b := range endOfTokenRunes {
		if b == a {
			return true
		}
	}
	return false
}
