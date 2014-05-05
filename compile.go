package main

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

// Path is a compiled path that can be applied to an interface{} to get matches
type Path struct {
	path  string
	steps []pathStep
}

// SyntaxError describes a path parser error
type SyntaxError struct {
	message string
	// The index in the string where the error was encountered
	Index int
}

// The error message
func (error *SyntaxError) Error() string {
	return error.message
}

type nodeKind int

const (
	object nodeKind = iota
	array
)

type pathStep struct {
	target    string
	name      string
	start     int
	end       int
	condition *expression
}

type expression struct {
	name      string
	arguments []interface{}
}

// MustCompile returns the compiled path, and panics if
// there are any errors.
func MustCompile(path string) *Path {
	compiled, err := Compile(path)
	if err != nil {
		panic(err)
	}
	return compiled
}

// Compile returns the compiled path.
func Compile(path string) (*Path, error) {
	c := compiler{path, 0}
	if path == "" {
		return nil, c.errorf("empty path")
	}
	p, err := c.parsePath()
	if err != nil {
		return nil, err
	}
	return p, nil
}

type compiler struct {
	path  string
	index int
}

func (c *compiler) errorf(format string, args ...interface{}) error {
	return fmt.Errorf("syntax error in path %q at character %d: %s", c.path, c.index, fmt.Sprintf(format, args...))
}

func (c *compiler) parsePath() (path *Path, err error) {
	var steps []pathStep
	var start = c.index

	for {
		step := pathStep{}

		if c.skip('.') {
			if c.skip('.') {
				step.target = "descendant"
			} else {
				step.target = "child"
			}

			mark := c.index
			if !c.skipName() {
				return nil, c.errorf("missing name")
			}
			step.name = c.path[mark:c.index]

			// Check if we're filtering children by expressions
			predError := c.parseExpressions(&step)
			if predError != nil {
				return nil, predError
			}

		} else if c.skip('[') {
			step.target = "item"
			mark := c.index

			if c.skip('*') {
				step.start = 0
				step.end = -1
			} else if c.skipInteger() {
				index, convErr := strconv.ParseInt(c.path[mark:c.index], 10, 64)

				if convErr != nil {
					return nil, c.errorf("failed to parse range offset")
				}

				step.start = int(index)

				if c.skip(':') {
					mark = c.index
					if c.skipInteger() {
						index, convErr = strconv.ParseInt(c.path[mark:c.index], 10, 64)
						if convErr != nil {
							return nil, c.errorf("failed to parse range end")
						}
						step.end = int(index)
					} else {
						step.end = -1
					}
				} else {
					step.end = step.start
				}
			} else if c.skip(':') {
				step.start = 0
				mark = c.index
				if c.skipInteger() {
					index, convErr := strconv.ParseInt(c.path[mark:c.index], 10, 64)
					if convErr != nil {
						return nil, c.errorf("failed to parse range end")
					}
					step.end = int(index)
				}
			}

			if !c.skip(']') {
				return nil, c.errorf("unexpected %q, expected ']'", c.path[c.index])
			}

			// Check if we're filtering items by expressions
			predError := c.parseExpressions(&step)
			if predError != nil {
				return nil, predError
			}
		} else {
			if (start == 0 || start == c.index) && c.index < len(c.path) {
				return nil, c.errorf("unexpected %q", c.path[c.index])
			}
			return &Path{steps: steps, path: c.path[start:c.index]}, nil
		}

		steps = append(steps, step)
	}
	panic("unreachable")
}

func (c *compiler) parseExpressions(step *pathStep) error {
	// The initial ( tells us that we're using filters, it's fine if it's missing
	// that just means that we don't have any expressions.
	if !c.skip('(') {
		return nil
	}

	c.skipAll(' ')

	// Read the name of the expression
	mark := c.index
	if !c.skipName() {
		return c.errorf("unexpected %q, expected expression name", c.path[c.index])
	}
	step.condition = &expression{name: c.path[mark:c.index]}

	// Parenthesis leading in to the argument list
	if !c.skip('(') {
		return c.errorf("unexpected %q, expected '('", c.path[c.index])
	}

	// Read arguments
	for {
		c.skipAll(' ')
		mark = c.index

		// A path reference
		if c.skip('@') {
			refCompiler := compiler{path: c.path, index: c.index}
			refPath, refError := refCompiler.parsePath()

			if refError != nil {
				return refError
			}

			step.condition.arguments = append(step.condition.arguments, refPath)
			c.index = refCompiler.index
		} else if c.peek('"') || c.peek('\'') { // A string literal

			stringArg, litError := c.parseStringLiteral()

			if litError != nil {
				return c.errorf("failed to parse string literal: %v", litError.Error())
			}

			step.condition.arguments = append(step.condition.arguments, stringArg)
		} else if isNumber, isFloat := c.skipNumber(); isNumber { // An integer or float
			if isFloat {
				value, convErr := strconv.ParseFloat(c.path[mark:c.index], 64)
				if convErr != nil {
					return c.errorf("failed to parse float literal")
				}
				step.condition.arguments = append(step.condition.arguments, value)
			} else {
				value, convErr := strconv.ParseInt(c.path[mark:c.index], 10, 64)
				if convErr != nil {
					return c.errorf("failed to parse integer literal")
				}
				step.condition.arguments = append(step.condition.arguments, value)
			}
		}

		// If the next character isn't a comma we don't have any more arguments
		if !c.skip(',') {
			break
		}
	}

	c.skipAll(' ')
	// Parenthesis ending the argument list
	if !c.skip(')') {
		return c.errorf("unexpected %q, expected ')'", c.path[c.index])
	}

	c.skipAll(' ')
	// Parenthesis ending the expression
	if !c.skip(')') {
		return c.errorf("unexpected %q, expected ')'", c.path[c.index])
	}

	return nil
}

func (c *compiler) parseStringLiteral() (string, error) {
	strChars := "\"'`"
	for i := 0; i < len(strChars); i++ {
		ch := strChars[i]
		if c.skip(ch) {
			mark := c.index
			if !c.skipUntil(ch) {
				return "", fmt.Errorf(`missing closing "%v"`, ch)
			}
			return c.path[mark : c.index-1], nil
		}
	}
	return "", c.errorf("unexpected %q, expected string literal", c.path[c.index])
}

func (c *compiler) skip(b byte) bool {
	if c.index < len(c.path) && c.path[c.index] == b {
		c.index++
		return true
	}
	return false
}

func (c *compiler) skipUntil(b byte) bool {
	for i := c.index; i < len(c.path); i++ {
		if c.path[i] == b {
			c.index = i + 1
			return true
		}
	}
	return false
}

func (c *compiler) peek(b byte) bool {
	return c.index < len(c.path) && c.path[c.index] == b
}

func (c *compiler) skipAll(b byte) bool {
	start := c.index
	for c.index < len(c.path) {
		if c.path[c.index] != b {
			break
		}
		c.index++
	}
	return c.index > start
}

func (c *compiler) skipString(s string) bool {
	if c.index+len(s) <= len(c.path) && c.path[c.index:c.index+len(s)] == s {
		c.index += len(s)
		return true
	}
	return false
}

func (c *compiler) skipInteger() bool {
	start := c.index

	if c.path[c.index] == '-' || c.path[c.index] == '+' {
		c.index++
	}

	for c.index < len(c.path) && isNumberByte(c.path[c.index]) {
		c.index++
	}
	return c.index > start
}

func (c *compiler) skipNumber() (bool, bool) {
	start := c.index
	c.skipInteger()
	isFloat := c.skip('.')

	if isFloat {
		for c.index < len(c.path) && isNumberByte(c.path[c.index]) {
			c.index++
		}
	}

	return c.index > start, isFloat
}

func isNumberByte(c byte) bool {
	return '0' <= c && c <= '9'
}

func (c *compiler) skipName() bool {
	if c.index >= len(c.path) {
		return false
	}
	if c.path[c.index] == '*' {
		c.index++
		return true
	}
	start := c.index
	for c.index < len(c.path) && (c.path[c.index] >= utf8.RuneSelf || isNameByte(c.path[c.index])) {
		c.index++
	}
	return c.index > start
}

func isNameByte(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '_' || c == '-'
}
