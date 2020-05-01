/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"fmt"
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

func (i lexeme) String() string {
	switch i.typ {
	case lexemeEOF:
		return "EOF"
	case lexemeError:
		return i.val
	}
	return fmt.Sprintf("%q", i.val)
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
	lexemeFilterConjunction
	lexemeFilterDisjunction
	lexemeFilterEquality
	lexemeFilterInequality
	lexemeFilterGreaterThan
	lexemeFilterGreaterThanOrEqual
	lexemeFilterLessThanOrEqual
	lexemeFilterLessThan
	lexemeFilterIntegerLiteral
	lexemeFilterFloatLiteral
	lexemeFilterStringLiteral
	lexemeEOF // lexing complete
)

// stateFn represents the state of the lexer as a function that returns the next state.
// A nil stateFn indicates lexing is complete.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name             string      // name of the lexer, used only for error reports
	input            string      // the string being scanned
	start            int         // start position of this item
	pos              int         // current position in the input
	width            int         // width of last rune read from input
	state            stateFn     // lexer state
	stack            []stateFn   // lexer stack
	items            chan lexeme // channel of scanned lexemes
	lastEmittedStart int         // start position of last scanned lexeme
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		state: lexPath,
		stack: make([]stateFn, 0),
		items: make(chan lexeme, 2),
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

// errorf returns an error lexeme and terminates the scan
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- lexeme{
		typ: lexemeError,
		val: fmt.Sprintf(format, args...),
	}
	return nil
}

const (
	root                         string = "$"
	dot                          string = "."
	leftBracket                  string = "["
	rightBracket                 string = "]"
	bracketQuote                 string = "['"
	quoteBracket                 string = "']"
	bracketFilter                string = "[?("
	filterBracket                string = ")]"
	filterOpenBracket            string = "("
	filterCloseBracket           string = ")"
	filterNot                    string = "!"
	filterAt                     string = "@"
	filterConjunction            string = "&&"
	filterDisjunction            string = "||"
	filterEquality               string = "=="
	filterInequality             string = "!="
	filterGreaterThan            string = ">"
	filterGreaterThanOrEqual     string = ">="
	filterLessThan               string = "<"
	filterLessThanOrEqual        string = "<="
	filterStringLiteralDelimiter string = "'"
	recursiveDescent             string = ".."
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
	if l.hasPrefix(")") {
		return l.pop()
	}
	if l.empty() {
		if !l.emptyStack() {
			return l.pop()
		}
		l.emit(lexemeIdentity)
		l.emit(lexemeEOF)
		return nil
	}

	if l.hasPrefix(recursiveDescent) {
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
			return l.errorf("child name missing after ..")
		}
		l.emit(lexemeRecursiveDescent)
		return lexSubPath

	}

	if l.hasPrefix(dot) {
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
			return l.errorf("child name missing after .")
		}
		l.emit(lexemeDotChild)

		if l.hasPrefix(leftBracket) && !l.hasPrefix(bracketQuote) && !l.hasPrefix(bracketFilter) {
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
				return l.errorf("subscript missing from []")
			}
			if !validateArrayIndex(l) {
				return nil
			}
			l.emit(lexemeArraySubscript)
		}

		le := l.peek()
		if le == ' ' || le == '&' || le == '|' || le == '=' || le == '!' || le == '>' || le == '<' {
			if l.emptyStack() {
				return l.errorf("invalid character %q at position %d in subpath, following %q", l.nextChar(), l.pos, l.context())
			}
			return l.pop()
		}

		return lexSubPath
	}

	if l.hasPrefix(bracketQuote) {
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
			return l.errorf("child name missing from ['']")
		}
		l.emit(lexemeBracketChild)

		if l.hasPrefix(leftBracket) && !l.hasPrefix(bracketQuote) && !l.hasPrefix(bracketFilter) {
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
				return l.errorf("subscript missing from []")
			}
			if !validateArrayIndex(l) {
				return nil
			}
			l.emit(lexemeArraySubscript)
		}

		return lexSubPath
	}

	if l.hasPrefix(bracketFilter) {
		l.next()
		l.next()
		l.next()
		l.emit(lexemeFilterBegin)
		l.push(lexEndBracketFilter)
		return lexFilterExprInitial
	}

	return l.errorf("invalid path syntax at position %d, following %q", l.pos, l.context())
}

func lexFilterExprInitial(l *lexer) stateFn {
	l.stripWhitespace()

	if nextState, present := lexNumericLiteral(l, lexFilterExpr); present {
		return nextState
	}

	if nextState, present := lexStringLiteral(l, lexFilterExpr); present {
		return nextState
	}

	if l.hasPrefix(filterOpenBracket) {
		l.next()
		l.emit(lexemeFilterOpenBracket)
		l.push(lexFilterExpr)
		return lexFilterExprInitial
	}

	if l.hasPrefix(filterCloseBracket) && !l.hasPrefix(filterBracket) { // ) which is not part of )]
		l.next()
		l.emit(lexemeFilterCloseBracket)
		return l.pop()
	}

	if l.hasPrefix(filterInequality) {
		return l.errorf("missing first operand for binary operator !=")
	}

	if l.hasPrefix(filterNot) {
		l.next()
		l.emit(lexemeFilterNot)
		return lexFilterExprInitial
	}

	if l.hasPrefix(filterAt) {
		l.next()
		l.emit(lexemeFilterAt)
		l.push(lexFilterExpr)
		return lexSubPath
	}

	if l.hasPrefix(root) {
		l.next()
		l.emit(lexemeRoot)
		l.push(lexFilterExpr)
		return lexSubPath
	}

	if l.hasPrefix(filterConjunction) {
		return l.errorf("missing first operand for binary operator &&")
	}

	if l.hasPrefix(filterDisjunction) {
		return l.errorf("missing first operand for binary operator ||")
	}

	if l.hasPrefix(filterEquality) {
		return l.errorf("missing first operand for binary operator ==")
	}

	if l.hasPrefix(filterGreaterThanOrEqual) {
		return l.errorf("missing first operand for binary operator >=")
	}

	if l.hasPrefix(filterGreaterThan) {
		return l.errorf("missing first operand for binary operator >")
	}

	if l.hasPrefix(filterLessThanOrEqual) {
		return l.errorf("missing first operand for binary operator <=")
	}

	if l.hasPrefix(filterLessThan) {
		return l.errorf("missing first operand for binary operator <")
	}

	return l.pop()
}

func lexFilterExpr(l *lexer) stateFn {
	l.stripWhitespace()

	if l.empty() {
		return l.errorf("missing end of filter at position %d, following %q", l.pos, l.context())
	}

	if l.hasPrefix(filterBracket) {
		return l.pop()
	}

	if l.hasPrefix(filterOpenBracket) {
		l.next()
		l.emit(lexemeFilterOpenBracket)
		l.push(lexFilterExpr)
		return lexFilterExprInitial
	}

	if l.hasPrefix(filterCloseBracket) {
		l.next()
		l.emit(lexemeFilterCloseBracket)
		return l.pop()
	}

	if l.hasPrefix(filterAt) {
		l.next()
		l.emit(lexemeFilterAt)
		l.push(lexFilterExpr)
		return lexSubPath
	}

	if l.hasPrefix(filterConjunction) {
		l.next()
		l.next()
		l.emit(lexemeFilterConjunction)
		l.stripWhitespace()
		return lexFilterExprInitial
	}

	if l.hasPrefix(filterDisjunction) {
		l.next()
		l.next()
		l.emit(lexemeFilterDisjunction)
		l.stripWhitespace()
		return lexFilterExprInitial
	}

	if l.hasPrefix(filterEquality) {
		l.next()
		l.next()
		l.emit(lexemeFilterEquality)
		l.push(lexFilterExpr)
		return lexFilterTerm
	}

	if l.hasPrefix(filterInequality) {
		l.next()
		l.next()
		l.emit(lexemeFilterInequality)
		l.push(lexFilterExpr)
		return lexFilterTerm
	}

	if l.hasPrefix(filterGreaterThanOrEqual) {
		if strings.HasPrefix(l.context(), filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using >= at position %d, following %q", l.pos, l.context())
		}
		l.next()
		l.next()
		l.emit(lexemeFilterGreaterThanOrEqual)

		l.stripWhitespace()
		if l.hasPrefix(filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using >= at position %d, following %q", l.pos, l.context())
		}

		l.push(lexFilterExpr)
		return lexFilterTerm
	}

	if l.hasPrefix(filterGreaterThan) {
		if strings.HasPrefix(l.context(), filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using > at position %d, following %q", l.pos, l.context())
		}
		l.next()
		l.emit(lexemeFilterGreaterThan)

		l.stripWhitespace()
		if l.hasPrefix(filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using > at position %d, following %q", l.pos, l.context())
		}

		l.push(lexFilterExpr)
		return lexFilterTerm
	}

	if l.hasPrefix(filterLessThanOrEqual) {
		if strings.HasPrefix(l.context(), filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using <= at position %d, following %q", l.pos, l.context())
		}
		l.next()
		l.next()
		l.emit(lexemeFilterLessThanOrEqual)

		l.stripWhitespace()
		if l.hasPrefix(filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using <= at position %d, following %q", l.pos, l.context())
		}

		l.push(lexFilterExpr)
		return lexFilterTerm
	}

	if l.hasPrefix(filterLessThan) {
		if strings.HasPrefix(l.context(), filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using < at position %d, following %q", l.pos, l.context())
		}
		l.next()
		l.emit(lexemeFilterLessThan)

		l.stripWhitespace()
		if l.hasPrefix(filterStringLiteralDelimiter) {
			return l.errorf("strings cannot be compared using < at position %d, following %q", l.pos, l.context())
		}

		l.push(lexFilterExpr)
		return lexFilterTerm
	}

	return l.errorf("invalid filter syntax starting at %q at position %d, following %q", l.nextChar(), l.pos, l.context())
}

func lexFilterTerm(l *lexer) stateFn {
	l.stripWhitespace()

	if l.hasPrefix(filterAt) {
		l.next()
		l.emit(lexemeFilterAt)
		return lexSubPath
	}

	if nextState, present := lexNumericLiteral(l, lexFilterExpr); present {
		return nextState
	}

	if nextState, present := lexStringLiteral(l, lexFilterExpr); present {
		return nextState
	}

	if l.hasPrefix(filterBracket) || l.hasPrefix(filterCloseBracket) {
		return l.errorf("missing filter term")
	}

	return l.errorf("invalid filter term")
}

func lexEndBracketFilter(l *lexer) stateFn {
	if l.hasPrefix(filterBracket) {
		l.next()
		l.next()
		l.emit(lexemeFilterEnd)
		return lexSubPath
	}

	return l.errorf("invalid filter syntax: missing %s", filterBracket)
}

func validateArrayIndex(l *lexer) bool {
	subscript := l.value()
	index := strings.TrimSuffix(strings.TrimPrefix(subscript, leftBracket), rightBracket)
	if index != "*" {
		sliceParms := strings.Split(index, ":")
		if len(sliceParms) > 3 {
			l.errorf("invalid array index, too many colons: %s", subscript)
			return false
		}
		for _, s := range sliceParms {
			if s != "" {
				if _, err := strconv.Atoi(s); err != nil {
					l.errorf("invalid array index containing non-integer value: %s", subscript)
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
				return l.errorf("invalid float literal %q: %s", err.Num, err.Unwrap()), true
			}
			l.emit(lexemeFilterFloatLiteral)
			return lexFilterExpr, true
		}
		// validate integer
		if _, err := strconv.Atoi(l.value()); err != nil {
			err := err.(*strconv.NumError)
			return l.errorf("invalid integer literal %q: %s", err.Num, err.Unwrap()), true
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
				return l.errorf(`unmatched string delimiter "'" at position %d, following %q`, pos, context), true
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
