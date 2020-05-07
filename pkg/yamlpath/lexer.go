/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// This lexer was based on Rob Pike's talk "Lexical Scanning in Go" (https://talks.golang.org/2011/lex.slide#1)

// a lexeme is a token returned from the lexer
type lexeme struct {
	typ lexemeType
	val string
}

func (l lexeme) String() string {
	switch l.typ {
	case lexemeEOF:
		return "EOF"

	case lexemeError:
		return l.val

	default:
		return fmt.Sprintf("%q", l.val)
	}
}

func (l lexeme) literalValue() string {
	switch l.typ {
	case lexemeEOF:
		return "EOF"

	case lexemeFilterStringLiteral:
		return l.val[1 : len(l.val)-1]

	case lexemeFilterRegularExpressionLiteral:
		return sanitiseRegularExpressionLiteral(l.val)

	default:
		return l.val
	}
}

func sanitiseRegularExpressionLiteral(re string) string {
	return strings.ReplaceAll(re[1:len(re)-1], `\/`, `/`)

}

func (l lexeme) comparator() comparator {
	switch l.typ {
	case lexemeFilterEquality:
		return equal

	case lexemeFilterInequality:
		return notEqual

	case lexemeFilterGreaterThan:
		return greaterThan

	case lexemeFilterGreaterThanOrEqual:
		return greaterThanOrEqual

	case lexemeFilterLessThan:
		return lessThan

	case lexemeFilterLessThanOrEqual:
		return lessThanOrEqual

	default:
		return falseComparator
	}
}

type lexemeType int

const (
	lexemeError lexemeType = iota // lexing error (error message is the lexeme value)
	lexemeIdentity
	lexemeRoot
	lexemeDotChild
	lexemeBracketChild
	lexemeRecursiveDescent
	lexemeArraySubscript
	lexemeFilterBegin
	lexemeFilterEnd
	lexemeFilterOpenBracket
	lexemeFilterCloseBracket
	lexemeFilterNot
	lexemeFilterAt
	lexemeFilterAnd
	lexemeFilterOr
	lexemeFilterEquality
	lexemeFilterInequality
	lexemeFilterGreaterThan
	lexemeFilterGreaterThanOrEqual
	lexemeFilterLessThanOrEqual
	lexemeFilterLessThan
	lexemeFilterMatchesRegularExpression
	lexemeFilterIntegerLiteral
	lexemeFilterFloatLiteral
	lexemeFilterStringLiteral
	lexemeFilterRegularExpressionLiteral
	lexemeEOF // lexing complete
)

// stateFn represents the state of the lexer as a function that returns the next state.
// A nil stateFn indicates lexing is complete.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name                  string      // name of the lexer, used only for error reports
	input                 string      // the string being scanned
	start                 int         // start position of this item
	pos                   int         // current position in the input
	width                 int         // width of last rune read from input
	state                 stateFn     // lexer state
	stack                 []stateFn   // lexer stack
	items                 chan lexeme // channel of scanned lexemes
	lastEmittedStart      int         // start position of last scanned lexeme
	lastEmittedLexemeType lexemeType  // type of last emitted lexeme (or lexemEOF if no lexeme has been emitted)
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:                  name,
		input:                 input,
		state:                 lexPath,
		stack:                 make([]stateFn, 0),
		items:                 make(chan lexeme, 2),
		lastEmittedLexemeType: lexemeEOF,
	}
	return l
}

// push pushes a state function on the stack which will be resumed when parsing terminates.
func (l *lexer) push(state stateFn) {
	l.stack = append(l.stack, state)
}

// pop pops a state function from the stack. If the stack is empty, returns an error function.
func (l *lexer) pop() stateFn {
	if len(l.stack) == 0 {
		return l.errorf("lexer stack underflow")
	}
	index := len(l.stack) - 1
	element := l.stack[index]
	l.stack = l.stack[:index]
	return element
}

// empty returns true if and onl if the stack of state functions is empty.
func (l *lexer) emptyStack() bool {
	return len(l.stack) == 0
}

// nextLexeme returns the next item from the input.
func (l *lexer) nextLexeme() lexeme {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			if l.state == nil {
				return lexeme{
					typ: lexemeEOF,
				}
			}
			l.state = l.state(l)
		}
	}
}

const eof rune = -1 // invalid Unicode code point

// next returns the next rune in the input.
func (l *lexer) next() (rune rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return rune
}

// peek returns the next rune in the input but without consuming it.
// it is equivalent to calling next() followed by backup()
func (l *lexer) peek() (rune rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	return rune
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// stripWhitespace strips out whitespace
// it should only be called immediately after emitting a lexeme
func (l *lexer) stripWhitespace() {
	// find whitespace
	for {
		nextRune := l.next()
		if !unicode.IsSpace(nextRune) {
			l.backup()
			break
		}
	}
	// strip any whitespace
	l.start = l.pos
}

// emit passes a lexeme back to the client.
func (l *lexer) emit(typ lexemeType) {
	l.items <- lexeme{
		typ: typ,
		val: l.value(),
	}
	l.lastEmittedStart = l.start
	l.start = l.pos
	l.lastEmittedLexemeType = typ
}

// value returns the portion of the current lexeme scanned so far
func (l *lexer) value() string {
	return l.input[l.start:l.pos]
}

// context returns the last emitted lexeme (if any) followed by the portion
// of the current lexeme scanned so far
func (l *lexer) context() string {
	return l.input[l.lastEmittedStart:l.pos]
}

// nextChar returns the next character in the input
func (l *lexer) nextChar() string {
	if l.pos >= len(l.input) {
		return ""
	}
	b := []byte(l.input[l.pos:])
	_, size := utf8.DecodeRune(b)
	return string(b[:size])
}

// emitSynthetic passes a lexeme back to the client which wasn't encountered in the input.
// The lexing position is not modified.
func (l *lexer) emitSynthetic(typ lexemeType, val string) {
	l.items <- lexeme{
		typ: typ,
		val: val,
	}
}

func (l *lexer) empty() bool {
	return l.pos >= len(l.input)
}

func (l *lexer) hasPrefix(p string) bool {
	return strings.HasPrefix(l.input[l.pos:], p)
}

// errorf returns an error lexeme with context and terminates the scan
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- lexeme{
		typ: lexemeError,
		val: fmt.Sprintf("%s at position %d, following %q", fmt.Sprintf(format, args...), l.pos, l.context()),
	}
	return nil
}

// rawErrorf returns an error lexeme with no context and terminates the scan
func (l *lexer) rawErrorf(format string, args ...interface{}) stateFn {
	l.items <- lexeme{
		typ: lexemeError,
		val: fmt.Sprintf(format, args...),
	}
	return nil
}

const (
	root                                    string = "$"
	dot                                     string = "."
	leftBracket                             string = "["
	rightBracket                            string = "]"
	bracketQuote                            string = "['"
	quoteBracket                            string = "']"
	filterBegin                             string = "[?("
	filterEnd                               string = ")]"
	filterOpenBracket                       string = "("
	filterCloseBracket                      string = ")"
	filterNot                               string = "!"
	filterAt                                string = "@"
	filterConjunction                       string = "&&"
	filterDisjunction                       string = "||"
	filterEquality                          string = "=="
	filterInequality                        string = "!="
	filterGreaterThan                       string = ">"
	filterGreaterThanOrEqual                string = ">="
	filterLessThan                          string = "<"
	filterLessThanOrEqual                   string = "<="
	filterMatchesRegularExpression          string = "=~"
	filterStringLiteralDelimiter            string = "'"
	filterRegularExpressionLiteralDelimiter string = "/"
	filterRegularExpressionEscape           string = `\`
	recursiveDescent                        string = ".."
)

func lexPath(l *lexer) stateFn {
	if l.empty() {
		l.emit(lexemeIdentity)
		l.emit(lexemeEOF)
		return nil
	}
	if l.hasPrefix(root) {
		return lexRoot
	}

	// emit implicit root
	l.emitSynthetic(lexemeRoot, root)
	return lexSubPath
}

func lexRoot(l *lexer) stateFn {
	l.pos += len(root)
	l.emit(lexemeRoot)
	return lexSubPath
}

func lexSubPath(l *lexer) stateFn {
	switch {
	case l.hasPrefix(")"):
		return l.pop()

	case l.empty():
		if !l.emptyStack() {
			return l.pop()
		}
		l.emit(lexemeIdentity)
		l.emit(lexemeEOF)
		return nil

	case l.hasPrefix(recursiveDescent):
		l.next()
		l.next()
		childName := false
		for {
			le := l.next()
			if le == '.' || le == '[' || le == eof {
				l.backup()
				break
			}
			childName = true
		}
		if !childName {
			return l.errorf("child name missing")
		}
		l.emit(lexemeRecursiveDescent)
		return lexSubPath

	case l.hasPrefix(dot):
		l.next()
		childName := false
		for {
			le := l.next()
			if le == '.' || le == '[' || le == ')' || le == ' ' || le == '&' || le == '|' || le == '=' || le == '!' || le == '>' || le == '<' || le == eof {
				l.backup()
				break
			}
			childName = true
		}
		if !childName {
			return l.errorf("child name missing")
		}
		l.emit(lexemeDotChild)

		return lexOptionalArrayIndex

	case l.hasPrefix(bracketQuote):
		l.next()
		l.next()
		childName := false
		for {
			if l.hasPrefix(quoteBracket) {
				l.next()
				l.next()
				break
			}
			le := l.next()
			if le == eof {
				return l.errorf("unmatched ['")
			}
			childName = true
		}
		if !childName {
			return l.rawErrorf("child name missing from [''] before position %d", l.pos)
		}
		l.emit(lexemeBracketChild)

		return lexOptionalArrayIndex

	case l.hasPrefix(filterBegin):
		l.next()
		l.next()
		l.next()
		l.emit(lexemeFilterBegin)
		l.push(lexFilterEnd)
		return lexFilterExprInitial

	default:
		return l.errorf("invalid path syntax at position %d, following %q", l.pos, l.context())
	}
}

func lexOptionalArrayIndex(l *lexer) stateFn {
	if l.hasPrefix(leftBracket) && !l.hasPrefix(bracketQuote) && !l.hasPrefix(filterBegin) {
		l.next()
		subscript := false
		for {
			if l.hasPrefix(rightBracket) {
				l.next()
				break
			}
			le := l.next()
			if le == eof {
				return l.errorf("unmatched [")
			}
			subscript = true
		}
		if !subscript {
			return l.rawErrorf("subscript missing from [] before position %d", l.pos)
		}
		if !validateArrayIndex(l) {
			return nil
		}
		l.emit(lexemeArraySubscript)
	}

	le := l.peek()
	if le == ' ' || le == '&' || le == '|' || le == '=' || le == '!' || le == '>' || le == '<' {
		if l.emptyStack() {
			return l.errorf("invalid character %q", l.nextChar())
		}
		return l.pop()
	}

	return lexSubPath
}

func lexFilterExprInitial(l *lexer) stateFn {
	l.stripWhitespace()

	if nextState, present := lexNumericLiteral(l, lexFilterExpr); present {
		return nextState
	}

	if nextState, present := lexStringLiteral(l, lexFilterExpr); present {
		return nextState
	}

	switch {
	case l.hasPrefix(filterOpenBracket):
		l.next()
		l.emit(lexemeFilterOpenBracket)
		l.push(lexFilterExpr)
		return lexFilterExprInitial

	case l.hasPrefix(filterCloseBracket) && !l.hasPrefix(filterEnd): // ) which is not part of )]
		l.next()
		l.emit(lexemeFilterCloseBracket)
		return l.pop()

	case l.hasPrefix(filterInequality):
		return l.errorf("missing first operand for binary operator !=")

	case l.hasPrefix(filterNot):
		l.next()
		l.emit(lexemeFilterNot)
		return lexFilterExprInitial

	case l.hasPrefix(filterAt):
		l.next()
		l.emit(lexemeFilterAt)
		l.push(lexFilterExpr)
		return lexSubPath

	case l.hasPrefix(root):
		l.next()
		l.emit(lexemeRoot)
		l.push(lexFilterExpr)
		return lexSubPath

	case l.hasPrefix(filterConjunction):
		return l.errorf("missing first operand for binary operator &&")

	case l.hasPrefix(filterDisjunction):
		return l.errorf("missing first operand for binary operator ||")

	case l.hasPrefix(filterEquality):
		return l.errorf("missing first operand for binary operator ==")

	case l.hasPrefix(filterGreaterThanOrEqual):
		return l.errorf("missing first operand for binary operator >=")

	case l.hasPrefix(filterGreaterThan):
		return l.errorf("missing first operand for binary operator >")

	case l.hasPrefix(filterLessThanOrEqual):
		return l.errorf("missing first operand for binary operator <=")

	case l.hasPrefix(filterLessThan):
		return l.errorf("missing first operand for binary operator <")

	default:
		return l.pop()
	}
}

func lexFilterExpr(l *lexer) stateFn {
	l.stripWhitespace()

	switch {
	case l.empty():
		return l.errorf("missing end of filter")

	case l.hasPrefix(filterEnd):
		return l.pop()

	case l.hasPrefix(filterOpenBracket):
		l.next()
		l.emit(lexemeFilterOpenBracket)
		l.push(lexFilterExpr)
		return lexFilterExprInitial

	case l.hasPrefix(filterCloseBracket):
		l.next()
		l.emit(lexemeFilterCloseBracket)
		return l.pop()

	case l.hasPrefix(filterAt):
		l.next()
		l.emit(lexemeFilterAt)
		l.push(lexFilterExpr)
		return lexSubPath

	case l.hasPrefix(filterConjunction):
		l.next()
		l.next()
		l.emit(lexemeFilterAnd)
		l.stripWhitespace()
		return lexFilterExprInitial

	case l.hasPrefix(filterDisjunction):
		l.next()
		l.next()
		l.emit(lexemeFilterOr)
		l.stripWhitespace()
		return lexFilterExprInitial

	case l.hasPrefix(filterEquality):
		l.next()
		l.next()
		l.emit(lexemeFilterEquality)
		l.push(lexFilterExpr)
		return lexFilterTerm

	case l.hasPrefix(filterInequality):
		l.next()
		l.next()
		l.emit(lexemeFilterInequality)
		l.push(lexFilterExpr)
		return lexFilterTerm

	case l.hasPrefix(filterGreaterThanOrEqual):
		if strings.HasPrefix(l.context(), filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using >=")
		}
		l.next()
		l.next()
		l.emit(lexemeFilterGreaterThanOrEqual)

		l.stripWhitespace()
		if l.hasPrefix(filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using >=")
		}

		l.push(lexFilterExpr)
		return lexFilterTerm

	case l.hasPrefix(filterGreaterThan):
		if strings.HasPrefix(l.context(), filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using >")
		}
		l.next()
		l.emit(lexemeFilterGreaterThan)

		l.stripWhitespace()
		if l.hasPrefix(filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using >")
		}

		l.push(lexFilterExpr)
		return lexFilterTerm

	case l.hasPrefix(filterLessThanOrEqual):
		if strings.HasPrefix(l.context(), filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using <=")
		}
		l.next()
		l.next()
		l.emit(lexemeFilterLessThanOrEqual)

		l.stripWhitespace()
		if l.hasPrefix(filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using <=")
		}

		l.push(lexFilterExpr)
		return lexFilterTerm

	case l.hasPrefix(filterLessThan):
		if strings.HasPrefix(l.context(), filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using <")
		}
		l.next()
		l.emit(lexemeFilterLessThan)

		l.stripWhitespace()
		if l.hasPrefix(filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using <")
		}

		l.push(lexFilterExpr)
		return lexFilterTerm

	case l.hasPrefix(filterMatchesRegularExpression):
		switch l.lastEmittedLexemeType {
		case lexemeFilterStringLiteral, lexemeFilterIntegerLiteral, lexemeFilterFloatLiteral:
			return l.errorf("literal cannot be matched using =~ starting at %q", l.nextChar())
		}
		l.next()
		l.next()
		l.emit(lexemeFilterMatchesRegularExpression)

		l.stripWhitespace()
		return lexRegularExpressionLiteral(l, lexFilterExpr)

	default:
		return l.errorf("invalid filter syntax starting at %q", l.nextChar())
	}
}

func lexFilterTerm(l *lexer) stateFn {
	l.stripWhitespace()

	if l.hasPrefix(filterAt) {
		l.next()
		l.emit(lexemeFilterAt)
		return lexSubPath
	}

	if l.hasPrefix(root) {
		l.next()
		l.emit(lexemeRoot)
		return lexSubPath
	}

	if nextState, present := lexNumericLiteral(l, lexFilterExpr); present {
		return nextState
	}

	if nextState, present := lexStringLiteral(l, lexFilterExpr); present {
		return nextState
	}

	if l.hasPrefix(filterEnd) || l.hasPrefix(filterCloseBracket) {
		return l.errorf("missing filter term")
	}

	return l.errorf("invalid filter term")
}

func lexFilterEnd(l *lexer) stateFn {
	if l.hasPrefix(filterEnd) {
		if l.lastEmittedLexemeType == lexemeFilterBegin {
			return l.errorf("missing filter")
		}
		l.next()
		l.next()
		l.emit(lexemeFilterEnd)
		return lexSubPath
	}

	return l.errorf("invalid filter syntax: missing %s", filterEnd)
}

func validateArrayIndex(l *lexer) bool {
	subscript := l.value()
	index := strings.TrimSuffix(strings.TrimPrefix(subscript, leftBracket), rightBracket)
	if index != "*" {
		sliceParms := strings.Split(index, ":")
		if len(sliceParms) > 3 {
			l.rawErrorf("invalid array index, too many colons: %s before position %d", subscript, l.pos)
			return false
		}
		for _, s := range sliceParms {
			if s != "" {
				if _, err := strconv.Atoi(s); err != nil {
					l.rawErrorf("invalid array index containing non-integer value: %s before position %d", subscript, l.pos)
					return false
				}
			}
		}
	}
	return true
}

func lexNumericLiteral(l *lexer, nextState stateFn) (stateFn, bool) {
	n := l.peek()
	if n == '.' || n == '-' || (n >= '0' && n <= '9') {
		float := n == '.'
		for {
			l.next()
			n := l.peek()
			if n == '.' {
				float = true
				continue
			}
			if !(n >= '0' && n <= '9') {
				break
			}
		}
		if float {
			// validate float
			if _, err := strconv.ParseFloat(l.value(), 64); err != nil {
				err := err.(*strconv.NumError)
				return l.rawErrorf("invalid float literal %q: %s before position %d", err.Num, err.Unwrap(), l.pos), true
			}
			l.emit(lexemeFilterFloatLiteral)
			return lexFilterExpr, true
		}
		// validate integer
		if _, err := strconv.Atoi(l.value()); err != nil {
			err := err.(*strconv.NumError)
			return l.rawErrorf("invalid integer literal %q: %s before position %d", err.Num, err.Unwrap(), l.pos), true
		}
		l.emit(lexemeFilterIntegerLiteral)
		return lexFilterExpr, true
	}
	return nil, false
}

func lexStringLiteral(l *lexer, nextState stateFn) (stateFn, bool) {
	if l.hasPrefix(filterStringLiteralDelimiter) {
		pos := l.pos
		context := l.context()
		for {
			if l.next() == eof {
				return l.rawErrorf(`unmatched string delimiter "'" at position %d, following %q`, pos, context), true
			}
			if l.hasPrefix(filterStringLiteralDelimiter) {
				break
			}
		}
		l.next()
		l.emit(lexemeFilterStringLiteral)

		return nextState, true
	}
	return nil, false
}

func lexRegularExpressionLiteral(l *lexer, nextState stateFn) stateFn {
	if !l.hasPrefix(filterRegularExpressionLiteralDelimiter) {
		return l.errorf("regular expression does not start with /")
	}
	pos := l.pos
	context := l.context()
	escape := false
	for {
		if l.next() == eof {
			return l.rawErrorf(`unmatched regular expression delimiter "/" at position %d, following %q`, pos, context)
		}
		if !escape && l.hasPrefix(filterRegularExpressionLiteralDelimiter) {
			break
		}
		if !escape && l.hasPrefix(filterRegularExpressionEscape) {
			escape = true
		} else {
			escape = false
		}
	}
	l.next()
	if _, err := regexp.Compile(sanitiseRegularExpressionLiteral(l.value())); err != nil {
		return l.rawErrorf(`invalid regular expression at position %d, following %q: %s`, pos, context, err)
	}
	l.emit(lexemeFilterRegularExpressionLiteral)

	return nextState
}
