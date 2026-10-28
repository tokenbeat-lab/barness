package main

import (
 "encoding/json"
 "os"

 "github.com/tokenbeat-lab/barness/ai"
)

func main() {
 catalog := ai.BuiltinCatalog()
 hash, err := catalog.Hash()
 if err != nil { panic(err) }
 encoder := json.NewEncoder(os.Stdout)
 encoder.SetIndent("", "  ")
 if err := encoder.Encode(struct {
  Hash string `json:"hash"`
  Catalog ai.Catalog `json:"catalog"`
 }{hash, catalog}); err != nil { panic(err) }
}
