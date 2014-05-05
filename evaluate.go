package main

import (
	"log"
	"reflect"
)

// ConditionFunction is a function that can be used to filter matches.
type ConditionFunction struct {
	test func(name string, arguments []interface{}) bool
	args int
}

// Context is context in which paths are evaluated against structures
type Context struct {
	conditionFuncs map[string]*ConditionFunction
}

func testEquals(name string, arguments []interface{}) bool {
	return true
}

// NewContext creates a new evaluation context
func NewContext() *Context {
	context := Context{}

	// Set up the default test functions
	context.conditionFuncs = map[string]*ConditionFunction{
		"eq": &ConditionFunction{
			test: testEquals,
			args: 2,
		},
	}

	return &context
}

// Evaluate finds everything matching an expression
func (context *Context) Evaluate(path *Path, object interface{}, result chan<- interface{}) {
	log.Print("Up & running")
	context.evaluateStep(path, 0, object, result)
	close(result)
}

func (context *Context) checkAndEvaluateNextStep(path *Path, index int, object interface{}, result chan<- interface{}) {
	step := path.steps[index]

	if step.condition != nil {
		testFunc := context.conditionFuncs[step.condition.name]
		if testFunc != nil && len(step.condition.arguments) == testFunc.args {

		}
	} else {
		context.evaluateStep(path, index+1, object, result)
	}
}

func (context *Context) evaluateStep(path *Path, index int, object interface{}, result chan<- interface{}) {
	if index >= len(path.steps) {
		result <- object
		return
	}

	zero := reflect.ValueOf(nil)
	step := path.steps[index]
	kind := reflect.TypeOf(object).Kind()
	v := reflect.ValueOf(object)

	if step.target == "child" || step.target == "descendant" {
		// We're looking for map item or struct fields

		if step.name == "*" {
			// Iterate over all child fields, keys or items.
			if kind == reflect.Map {
				for _, key := range v.MapKeys() {
					child := v.MapIndex(key)
					context.checkAndEvaluateNextStep(path, index, child.Interface(), result)
				}
			} else if kind == reflect.Struct {
				length := v.NumField()
				for i := 0; i < length; i++ {
					context.checkAndEvaluateNextStep(path, index, v.Field(i).Interface(), result)
				}
			} else if kind == reflect.Array || kind == reflect.Slice {
				length := v.Len()
				for i := 0; i < length; i++ {
					context.checkAndEvaluateNextStep(path, index, v.Index(i).Interface(), result)
				}
			}
		} else {
			// Step to a named child key or field.
			if kind == reflect.Map {
				child := v.MapIndex(reflect.ValueOf(step.name))

				if child != zero {
					context.checkAndEvaluateNextStep(path, index, child.Interface(), result)
				}
			} else if kind == reflect.Struct {
				child := v.FieldByName(step.name)
				if child != zero {
					context.checkAndEvaluateNextStep(path, index, child.Interface(), result)
				}
			}
		}

		// If we're dealing with a descendant selector we want to step down in the
		// data structure without moving on to the next path part.
		if step.target == "descendant" {
			if kind == reflect.Map {
				for _, key := range v.MapKeys() {
					context.evaluateStep(path, index, v.MapIndex(key).Interface(), result)
				}
			} else if kind == reflect.Struct {
				length := v.NumField()
				for i := 0; i < length; i++ {
					context.evaluateStep(path, index, v.Field(i).Interface(), result)
				}
			} else if kind == reflect.Array || kind == reflect.Slice {
				length := v.Len()
				for i := 0; i < length; i++ {
					context.evaluateStep(path, index, v.Index(i).Interface(), result)
				}
			}
		}
	} else if step.target == "item" {
		// We're looking for items in an array or slice

		if kind == reflect.Array || kind == reflect.Slice {
			length := v.Len()
			startSlice := sliceBound(step.start, length)
			endSlice := sliceBound(step.end, length)

			for i := startSlice; i <= endSlice; i++ {
				if step.condition != nil {

				}
				context.checkAndEvaluateNextStep(path, index, v.Index(i).Interface(), result)
			}
		}
	}
}

func sliceBound(index int, length int) int {
	if index < 0 {
		index = length + index
	}

	if index < 0 || length == 0 {
		index = 0
	} else if index >= length {
		index = length - 1
	}

	return index
}
