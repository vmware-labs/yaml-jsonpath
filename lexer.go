/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"fmt"
	"strings"
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
	lexemeEOF // lexing complete
)

// stateFn represents the state of the lexer as a function that returns the next state.
// A nil stateFn indicates lexing is complete.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name  string      // name of the lexer, used only for error reports
	input string      // the string being scanned
	start int         // start position of this item
	pos   int         // current position in the input
	width int         // width of last rune read from input
	state stateFn     // lexer state
	items chan lexeme // channel of scanned lexemes
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		state: lexPath,
		items: make(chan lexeme, 2),
	}
	return l
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

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes a lexeme back to the client.
func (l *lexer) emit(typ lexemeType) {
	l.items <- lexeme{
		typ: typ,
		val: l.input[l.start:l.pos],
	}
	l.start = l.pos
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
	root             string = "$"
	dot              string = "."
	leftBracket      string = "["
	rightBracket     string = "]"
	bracketQuote     string = "['"
	quoteBracket     string = "']"
	recursiveDescent string = ".."
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
	if l.empty() {
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
			if le == '.' || le == '[' || le == eof {
				l.backup()
				break
			}
			childName = true
		}
		if !childName {
			return l.errorf("child name missing after .")
		}
		l.emit(lexemeDotChild)

		if l.hasPrefix(leftBracket) && !l.hasPrefix(bracketQuote) {
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
			l.emit(lexemeArraySubscript)
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

		if l.hasPrefix(leftBracket) && !l.hasPrefix(bracketQuote) {
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
			l.emit(lexemeArraySubscript)
		}

		return lexSubPath
	}

	panic("not implemented!")
}
