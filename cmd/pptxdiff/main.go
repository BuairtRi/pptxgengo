// Command pptxdiff compares two rendered PNG images and writes visual QA artifacts.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

const diffGain = 4

type report struct {
	Width                 int          `json:"width"`
	Height                int          `json:"height"`
	ExactMismatchCount    int          `json:"exact_mismatch_count"`
	ExactMismatchFraction float64      `json:"exact_mismatch_fraction"`
	MeanAbsoluteRGBError  float64      `json:"mean_absolute_rgb_channel_error"`
	MaxDelta              uint8        `json:"max_delta_rgb"`
	PixelsOver1           int          `json:"pixels_over_1_rgb_max_delta"`
	PixelsOver3           int          `json:"pixels_over_3_rgb_max_delta"`
	PixelsOver10          int          `json:"pixels_over_10_rgb_max_delta"`
	MismatchBounds        *boundingBox `json:"mismatch_bounding_box,omitempty"`
	ReferenceSHA256       string       `json:"reference_sha256"`
	CandidateSHA256       string       `json:"candidate_sha256"`
	Algorithm             string       `json:"algorithm"`
}

type boundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type result struct {
	report  report
	diff    *image.NRGBA
	overlay *image.NRGBA
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("pptxdiff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	reference := fs.String("reference", "", "reference PNG path")
	candidate := fs.String("candidate", "", "candidate PNG path")
	outDir := fs.String("out", "", "output directory for report.json and PNG artifacts")
	overwrite := fs.Bool("overwrite", false, "replace existing output files")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *reference == "" || *candidate == "" || *outDir == "" || fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: pptxdiff --reference reference.png --candidate candidate.png --out DIRECTORY [--overwrite]")
		return 2
	}
	if err := execute(*reference, *candidate, *outDir, *overwrite); err != nil {
		fmt.Fprintf(stderr, "pptxdiff: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote comparison report and images to %s\n", *outDir)
	return 0
}

func execute(referencePath, candidatePath, outDir string, overwrite bool) error {
	refBytes, err := os.ReadFile(referencePath)
	if err != nil {
		return fmt.Errorf("read reference %q: %w", referencePath, err)
	}
	candBytes, err := os.ReadFile(candidatePath)
	if err != nil {
		return fmt.Errorf("read candidate %q: %w", candidatePath, err)
	}
	ref, err := decodePNG(referencePath, refBytes)
	if err != nil {
		return err
	}
	cand, err := decodePNG(candidatePath, candBytes)
	if err != nil {
		return err
	}
	if ref.Bounds().Dx() != cand.Bounds().Dx() || ref.Bounds().Dy() != cand.Bounds().Dy() {
		return fmt.Errorf("image dimensions differ: reference is %dx%d, candidate is %dx%d; render both at the same pixel dimensions before comparing", ref.Bounds().Dx(), ref.Bounds().Dy(), cand.Bounds().Dx(), cand.Bounds().Dy())
	}
	res := compare(ref, cand, sha256.Sum256(refBytes), sha256.Sum256(candBytes))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", outDir, err)
	}
	outputs := []string{"report.json", "difference.png", "overlay.png"}
	if !overwrite {
		for _, name := range outputs {
			p := filepath.Join(outDir, name)
			if _, err := os.Lstat(p); err == nil {
				return fmt.Errorf("output %q already exists; pass --overwrite to replace it", p)
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("check output %q: %w", p, err)
			}
		}
	}
	if err := writeJSON(filepath.Join(outDir, "report.json"), res.report); err != nil {
		return err
	}
	if err := writePNG(filepath.Join(outDir, "difference.png"), res.diff); err != nil {
		return err
	}
	if err := writePNG(filepath.Join(outDir, "overlay.png"), res.overlay); err != nil {
		return err
	}
	return nil
}

func decodePNG(path string, data []byte) (image.Image, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode PNG %q: %w", path, err)
	}
	if format != "png" {
		return nil, fmt.Errorf("input %q is %s, expected PNG", path, format)
	}
	return img, nil
}

func compare(ref, cand image.Image, refHash, candHash [32]byte) result {
	b := ref.Bounds()
	w, h := b.Dx(), b.Dy()
	diff, overlay := image.NewNRGBA(image.Rect(0, 0, w, h)), image.NewNRGBA(image.Rect(0, 0, w, h))
	r := report{Width: w, Height: h, ReferenceSHA256: hex.EncodeToString(refHash[:]), CandidateSHA256: hex.EncodeToString(candHash[:]), Algorithm: "Pixels are converted to straight-alpha 8-bit NRGBA. Exact mismatches compare all RGBA channels. RGB error is absolute per-channel difference; per-pixel thresholds use the maximum RGB delta. Difference RGB is abs(delta)*4 clamped to 255 (unchanged pixels black); overlay averages all four converted RGBA channels. This is a pixel comparison and does not certify visual fidelity."}
	var total uint64
	minX, minY, maxX, maxY := w, h, -1, -1
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Convert both sources to the same straight-alpha representation. This
			// handles PNG palette, grayscale, and premultiplied Go image types consistently.
			a := color.NRGBAModel.Convert(ref.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			c := color.NRGBAModel.Convert(cand.At(cand.Bounds().Min.X+x, cand.Bounds().Min.Y+y)).(color.NRGBA)
			dr, dg, db := abs8(a.R, c.R), abs8(a.G, c.G), abs8(a.B, c.B)
			md := max8(dr, dg, db)
			if a != c {
				r.ExactMismatchCount++
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
			if md > r.MaxDelta {
				r.MaxDelta = md
			}
			if md > 1 {
				r.PixelsOver1++
			}
			if md > 3 {
				r.PixelsOver3++
			}
			if md > 10 {
				r.PixelsOver10++
			}
			total += uint64(dr) + uint64(dg) + uint64(db)
			diff.SetNRGBA(x, y, color.NRGBA{R: gain(dr), G: gain(dg), B: gain(db), A: 255})
			overlay.SetNRGBA(x, y, color.NRGBA{R: uint8((uint16(a.R) + uint16(c.R)) / 2), G: uint8((uint16(a.G) + uint16(c.G)) / 2), B: uint8((uint16(a.B) + uint16(c.B)) / 2), A: uint8((uint16(a.A) + uint16(c.A)) / 2)})
		}
	}
	if w*h > 0 {
		r.ExactMismatchFraction = float64(r.ExactMismatchCount) / float64(w*h)
		r.MeanAbsoluteRGBError = float64(total) / float64(w*h*3)
	}
	if maxX >= 0 {
		box := boundingBox{X: minX, Y: minY, Width: maxX - minX + 1, Height: maxY - minY + 1}
		r.MismatchBounds = &box
	}
	return result{r, diff, overlay}
}

func abs8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
func max8(a, b, c uint8) uint8 {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}
func gain(v uint8) uint8 {
	n := uint16(v) * diffGain
	if n > 255 {
		return 255
	}
	return uint8(n)
}

func writeJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write report %q: %w", path, err)
	}
	return nil
}
func writePNG(path string, img image.Image) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create image %q: %w", path, err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return fmt.Errorf("encode image %q: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close image %q: %w", path, err)
	}
	return nil
}
