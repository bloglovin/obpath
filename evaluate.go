package obpath

import (
	"reflect"
)

// Evaluate finds everything matching an expression
func (path *Path) Evaluate(object interface{}, result chan<- interface{}) {
	path.evaluateStep(0, object, result)
	close(result)
}

func (path *Path) checkAndEvaluateNextStep(index int, object interface{}, result chan<- interface{}) {
	step := path.steps[index]

	if step.condition != nil {
		args := make([]ExpressionArgument, len(step.condition.Arguments))
		for idx, arg := range step.condition.Arguments {
			if arg.Type&PathArg == PathArg {
				path := arg.Value.(*Path)
				result := make(chan interface{})
				go path.Evaluate(object, result)

				values := []interface{}{}
				for item := range result {
					values = append(values, item)
				}
				args[idx] = ExpressionArgument{
					Type:  PathArg,
					Value: values,
				}
			} else {
				args[idx] = arg
			}
		}

		match := step.condition.Condition.TestFunction(args)
		if step.condition.Inverse {
			match = !match
		}
		if match {
			path.evaluateStep(index+1, object, result)
		}
	} else {
		path.evaluateStep(index+1, object, result)
	}
}

func (path *Path) evaluateStep(index int, object interface{}, result chan<- interface{}) {
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
					path.checkAndEvaluateNextStep(index, child.Interface(), result)
				}
			} else if kind == reflect.Struct {
				length := v.NumField()
				for i := 0; i < length; i++ {
					path.checkAndEvaluateNextStep(index, v.Field(i).Interface(), result)
				}
			} else if kind == reflect.Array || kind == reflect.Slice {
				length := v.Len()
				for i := 0; i < length; i++ {
					path.checkAndEvaluateNextStep(index, v.Index(i).Interface(), result)
				}
			}
		} else {
			// Step to a named child key or field.
			if kind == reflect.Map {
				child := v.MapIndex(reflect.ValueOf(step.name))

				if child != zero {
					path.checkAndEvaluateNextStep(index, child.Interface(), result)
				}
			} else if kind == reflect.Struct {
				child := v.FieldByName(step.name)
				if child != zero {
					path.checkAndEvaluateNextStep(index, child.Interface(), result)
				}
			}
		}

		// If we're dealing with a descendant selector we want to step down in the
		// data structure without moving on to the next path part.
		if step.target == "descendant" {
			if kind == reflect.Map {
				for _, key := range v.MapKeys() {
					path.evaluateStep(index, v.MapIndex(key).Interface(), result)
				}
			} else if kind == reflect.Struct {
				length := v.NumField()
				for i := 0; i < length; i++ {
					path.evaluateStep(index, v.Field(i).Interface(), result)
				}
			} else if kind == reflect.Array || kind == reflect.Slice {
				length := v.Len()
				for i := 0; i < length; i++ {
					path.evaluateStep(index, v.Index(i).Interface(), result)
				}
			}
		}
	} else if step.target == "item" {
		// We're looking for items in an array or slice

		if kind == reflect.Array || kind == reflect.Slice {
			length := v.Len()
			startSlice := sliceBound(step.start, length)
			endSlice := sliceBound(step.end, length)

			for i := startSlice; i <= endSlice && i < length; i++ {
				path.checkAndEvaluateNextStep(index, v.Index(i).Interface(), result)
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
