package obpath_test

import (
	"github.com/bloglovin/obpath"
	"reflect"
	"testing"
)

type book struct {
	Category string
	Author   string
	Title    string
	ISBN     string
	Price    float32
}

type bike struct {
	Color string
	Price float32
}

type stringMap map[string]interface{}

func Test_SyntaxError(t *testing.T) {
	badPath := ".leftOpen[0"
	failures := []string{
		badPath,
		".missingName.",
		".badPredicateName(#(@))",
		".the(predicateNameIsUnknown(@))",
		".missingPredicateArgs(has)",
		"(badPathStart)",
		".badStringLiteral(eq('))",
		".tooManyArgs(gt(@.Price,0,1,2))",
		".tooManyArgs(gt(@.Price))",
		".badArgType(gt('foo',@.Name))",
		".badArgType2(gt(@.Name, @.Role))",
		".predicateCutOff(gt(@.Price,2",
		".epressionCutOff(gt(@.Price,2)",
	}
	context := obpath.NewContext()
	for _, path := range failures {
		compilePathHelper(t, path, context)
	}

	_, error := obpath.Compile("", context)
	if error == nil {
		t.Error("Expected an error on empty path expression")
	} else {
		t.Logf("Got syntax error as expected for empty string: %v", error.Error())
	}

	_, error = obpath.Compile("..property", context)
	if error == nil {
		t.Error("Expected descendant selector to be disabled by default.")
	} else {
		t.Logf("Descendant selector disabled by default as expected: %v", error.Error())
	}

	context.AllowDescendants = true
	_, error = obpath.Compile("..property", context)
	if error == nil {
		t.Logf("Successfully enabled descendant selector")
	} else {
		t.Errorf("Descendant selector triggered a syntax error even when it was enabled: %v", error.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("MustCompile triggered panic as expected: %v", r)
		} else {
			t.Errorf("MustCompile didn't trigger a panic on syntax error in: %v", badPath)
		}
	}()
	_ = obpath.MustCompile(badPath, context)
}

func compilePathHelper(t *testing.T, path string, context *obpath.Context) {
	_, error := obpath.Compile(path, context)

	if error == nil {
		t.Errorf("Expected a syntax error when compiling the path: %v", path)
	} else {
		t.Logf("Got syntax error as expected: %v", error.Error())
	}
}

func Test_SampleRun(t *testing.T) {
	books := []interface{}{
		stringMap{
			"Category": "reference",
			"Author":   "Nigel Rees",
			"Title":    "Sayings of the Century",
			"Price":    8.95,
		},
		book{
			Category: "fiction",
			Author:   "Evelyn Waugh",
			Title:    "Sword of Honour",
			Price:    12.99,
		},
		book{
			Category: "fiction",
			Author:   "Louis L'Amour",
			Title:    "Westward the Tide",
			ISBN:     "0-553-24766-2",
			Price:    5.52,
		},
		stringMap{
			"Category": "fiction",
			"Author":   "Herman Melville",
			"Title":    "Moby Dick",
			"ISBN":     "0-553-21311-3",
			"Price":    8.99,
			"Metadata": stringMap{
				"Info": "foobar",
			},
		},
		stringMap{
			"Category": "fiction",
			"Author":   "J. R. R. Tolkien",
			"Title":    "The Lord of the Rings",
			"ISBN":     "0-395-19395-8",

			"Price": 22.99,
		},
	}
	bikes := []bike{
		bike{
			Color: "red",
			Price: 19.95,
		},
	}
	testData := map[string]stringMap{
		"store": stringMap{
			"books":    books,
			"bicycles": bikes,
			"counts":   []string{"one", "two", "three", "four"},
			"wombats":  []interface{}{},
		},
	}

	tests := map[string][]interface{}{
		"[*]":                                                []interface{}{},
		".store":                                             []interface{}{testData["store"]},
		".store.books":                                       []interface{}{testData["store"]["books"]},
		".store.counts[*]":                                   []interface{}{"one", "two", "three", "four"},
		".store.counts[3]":                                   []interface{}{"four"},
		".store.counts[3:10]":                                []interface{}{"four"},
		".store.counts[1:2]":                                 []interface{}{"two", "three"},
		".store.counts[-2:]":                                 []interface{}{"three", "four"},
		".store.counts[:1]":                                  []interface{}{"one", "two"},
		".store.counts[:1].Price":                            []interface{}{},
		".store.wombats[0:10]":                               []interface{}{},
		"..books[*](gt(@.Title, 10))":                        []interface{}{},
		"..books[*](gte(@.Title, 10))":                       []interface{}{},
		"..books[*](lt(@.Title, 10))":                        []interface{}{},
		"..books[*](lte(@.Title, 10))":                       []interface{}{},
		"..books[*](between(@.Title, 10, 20))":               []interface{}{},
		"..books[*](lt(@.Price, 6)).Title":                   []interface{}{"Westward the Tide"},
		"..books[*](lte(@.Price, 6)).Title":                  []interface{}{"Westward the Tide"},
		"..books[*](between(@.Price, 12, 13)).Title":         []interface{}{"Sword of Honour"},
		"..books[*](has(@.ISBN))":                            books[1:],
		".store.books[*](!empty(@.ISBN))":                    books[2:],
		".store.books[*](eq(@.Price, 8.99))":                 books[3:4],
		".store.books[0:4](eq(@.Author, \"Louis L'Amour\"))": books[2:3],
		"..books[*](has(@.Metadata))":                        books[3:4],
		"..books[*](nonfiction(@.Category))":                 books[0:1],
		"..books[*](contains(@.Title, 'R')).Title":           []interface{}{"The Lord of the Rings"},
		".store.*[*](gt(@.Price, 18))":                       []interface{}{books[4], bikes[0]},
		".store.*[*](gte(@.Price, 18))":                      []interface{}{books[4], bikes[0]},
		"..bicycles[0].*":                                    []interface{}{"red", bikes[0].Price},
		".store.*": []interface{}{
			testData["store"]["books"],
			testData["store"]["bicycles"],
			testData["store"]["counts"],
			testData["store"]["wombats"],
		},
		"..Author": []interface{}{
			"Nigel Rees", "Evelyn Waugh", "Louis L'Amour",
			"Herman Melville", "J. R. R. Tolkien",
		},
		"..books[*](cicontains(@.Title, 'R')).Title": []interface{}{
			"Sayings of the Century", "Sword of Honour",
			"Westward the Tide", "The Lord of the Rings",
		},
		"..books.*(between(@.Price, 8, 10)).Title": []interface{}{
			"Sayings of the Century",
			"Moby Dick",
		},
		"..books[*](gt(@.Price, 9))": []interface{}{
			books[1],
			books[4],
		},
	}

	context := obpath.NewContext()
	context.AllowDescendants = true

	// Add a pretty stupid custom condition
	context.ConditionFunctions["nonfiction"] = &obpath.ConditionFunction{
		TestFunction: func(arguments []obpath.ExpressionArgument) bool {
			matches := arguments[0].Value.([]interface{})
			for _, match := range matches {
				if reflect.ValueOf(match).String() != "fiction" {
					return true
				}
			}
			return false
		},
		Arguments: []int{
			obpath.PathArg,
		},
	}

	for pathExpression, expected := range tests {
		t.Logf("Testing path: %v", pathExpression)
		path := obpath.MustCompile(pathExpression, context)

		result := make(chan interface{})
		go path.Evaluate(testData, result)

		cursor := 0
		for item := range result {
			if cursor >= len(expected) {
				t.Logf("Extraneous result: %v", item)
				t.Errorf("Got more results that expected for %v", pathExpression)
				return
			}

			expectedItem := expected[cursor]
			if !reflect.DeepEqual(item, expectedItem) {
				t.Logf("Match: %v", item)
				t.Logf("Expected: %v", expectedItem)
				t.Errorf("Item %v for result of %v doesn't match the expected value", cursor, pathExpression)
				return
			}
			cursor++
		}
		if cursor < len(expected) {
			t.Logf("Expected: %v", expected[cursor])
			t.Errorf("Got less results that expected for %v", pathExpression)
			return
		}
	}
}
