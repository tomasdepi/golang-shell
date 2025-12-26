package parser

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/shell-starter-go/app/lexer"
)

type SingleCommand struct {
	Args   []string
	Redirs []Redirection
}

type Redirection struct {
	FD     int
	To     string
	Append bool
}

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *Parser) peek() *lexer.Token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.pos]
}

func (p *Parser) next() *lexer.Token {
	t := p.peek()
	if t != nil {
		p.pos++
	}
	return t
}

func (p *Parser) ParseSingleCommand() (*SingleCommand, error) {

	sc := &SingleCommand{}

	for {
		token := p.peek()
		if token == nil {
			break
		}

		switch token.Kind {
		case lexer.WORD:
			sc.Args = append(sc.Args, token.Value)
		case lexer.OPERATOR:
			if isRedirectionOperator(token.Value) {
				err := p.parseRedirection(token, sc)
				if err != nil {
					return nil, err
				}
			}
		}

		p.next()
	}

	return sc, nil
}

func (p *Parser) parseRedirection(token *lexer.Token, sc *SingleCommand) error {

	fd := 1

	// check previous token for explicit FD
	if len(sc.Args) > 0 {
		lastArg := sc.Args[len(sc.Args)-1]

		value, ok := toDigit(lastArg)

		if ok { //it's a number, so valid fd
			sc.Args = sc.Args[:len(sc.Args)-1] //remove last WORD
			fd = value
		}

		p.next()
		nextToken := p.peek()

		if nextToken == nil || nextToken.Kind != lexer.WORD {
			return fmt.Errorf("parse error")
		}

		redir := Redirection{
			FD:     fd,
			To:     nextToken.Value,
			Append: token.Value == ">>",
		}

		sc.Redirs = append(sc.Redirs, redir)

		return nil

	} else {
		return fmt.Errorf("parse error")
	}
}

func isRedirectionOperator(op string) bool {
	return op == ">" || op == ">>"
}

func toDigit(str string) (int, bool) {

	value, err := strconv.Atoi(str)

	if err != nil {
		return 1, false
	}

	return value, true
}
