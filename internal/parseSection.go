package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
)

func main() {
	doc := `
    Usage: prog [-hv] ARG
           prog N M

    prog is a program`
	usage := docopt.Int_parseSection("usage:", doc)[0]

  fmt.Println(usage)
}

