package main

import (
	"encoding/json"
	"flag"
	"github.com/bloglovin/obpath"
	"io"
	"log"
	"os"
)

func main() {
	path := flag.String("path", ".*", "Path expression")
	stream := flag.Bool("stream", true, "Emit the results as a newline delimited JSON stream")
	flag.Parse()

	dec := json.NewDecoder(os.Stdin)
	enc := json.NewEncoder(os.Stdout)

	context := obpath.NewContext()
	context.AllowDescendants = true

	compiled, error := obpath.Compile(*path, context)
	if error != nil {
		log.Fatalf("Could not compile path: %v", error)
	}

	index := 0
	len := 8
	buffer := make([]interface{}, len)

	for {
		var input interface{}
		result := make(chan interface{})

		if err := dec.Decode(&input); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalf("Read JSON from stdin: %v", error)
			}
		}
		go compiled.Evaluate(input, result)

		if *stream {
			for item := range result {
				if err := enc.Encode(&item); err != nil {
					log.Println(err)
				}
			}
		} else {
			for item := range result {
				if index == len {
					len *= 2
					resized := make([]interface{}, len)
					copy(resized, buffer)
					buffer = resized
				}
				buffer[index] = item
				index++
			}
		}
	}

	if !(*stream) {
		slice := buffer[:index]
		if err := enc.Encode(&slice); err != nil {
			log.Fatalf("Could not write JSON to stdout: %v", error)
		}
	}
}
