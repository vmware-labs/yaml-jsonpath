/*
 * Copyright 2020 Go YAML Path Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package yamlpath

import (
	"fmt"
	"strings"
)

/*
   filterNode represents a node of a filter expression parse tree. Each node is labelled with a lexeme.

   Terminal nodes have one of the following lexemes: root, lexemeFilterAt, lexemeFilterIntegerLiteral,
   lexemeFilterFloatLiteral, lexemeFilterStringLiteral.
   root and lexemeFilterAt nodes also have a slice of lexemes representing the subpath of `$`` or `@``,
   respectively.

   Non-terminal nodes represent either basic filters (simpler predicates of one or two terminal
   nodes) or filter expressions (more complex predicates of basic filters). A filter existence expression
   is represented as a terminal node with lexemeFilterAt or (less commonly) root.

   The following examples illustrate the approach.

   The basic filter `@.child > 3` is represented as the following parse tree (where each node is indicated by
   its lexeme and `<...>` represents the node's children):

       lexemeFilterGreaterThan<lexemeFilterAt,lexemeFilterIntegerLiteral>

   or, graphically:

               >
              / \
       @.child   3

   The filter expression `@.child > 3 && @.other` is represented as the parse tree:

       lexemeFilterConjunction<lexemeFilterGreaterThan<lexemeFilterAt,lexemeFilterIntegerLiteral>,lexemeFilterAt>

   or, graphically:

                               &&
                             /    \
                            >      @.other
                           / \
                    @.child   3

   The filter expression `(@.child < 5 || @.child > 10) && @.other == 'x'` is represented as the parse tree:

       lexemeFilterConjunction<lexemeFilterDisjunction<lexemeFilterLessThan<lexemeFilterAt,lexemeFilterIntegerLiteral>,
                                                       lexemeFilterGreaterThan<lexemeFilterAt,lexemeFilterIntegerLiteral>
                                                      >,
                               lexemeFilterEquality<lexemeFilterAt,lexemeFilterStringLiteral>
                              >

   or, graphically:

                               &&
                        /               \
                      ||                 ==
                  /        \            /  \
               <            >    @.other    'x'
              / \          / \
       @.child   5  @.child   10

   Note that brackets do not appear in the parse tree.
*/
type filterNode struct {
	lexeme   lexeme
	subpath  []lexeme // empty unless lexeme is root or lexemeFilterAt
	children []*filterNode
}

func newFilterNode(lexemes []lexeme) *filterNode {
	return newParser(lexemes).parse()
}

func (n *filterNode) String() string {
	return "---\n" + n.indentedString(0) + "\n---\n"
}

func (n *filterNode) indentedString(indent int) string {
	i := strings.Repeat("    ", indent)
	s := n.lexeme.val
	for _, l := range n.subpath {
		s += l.val
	}
	c := ""
	for _, child := range n.children {
		c += "\n" + child.indentedString(indent+1)
	}
	return fmt.Sprintf("%s%s%s", i, s, c)
}

// parser holds the state of the filter expression parser.
type parser struct {
	input []lexeme      // the lexemes being scanned
	pos   int           // current position in the input
	stack []*filterNode // parser stack
	tree  *filterNode   // parse tree
}

// newParser creates a new parser for the input slice of lexemes.
func newParser(input []lexeme) *parser {
	l := &parser{
		input: input,
		stack: make([]*filterNode, 0),
	}
	return l
}

// push pushes a parse tree on the stack.
func (p *parser) push(tree *filterNode) {
	p.stack = append(p.stack, tree)
}

// pop pops a prase tree from the stack. If the stack is empty, it panics.
func (p *parser) pop() *filterNode {
	if len(p.stack) == 0 {
		panic("parser stack underflow")
	}
	index := len(p.stack) - 1
	element := p.stack[index]
	p.stack = p.stack[:index]
	return element
}

// nextLexeme returns the next item from the input.
func (p *parser) nextLexeme() lexeme {
	if p.pos >= len(p.input) {
		return lexeme{lexemeEOF, ""}
	}
	next := p.input[p.pos]
	p.pos++
	return next
}

// peek returns the next item from the input without consuming the item.
func (p *parser) peek() lexeme {
	if p.pos >= len(p.input) {
		return lexeme{lexemeEOF, ""}
	}
	return p.input[p.pos]
}

func (p *parser) parse() *filterNode {
	if p.peek().typ == lexemeEOF {
		return nil
	}
	p.expression()
	return p.tree
}

func (p *parser) expression() {
	p.conjunction()
	for p.peek().typ == lexemeFilterDisjunction {
		p.push(p.tree)
		p.or()
	}
}

func (p *parser) or() {
	n := p.nextLexeme()
	p.conjunction()
	p.tree = &filterNode{
		lexeme:  n,
		subpath: []lexeme{},
		children: []*filterNode{
			p.pop(),
			p.tree,
		},
	}
}

func (p *parser) conjunction() {
	p.basicFilter()
	for p.peek().typ == lexemeFilterConjunction {
		p.push(p.tree)
		p.and()
	}
}

func (p *parser) and() {
	n := p.nextLexeme()
	p.basicFilter()
	p.tree = &filterNode{
		lexeme:  n,
		subpath: []lexeme{},
		children: []*filterNode{
			p.pop(),
			p.tree,
		},
	}
}

// basicFilter consumes then next basic filter and sets it as the parser's tree. If a basic filter it not next, nil is set.
func (p *parser) basicFilter() {
	n := p.peek()
	if n.typ == lexemeFilterNot {
		p.nextLexeme()
		p.basicFilter()
		p.tree = &filterNode{
			lexeme:  n,
			subpath: []lexeme{},
			children: []*filterNode{
				p.tree,
			},
		}
		return
	}

	p.filterTerm()
	n = p.peek()
	switch n.typ {
	case lexemeFilterEquality, lexemeFilterInequality,
		lexemeFilterGreaterThan, lexemeFilterGreaterThanOrEqual,
		lexemeFilterLessThan, lexemeFilterLessThanOrEqual:
		p.nextLexeme()
		filterTerm := p.tree
		p.filterTerm()
		p.tree = &filterNode{
			lexeme:  n,
			subpath: []lexeme{},
			children: []*filterNode{
				filterTerm,
				p.tree,
			},
		}
	}
}

// filterTerm consumes the next filter term and sets it as the parser's tree. If a filter term is not next, nil is set.
func (p *parser) filterTerm() {
	n := p.peek()
	switch n.typ {
	case lexemeEOF:
		p.tree = nil

	case lexemeFilterAt:
		p.nextLexeme()
		subpath := []lexeme{}
	f:
		for {
			s := p.peek()
			switch s.typ {
			case lexemeIdentity, lexemeDotChild, lexemeBracketChild, lexemeRecursiveDescent, lexemeArraySubscript:
				// TODO: nested filters
				subpath = append(subpath, s)
			default:
				break f
			}
			p.nextLexeme()
		}
		p.tree = &filterNode{
			lexeme:   n,
			subpath:  subpath,
			children: []*filterNode{},
		}

	// TODO: lexemeRoot ($)

	case lexemeFilterIntegerLiteral, lexemeFilterFloatLiteral, lexemeFilterStringLiteral:
		p.nextLexeme()
		p.tree = &filterNode{
			lexeme:   n,
			subpath:  []lexeme{},
			children: []*filterNode{},
		}

	default:
		panic("unexpected lexeme " + n.String())
	}
}
