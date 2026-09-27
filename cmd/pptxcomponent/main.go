// pptxcomponent applies source-bound component contracts to native scenes.
package main

import (
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/component"
)

func main() {
	if err := component.Run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
