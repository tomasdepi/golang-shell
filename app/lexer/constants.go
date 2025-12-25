package lexer

type State int

const (
	StateNormal State = iota
	StateSingleQuote
	StateDoubleQuote
	StateBackslash
	StateOperator
	StateRedirection
)

const (
	SPACE        rune = ' '
	SINGLE_QUOTE rune = '\''
	DOUBLE_QUOTE rune = '"'
	BACKSLASH    rune = '\\'
	DOLAR        rune = '$'
	BACKTICK     rune = '`'
	GREATER      rune = '>'
	LESSER       rune = '<'
	PIPE         rune = '|'
	AND          rune = '&'
	P_OPEN       rune = '('
	P_CLOSE      rune = ')'
)

var DOUBLE_QUOTE_BACKSLASH_CAN_ESCAPE = []rune{
	DOUBLE_QUOTE,
	BACKSLASH,
	DOLAR,
	BACKTICK,
}
