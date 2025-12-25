package lexer

import "fmt"

type TokenKind int

const (
	WORD TokenKind = iota
	OPERATOR
)

var TokenKindToStringMap = map[TokenKind]string{
	WORD:     "WORD",
	OPERATOR: "OPERATOR",
}

type Token struct {
	Kind  TokenKind
	Value string
}

func (t *Token) PrintDebug() {
	kind := TokenKindToStringMap[t.Kind]
	fmt.Printf("%s(%s)\n", kind, t.Value)
}
