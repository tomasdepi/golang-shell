package lexer

import (
	"fmt"
	"slices"
	"strings"
)

type Lexer struct {
	//input []rune
	pos int

	state         State
	previousState State

	current strings.Builder
	tokens  []Token
}

func (l *Lexer) Lex(input string) ([]Token, error) {

	l.pos = 0
	l.tokens = []Token{}

	for l.pos < len(input) {

		ch := input[l.pos]

		switch l.state {
		case StateNormal:
			l.lexNormal(rune(ch))
		case StateSingleQuote:
			l.lexSingleQuote(rune(ch))
		case StateDoubleQuote:
			l.lexDoubleQuote(rune(ch))
		case StateBackslash:
			l.lexBackslash(rune(ch))
		case StateOperator:
			l.lexOperator(rune(ch))
		}

		l.pos++
	}

	switch l.state {
	case StateOperator:
		l.emitToken(OPERATOR)
	case StateNormal:
		l.emitToken(WORD)
	default:
		return nil, fmt.Errorf("unterminated quote")
	}

	l.pos = 0

	return l.tokens, nil
}

func (l *Lexer) TokenToStringSlice() []string {
	stringSlice := make([]string, len(l.tokens))

	for i, token := range l.tokens {
		stringSlice[i] = token.Value
	}

	return stringSlice
}

func (l *Lexer) lexNormal(ch rune) {

	switch {
	case ch == SPACE:
		l.emitToken(WORD)
	case ch == SINGLE_QUOTE:
		l.state = StateSingleQuote
	case ch == DOUBLE_QUOTE:
		l.state = StateDoubleQuote
	case ch == BACKSLASH:
		l.state = StateBackslash
		l.previousState = StateNormal
	case isOperatorChar(ch):
		l.emitToken(WORD)
		l.state = StateOperator
		l.current.WriteRune(ch)
	default:
		l.current.WriteRune(ch)
	}
}

func (l *Lexer) lexSingleQuote(ch rune) {

	switch ch {
	case SINGLE_QUOTE:
		l.state = StateNormal
	default:
		l.current.WriteRune(ch)
	}
}

func (l *Lexer) lexDoubleQuote(ch rune) {
	switch ch {
	case DOUBLE_QUOTE:
		l.state = StateNormal
	case BACKSLASH:
		l.state = StateBackslash
		l.previousState = StateDoubleQuote
	default:
		l.current.WriteRune(ch)
	}
}

func (l *Lexer) lexBackslash(ch rune) {

	if l.previousState == StateDoubleQuote {
		// Within double quotes, a backslash only escapes certain special characters: ", \, $, `, and newline.
		// For all other characters, the backslash is treated literally.
		if !slices.Contains(DOUBLE_QUOTE_BACKSLASH_CAN_ESCAPE, ch) {
			l.current.WriteRune(BACKSLASH)
		}
	}

	l.current.WriteRune(ch)

	l.state = l.previousState
}

func (l *Lexer) lexOperator(ch rune) {
	switch {
	case isOperatorChar(ch):
		l.current.WriteRune(ch)
	default:
		l.emitToken(OPERATOR)
		l.state = StateNormal
		// reprocess this character
		l.lexNormal(ch)
	}
}

func (l *Lexer) emitToken(kind TokenKind) {
	if l.current.Len() != 0 {

		value := l.current.String()

		newToken := Token{
			Kind:  kind,
			Value: value,
		}

		l.tokens = append(l.tokens, newToken)
	}

	l.current.Reset()
}
