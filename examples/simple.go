package main

import (
	"github.com/bloglovin/obpath"
	"log"
)

func main() {
	context := obpath.NewContext()

	// Get all trees up until the second last one
	trees := obpath.MustCompile(".trees[:-2]", context)

	data := map[string]interface{}{
		"trees":   []string{"Elm", "Oak", "Fir"},
		"animals": []string{"Cat", "Dog", "Horse"},
	}

	result := make(chan interface{})
	go trees.Evaluate(data, result)

	for match := range result {
		log.Printf("Match: %#v", match)
	}
}
