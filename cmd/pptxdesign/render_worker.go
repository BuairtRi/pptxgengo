package main

import (
	"context"
	"fmt"
	"os"

	"github.com/buairtri/pptxgengo/internal/nativeexport"
)

func runRenderWorker(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("render-native-worker accepts no positional arguments")
	}
	return nativeexport.RenderWorker(context.Background(), os.Stdin, os.Stdout)
}
