package main

import (
	"encoding/json"
	"flag"
	"github.com/bloglovin/obpath"
	"io"
	"log"
	"os"
	"reflect"
)

type pathResult struct {
	Name    string
	Path    string
	Results []interface{}
}

func main() {
	dataPath := flag.String("data", "testdata/data.json", "Path to file with test data")
	queriesPath := flag.String("queries", "testdata/queries.json", "Path to queries file")
	resultsPath := flag.String("expected", "testdata/expect.jsonstream", "Path to file to read expected values from (or write to, when generating)")
	errorsPath := flag.String("errors", "testdata/syntax_errors.json", "Path to file with path expressions with invalid syntax")
	generate := flag.Bool("generate", false, "Re-generate the expected values")

	if *generate {
		generateExpectedValues(*dataPath, *queriesPath, *resultsPath)
	} else {
		verifyExpectedValues(*dataPath, *resultsPath)
		verifySyntaxErrors(*errorsPath)
	}
}

func generateExpectedValues(dataPath string, queriesPath string, resultsPath string) {
	resultsFile, error := os.Create(resultsPath)
	if error != nil {
		log.Fatal("Could not open exptected values file for writing")
	}
	enc := json.NewEncoder(resultsFile)

	data := readAndDecode(dataPath, "data")
	queries := readAndDecode(queriesPath, "query")

	result := make(chan pathResult)

	for _, queryTuple := range queries.([]interface{}) {
		tuple := queryTuple.([]interface{})
		queryName := tuple[0].(string)
		query := tuple[1].(string)

		go evaluatePath(data, queryName, query, result)
	}

	count := 0
	for item := range result {
		if err := enc.Encode(&item); err != nil {
			log.Println(err)
		}

		count++
		if count == len(queries.([]interface{})) {
			close(result)
		}
	}
}

func verifyExpectedValues(dataPath string, resultsPath string) {
	resultsFile, error := os.Open(resultsPath)
	if error != nil {
		log.Fatal("Could not open expected values file for reading")
	}
	decoder := json.NewDecoder(resultsFile)
	data := readAndDecode(dataPath, "data")
	result := make(chan pathResult)

	var expected interface{}

	log.Printf("Testing expected matches")

	for error == nil {
		if error = decoder.Decode(&expected); error != nil {
			if error == io.EOF {
				break
			} else {
				log.Fatalf("Could not decode expects: %v", error)
			}
		}
		expectMap := expected.(map[string]interface{})
		go evaluatePath(data, expectMap["Name"].(string), expectMap["Path"].(string), result)

		results := <-result

		if !reflect.DeepEqual(results.Results, expectMap["Results"]) {
			log.Printf("Results: %v", results.Results)
			log.Printf("Expected: %v", expectMap["Results"])
			log.Fatalf("Did not get the expected results for %#v", expectMap["Name"])
		} else {
			log.Printf("%#v passed", expectMap["Name"])
		}
	}

	log.Printf("Success")
}

func verifySyntaxErrors(path string) {
	context := obpath.NewContext()
	expressions := readAndDecode(path, "errors").([]interface{})

	log.Printf("Test for syntax errors")

	for _, expression := range expressions {
		_, error := obpath.Compile(expression.(string), context)
		if error == nil {
			log.Fatalf("Expected path expression %#v to result in an error", expression)
		} else {
			log.Printf("%#v passed: %v", expression, error)
		}
	}
}

func readAndDecode(path string, name string) interface{} {
	file, error := os.Open(path)
	if error != nil {
		log.Fatalf("Could not open %v file for reading", name)
	}

	decoder := json.NewDecoder(file)
	var data interface{}
	if error = decoder.Decode(&data); error != nil {
		log.Fatalf("Could not parse %v file", name)
	}

	return data
}

func evaluatePath(data interface{}, name string, path string, result chan<- pathResult) {
	context := obpath.NewContext()
	context.AllowDescendants = true
	compiled := obpath.MustCompile(path, context)

	results := pathResult{
		Name:    name,
		Path:    path,
		Results: []interface{}{},
	}
	matches := make(chan interface{})
	go compiled.Evaluate(data, matches)

	for item := range matches {
		results.Results = append(results.Results, item)
	}
	result <- results
}
