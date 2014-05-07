# ObPath

ObPath matches path expressions against maps (map[string]*) and structs, much like jsonpath does for json.

To install run "go get github.com/bloglovin/obpath/obp".

This should allow you to run obpath like this to get books that cost more than 10 simoleons:

```bash
wget "https://raw.githubusercontent.com/bloglovin/obpath/master/testdata/sample.json"
cat sample.json | obp --path=".store.books[*](gt(@.Price, 10))"
```

Some sample queries:

```js
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
"..books[*](has(@.ISBN))",
".store.books[*](!empty(@.ISBN))",
".store.books[*](eq(@.Price, 8.99))",
".store.books[0:4](eq(@.Author, \"Louis L'Amour\"))",
"..books.*(between(@.Price, 8, 10)).Title",
"..books[*](gt(@.Price, 9))",
"..books[*](has(@.Metadata))",
"..books[*](contains(@.Title, 'R')).Title",
"..books[*](cicontains(@.Title, 'R')).Title",
".store.*[*](gt(@.Price, 18))"
```

`obp` can handle a newline delimited JSON stream as input and that is also the default output format. To get all matches as an array, specify "--stream=false".

## Programmatic usage

```Go
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
```
