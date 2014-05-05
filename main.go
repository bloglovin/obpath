package main

import (
	"log"
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

func main() {
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
				stringMap{
					"Category": "fiction",
					"Author":   "Evelyn Waugh",
					"Title":    "Sword of Honour",
					"Price":    12.99,
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
		"..books[*](eq(@.Price, 8.99))",
		"..books[0:4](eq(@.Author, \"Louis L'Amour\"))",
		"..books[*](between(@.Price, 9, 10)).Title",
		"..books[](gt(@.Price, 9))",
	}

	for _, path := range tests {
		log.Printf("Path: %v", path)
		compiled := MustCompile(path)
		log.Printf("Compiled: %v", compiled)

		result := make(chan interface{})
		context := NewContext()
		go context.Evaluate(compiled, testData, result)

		for item := range result {
			log.Printf("Got match %v", item)
		}
	}
}
