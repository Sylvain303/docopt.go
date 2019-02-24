// exploring the code can also be done using test
// run with: go test -run TestFormalUsage2
// as go test will try to run all test in the folder

package docopt

import (
	"fmt"
	"testing"
)

func TestFormalUsage2(t *testing.T) {
	doc := `
    Usage: prog [-hv] ARG
           prog N M

    prog is a program`
	usage := parseSection("usage:", doc)
  fmt.Printf("%v\n", usage)
}
