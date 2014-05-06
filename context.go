package main

import (
	"reflect"
	"strings"
)

const (
	// PathArg arguments references items relative to the current item represented as an array of interface{}
	PathArg = 1 << iota
	// FloatArg arguments are number literals with an optional fractional part represented as a 64 bit floats
	FloatArg = 1 << iota
	// IntegerArg arguments are number literals without a fractional part represented as a 64 bit integers
	IntegerArg = 1 << iota
	// StringArg are strings literals bounded by ", ' or ` represented as strings, no escape sequences are recognised
	StringArg = 1 << iota
	// LiteralArg can be any of the literal arguments
	LiteralArg = StringArg | FloatArg | StringArg
)

// TypeNames returns the names of one or more type flags
func TypeNames(argType int) []string {
	names := []string{}
	if argType&PathArg == PathArg {
		names = append(names, "path")
	}
	if argType&FloatArg == FloatArg {
		names = append(names, "float")
	}
	if argType&IntegerArg == IntegerArg {
		names = append(names, "integer")
	}
	if argType&StringArg == StringArg {
		names = append(names, "string")
	}
	return names
}

// ConditionFunction is a function that can be used to filter matches.
type ConditionFunction struct {
	// TestFunction is the function that will be run to determine the truthiness of the expression.
	TestFunction func(arguments []ExpressionArgument) bool
	// Arguments are the accepted argument types
	Arguments []int
}

// Expression is a condition on a path segment
type expression struct {
	Condition *ConditionFunction
	Inverse   bool
	Arguments []ExpressionArgument
}

// ExpressionArgument is an argument that gets passed to a ConditionFunction
type ExpressionArgument struct {
	// Type is the type of the argument
	Type int
	// Value is the value of the argument
	Value interface{}
}

// Context is context in which paths are evaluated against structures
type Context struct {
	// ConditionFuncs are the
	ConditionFunctions map[string]*ConditionFunction
	AllowDescendants   bool
}

// ConditionNames gets the names of the available conditions
func (context *Context) ConditionNames() []string {
	names := make([]string, len(context.ConditionFunctions))

	index := 0
	for name := range context.ConditionFunctions {
		names[index] = name
		index++
	}

	return names
}

func testEquals(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})
	for _, match := range matches {
		if match == arguments[1].Value {
			return true
		}
	}
	return false
}

func testContains(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})
	substring := arguments[1].Value.(string)

	for _, match := range matches {
		if strings.Contains(reflect.ValueOf(match).String(), substring) {
			return true
		}
	}
	return false
}

func testCiContains(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})
	substring := strings.ToLower(arguments[1].Value.(string))

	for _, match := range matches {
		if strings.Contains(strings.ToLower(reflect.ValueOf(match).String()), substring) {
			return true
		}
	}
	return false
}

func testHas(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})
	return len(matches) > 0
}

func testEmpty(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})
	if len(matches) == 0 {
		return true
	}

	allEmpty := true
	for _, match := range matches {
		if match != reflect.Zero(reflect.TypeOf(match)).Interface() {
			allEmpty = false
			break
		}
	}

	return allEmpty
}

func testGreater(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})

	error, f1 := FloatCast(arguments[1].Value)
	if error != nil {
		return false
	}

	for _, match := range matches {
		error, f0 := FloatCast(match)
		if error != nil {
			continue
		}

		if f0 > f1 {
			return true
		}
	}
	return false
}

func testLess(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})

	error, f1 := FloatCast(arguments[1].Value)
	if error != nil {
		return false
	}

	for _, match := range matches {
		error, f0 := FloatCast(match)
		if error != nil {
			continue
		}

		if f0 < f1 {
			return true
		}
	}
	return false
}

func testGreaterOrEqual(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})

	error, f1 := FloatCast(arguments[1].Value)
	if error != nil {
		return false
	}

	for _, match := range matches {
		error, f0 := FloatCast(match)
		if error != nil {
			continue
		}

		if f0 >= f1 {
			return true
		}
	}
	return false
}

func testLessOrEqual(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})

	error, f1 := FloatCast(arguments[1].Value)
	if error != nil {
		return false
	}

	for _, match := range matches {
		error, f0 := FloatCast(match)
		if error != nil {
			continue
		}

		if f0 <= f1 {
			return true
		}
	}
	return false
}

func testBetween(arguments []ExpressionArgument) bool {
	matches := arguments[0].Value.([]interface{})

	error, f1 := FloatCast(arguments[1].Value)
	if error != nil {
		return false
	}

	error, f2 := FloatCast(arguments[2].Value)
	if error != nil {
		return false
	}

	for _, match := range matches {
		error, f0 := FloatCast(match)
		if error != nil {
			continue
		}

		if f0 > f1 && f0 < f2 {
			return true
		}
	}
	return false
}

// NewContext creates a new evaluation context
func NewContext() *Context {
	context := Context{}

	// Set up standard condition functions
	context.ConditionFunctions = map[string]*ConditionFunction{
		"eq": &ConditionFunction{
			TestFunction: testEquals,
			Arguments: []int{
				PathArg,
				LiteralArg,
			},
		},
		"contains": &ConditionFunction{
			TestFunction: testContains,
			Arguments: []int{
				PathArg,
				StringArg,
			},
		},
		"cicontains": &ConditionFunction{
			TestFunction: testCiContains,
			Arguments: []int{
				PathArg,
				StringArg,
			},
		},
		"gt": &ConditionFunction{
			TestFunction: testGreater,
			Arguments: []int{
				PathArg,
				FloatArg,
			},
		},
		"lt": &ConditionFunction{
			TestFunction: testLess,
			Arguments: []int{
				PathArg,
				FloatArg,
			},
		},
		"gte": &ConditionFunction{
			TestFunction: testGreaterOrEqual,
			Arguments: []int{
				PathArg,
				FloatArg,
			},
		},
		"lte": &ConditionFunction{
			TestFunction: testLessOrEqual,
			Arguments: []int{
				PathArg,
				FloatArg,
			},
		},
		"between": &ConditionFunction{
			TestFunction: testBetween,
			Arguments: []int{
				PathArg,
				FloatArg,
				FloatArg,
			},
		},
		"has": &ConditionFunction{
			TestFunction: testHas,
			Arguments: []int{
				PathArg,
			},
		},
		"empty": &ConditionFunction{
			TestFunction: testEmpty,
			Arguments: []int{
				PathArg,
			},
		},
	}

	return &context
}
