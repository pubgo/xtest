package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	gt, err := GenerateTests(os.Args[1], nil)
	if err != nil {
		panic(err)
	}

	for _, g := range gt[:] {
		fmt.Println(g.Path)
		for _, fn := range g.Functions {
			_dt, _ := json.MarshalIndent(fn, " ", "\t")
			fmt.Println(string(_dt))
		}
	}
}
