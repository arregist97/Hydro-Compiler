package tokenizer

import (
	"errors"
	"fmt"
	"log"
	"unicode/utf8"
)

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

func buildToken(content string, iOpt ...int) (string, string, error) {
	var updatedToken string
	var updatedContent string
	var err error = nil
	var i int = 0

	if len(iOpt) > 0 {
		i = iOpt[0]
	}

	if len(content) <= i {
		return "", "", err
	}

	r, size := utf8.DecodeRuneInString(content[i:])
	if r == utf8.RuneError {
		if size == 1 {
			return "", "", errors.New("could not recognize token " + content[i:i+1])
		} else {
			return "", "", errors.New("empty decode error")
		}
	}

	peek, _ := utf8.DecodeRuneInString(content[i+size:])
	if peek == utf8.RuneError {
		updatedToken = string(r)
		updatedContent = content[i+size:]
		return updatedToken, updatedContent, err
	}
	if r == ' ' && i == 0 {
		updatedToken, updatedContent, err = buildToken(content[size:])
	} else if r == '/' && (peek == '/' || peek == '*') {
		lnCmt := false
		if peek == '/' {
			lnCmt = true
		}
		updatedToken = ""
		updatedContent, err = skipComment(content, i+2, lnCmt)
	} else if r == '\n' {
		updatedToken = "\n"
		updatedContent, err = skipBlankSpace(content, i+1)
	} else if isEndOfToken(r) || isEndOfToken(peek) {
		updatedToken = string(r)
		updatedContent = content[i+size:]
	} else {
		var token string
		token, updatedContent, err = buildToken(content, i+size)
		updatedToken = string(r) + token
	}
	return updatedToken, updatedContent, err
}

func skipBlankSpace(content string, i int) (string, error) {
	r, size := utf8.DecodeRuneInString(content[i:])
	if r == utf8.RuneError {
		if size == 1 {
			return "", errors.New("could not recognize token " + content[i:i+1])
		} else {
			return "", errors.New("empty decode error")
		}
	}

	if r == '\n' || r == ' ' {
		return skipBlankSpace(content, i+1)
	}

	return content[i:], nil

}

func skipComment(content string, i int, lineComment bool) (string, error) {
	r, s := utf8.DecodeRuneInString(content[i:])
	if r == utf8.RuneError {
		if s == 1 {
			return "", errors.New("could not recognize token " + content[i:i+1])
		} else {
			return "", errors.New("empty decode error")
		}
	}

	size := s

	if lineComment && r == '\n' {
		updatedContent := content[i:]
		return updatedContent, nil
	}

	if r != '*' || lineComment {
		return skipComment(content, i+1, lineComment)
	}

	var peek rune
	peek, s = utf8.DecodeRuneInString(content[i+size:])
	if peek == utf8.RuneError {
		if s == 1 {
			return "", errors.New("could not recognize token " + content[i:i+1])
		} else {
			return "", errors.New("empty decode error")
		}
	}

	size = size + s
	if peek == '/' {
		updatedContent := content[i+size:]
		return updatedContent, nil
	}

	return skipComment(content, i+1, lineComment)

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
