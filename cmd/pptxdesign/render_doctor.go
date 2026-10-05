package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/nativeexport"
)

func runRenderDoctor(args []string) error {
	flags := flag.NewFlagSet("render-doctor", flag.ContinueOnError)
	staging := flags.String("staging-dir", "", "stable folder to probe with native PPTX open/PDF write (use same value for render; or PPTXGENGO_NATIVE_STAGING)")
	timeout := flags.Duration("timeout", 20*time.Second, "total diagnostic time budget")
	jsonOutput := flags.Bool("json", false, "emit structured checks")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("render-doctor accepts no positional arguments")
	}
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	checks := nativeexport.Doctor(context.Background(), nativeexport.DoctorOptions{StagingRoot: *staging, Timeout: *timeout})
	if *jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(checks); err != nil {
			return err
		}
	} else {
		for _, check := range checks {
			fmt.Printf("%s %s: %s", check.Status, check.Check, strings.Join(strings.Fields(check.Detail), " "))
			if check.Fix != "" {
				fmt.Printf("; fix: %s", strings.Join(strings.Fields(check.Fix), " "))
			}
			fmt.Println()
		}
	}
	for _, check := range checks {
		if check.Status == "fail" {
			return fmt.Errorf("native render prerequisites failed; see render-doctor checks")
		}
	}
	return nil
}
