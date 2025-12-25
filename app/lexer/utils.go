package lexer

func isOperatorChar(r rune) bool {
	return r == '>' || r == '<' || r == '|' || r == '&' || r == '!'
}
