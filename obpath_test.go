package obpath_test

import (
	"github.com/bloglovin/obpath"
	"log"
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

func Test_SampleRun(t *testing.T) {
	testData := map[string]stringMap{
		"store": stringMap{
			"books": []interface{}{
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
			},
			"bicycles": []bike{
				bike{
					Color: "red",
					Price: 19.95,
				},
			},
			"counts": []string{"one", "two", "three", "four"},
		},
	}

	tests := [...]string{
		".store",
		".store.books",
		".store.*",
		"..Author",
		".store.counts[*]",
		".store.counts[3]",
		".store.counts[1:2]",
		".store.counts[-2:]",
		".store.counts[:1]",
		".store.counts[:1].Price",
		"[*]",
		"..books[*](has(@.ISBN))",
		".store.books[*](!empty(@.ISBN))",
		".store.books[*](eq(@.Price, 8.99))",
		".store.books[0:4](eq(@.Author, \"Louis L'Amour\"))",
		"..books.*(between(@.Price, 8, 10)).Title",
		"..books[*](gt(@.Price, 9))",
		"..books[*](has(@.Metadata))",
		"..books[*](nonfiction(@.Category))",
		"..books[*](contains(@.Title, 'R')).Title",
		"..books[*](cicontains(@.Title, 'R')).Title",
		".store.*[*](gt(@.Price, 18))",
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

	for _, pathExpression := range tests {
		log.Printf("Path: %v", pathExpression)
		path := obpath.MustCompile(pathExpression, context)

		result := make(chan interface{})
		go path.Evaluate(testData, result)

		for item := range result {
			log.Printf("Got match %v", item)
		}
	}
}
