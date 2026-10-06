package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestMeasureStyleEmitsDensityResolvedJSONOnly(t *testing.T) {
	bundle, err := filepath.Abs("../../planning/wm-design-contracts/v11/intake-20261006-649-frozen/bundle")
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := runMeasureStyle([]string{"--bundle", bundle, "--style", "small", "--text", "A compact cell value", "--width", "180", "--density", "dense", "--scope", "cell"})
	_ = w.Close()
	os.Stdout = stdout
	if runErr != nil {
		t.Fatal(runErr)
	}
	raw, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	var measured measuredStyle
	if err := json.Unmarshal(raw, &measured); err != nil {
		t.Fatalf("stdout was not a single JSON result: %v\n%s", err, raw)
	}
	if measured.Status != "measurement_only" || measured.NativeFit != "not_evaluated" || measured.Density != "dense" || measured.Scope != "cell" {
		t.Fatalf("unexpected result metadata: %+v", measured)
	}
	if measured.Layout.Style.Size != 8 || measured.Layout.Style.Leading != 11 || measured.Layout.NativeQualified {
		t.Fatalf("dense cell-small should report 8/11 pt as unqualified measurement: %+v", measured.Layout)
	}
	if len(measured.Layout.Lines) == 0 {
		t.Fatal("missing shaped line data")
	}
	if strings.TrimSpace(string(raw)) == "" || strings.TrimSpace(string(raw))[0] != '{' {
		t.Fatalf("stdout contains non-JSON text: %q", raw)
	}
}

func TestMeasureStyleRejectsInvalidDensityAndScope(t *testing.T) {
	bundle, err := filepath.Abs("../../planning/wm-design-contracts/v11/intake-20261006-649-frozen/bundle")
	if err != nil {
		t.Fatal(err)
	}
	base := []string{"--bundle", bundle, "--style", "small", "--text", "Example", "--width", "180"}
	for _, tc := range []struct {
		name, option, value, want string
	}{
		{name: "level", option: "--density", value: "tiny", want: "density.unknown_level"},
		{name: "scope", option: "--scope", value: "table", want: "density.unknown_scope"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append(append([]string{}, base...), tc.option, tc.value)
			err := runMeasureStyle(args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %s, got %v", tc.want, err)
			}
		})
	}
}

func TestMeasureStyleKeepsFixedRoleAtSelectedDensity(t *testing.T) {
	bundle, err := filepath.Abs("../../planning/wm-design-contracts/v11/intake-20261006-649-frozen/bundle")
	if err != nil {
		t.Fatal(err)
	}
	source, err := wmdesign.Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	base, err := source.Style("source")
	if err != nil {
		t.Fatal(err)
	}
	dense, err := source.StyleForDensity("source", "dense", "body")
	if err != nil {
		t.Fatal(err)
	}
	if dense.Size != base.Size || dense.Leading != base.Leading {
		t.Fatalf("fixed source role changed with body density: base=%+v dense=%+v", base, dense)
	}
}

func TestMeasureStyleDefaultUsesLegacySourceStyle(t *testing.T) {
	bundle, err := filepath.Abs("../../planning/wm-design-contracts/v10/intake-20261006-649-frozen/bundle")
	if err != nil {
		t.Fatal(err)
	}
	source, err := wmdesign.Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	legacyStyle, err := source.Style("small")
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := runMeasureStyle([]string{"--bundle", bundle, "--style", "small", "--text", "Legacy default", "--width", "180"})
	_ = w.Close()
	os.Stdout = stdout
	if runErr != nil {
		t.Fatal(runErr)
	}
	raw, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	var measured measuredStyle
	if err := json.Unmarshal(raw, &measured); err != nil {
		t.Fatalf("stdout was not a single JSON result: %v\n%s", err, raw)
	}
	if measured.Density != "" || measured.Scope != "body" || measured.Layout.Style.Size != legacyStyle.Size || measured.Layout.Style.Leading != legacyStyle.Leading {
		t.Fatalf("measure-style default changed legacy source style: want=%+v got=%+v", legacyStyle, measured)
	}
}
