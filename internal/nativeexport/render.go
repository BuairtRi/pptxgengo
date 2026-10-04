// Package nativeexport renders a task copy through local Microsoft PowerPoint.
// It never substitutes another renderer for PowerPoint.
package nativeexport

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

//go:embed powerpoint.applescript
var exportScript []byte

//go:embed pdf.swift
var pdfScript []byte

type Options struct {
	PPTX, Out               string
	PDF, PNG, IncludeHidden bool
	Timeout                 time.Duration
}
type Artifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}
type Receipt struct {
	Renderer                string     `json:"renderer"`
	Source                  Artifact   `json:"source"`
	ReviewCopySHA256        string     `json:"review_copy_sha256"`
	IncludeHidden           bool       `json:"include_hidden"`
	HiddenSlidesMadeVisible []string   `json:"hidden_slides_made_visible"`
	Slides                  int        `json:"source_slides"`
	Pages                   int        `json:"exported_pages"`
	PDF                     *Artifact  `json:"pdf,omitempty"`
	PNGs                    []Artifact `json:"pngs,omitempty"`
}

type runner func(context.Context, string, ...string) ([]byte, error)

func command(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 2 * time.Second
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%s: %w: %s", name, ctx.Err(), strings.TrimSpace(string(output)))
		}
		return nil, fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func Render(ctx context.Context, opts Options) (*Receipt, error) {
	return render(ctx, opts, command, runtime.GOOS)
}
func render(ctx context.Context, opts Options, run runner, platform string) (_ *Receipt, err error) {
	if platform != "darwin" {
		return nil, fmt.Errorf("native rendering requires macOS with Microsoft PowerPoint and the Swift/PDFKit tools installed")
	}
	if opts.PPTX == "" || opts.Out == "" || (!opts.PDF && !opts.PNG) {
		return nil, fmt.Errorf("--pptx, --out, and at least one of --pdf or --png are required")
	}
	if opts.Timeout <= 0 {
		return nil, fmt.Errorf("timeout must be positive")
	}
	source, err := filepath.Abs(opts.PPTX)
	if err != nil {
		return nil, err
	}
	out, err := filepath.Abs(opts.Out)
	if err != nil {
		return nil, err
	}
	original, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}
	review, total, hidden, err := reviewCopy(original, opts.IncludeHidden)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return nil, err
	}
	if err = os.Mkdir(out, 0755); err != nil {
		return nil, fmt.Errorf("output directory must be new: %w", err)
	}
	// Keep a failure explanation, never a success receipt for an incomplete export.
	defer func() {
		if err != nil {
			_ = os.WriteFile(filepath.Join(out, "render-error.txt"), []byte(err.Error()+"\n"), 0644)
		}
	}()
	work, err := os.MkdirTemp(out, ".native-work-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(work)
	taskName := filepath.Base(work) + ".pptx"
	taskPath := filepath.Join(work, taskName)
	scriptPath := filepath.Join(work, "export.applescript")
	swiftPath := filepath.Join(work, "pdf.swift")
	for path, data := range map[string][]byte{taskPath: review, scriptPath: exportScript, swiftPath: pdfScript} {
		if err = os.WriteFile(path, data, 0600); err != nil {
			return nil, err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	pdfPath := filepath.Join(out, "deck.pdf")
	seconds := max(1, int(opts.Timeout.Seconds()))
	if _, err = run(ctx, "/usr/bin/osascript", scriptPath, taskPath, pdfPath, taskName, fmt.Sprint(seconds)); err != nil {
		// Resolve the exact task-copy path again; never close by display name.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, _ = run(cleanupCtx, "/usr/bin/osascript", scriptPath, taskPath, pdfPath, taskName, "3", "close")
		cancel()
		return nil, fmt.Errorf("PowerPoint PDF export failed (run from a logged-in macOS GUI session; allow terminal automation of PowerPoint): %w", err)
	}
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("PowerPoint returned without a PDF: %w", err)
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		return nil, fmt.Errorf("PowerPoint output is not a PDF")
	}
	mode := "inspect"
	if opts.PNG {
		mode = "png"
	}
	metadata, err := run(ctx, "/usr/bin/swift", swiftPath, pdfPath, filepath.Join(out, "native-pages"), mode)
	if err != nil {
		return nil, err
	}
	var result struct {
		Pages int `json:"pages"`
	}
	if err = json.Unmarshal(metadata, &result); err != nil {
		return nil, fmt.Errorf("PDFKit returned invalid metadata: %w", err)
	}
	expected := total
	if !opts.IncludeHidden {
		expected -= len(hidden)
	}
	if result.Pages != expected {
		return nil, fmt.Errorf("PowerPoint exported %d pages; expected %d (source %d slides, %d hidden)", result.Pages, expected, total, len(hidden))
	}
	receipt := &Receipt{Renderer: "Microsoft PowerPoint (local native PDF); macOS PDFKit PNG", Source: Artifact{Path: source, SHA256: hash(original)}, ReviewCopySHA256: hash(review), IncludeHidden: opts.IncludeHidden, HiddenSlidesMadeVisible: []string{}, Slides: total, Pages: result.Pages}
	if opts.IncludeHidden {
		receipt.HiddenSlidesMadeVisible = hidden
	}
	if opts.PDF {
		receipt.PDF = &Artifact{Path: "deck.pdf", SHA256: hash(pdfBytes)}
	} else {
		if err = os.Remove(pdfPath); err != nil {
			return nil, err
		}
	}
	if opts.PNG {
		for page := 1; page <= result.Pages; page++ {
			relative := fmt.Sprintf("native-pages/slide-%03d.png", page)
			data, e := os.ReadFile(filepath.Join(out, filepath.FromSlash(relative)))
			if e != nil {
				return nil, e
			}
			config, format, e := image.DecodeConfig(bytes.NewReader(data))
			if e != nil || format != "png" {
				return nil, fmt.Errorf("invalid PNG page %d: %v", page, e)
			}
			receipt.PNGs = append(receipt.PNGs, Artifact{Path: relative, SHA256: hash(data), Width: config.Width, Height: config.Height})
		}
	}
	// Detect even concurrent external source edits; the command never opens it.
	after, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(original, after) {
		return nil, fmt.Errorf("source changed during rendering; no successful receipt issued")
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(filepath.Join(out, "render-manifest.json"), append(data, '\n'), 0644); err != nil {
		return nil, err
	}
	return receipt, nil
}
func hash(data []byte) string { h := sha256.Sum256(data); return hex.EncodeToString(h[:]) }

// Match complete attributes, so text resembling an attribute inside another
// quoted value cannot be removed from the source.
var rootAttribute = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_.:-]*\s*=\s*(?:"[^"]*"|'[^']*')`)

func removeShow(tag []byte) []byte {
	for _, location := range rootAttribute.FindAllIndex(tag, -1) {
		attribute := tag[location[0]:location[1]]
		if strings.TrimSpace(string(attribute[:bytes.IndexByte(attribute, '=')])) == "show" {
			start := location[0]
			for start > 0 && (tag[start-1] == ' ' || tag[start-1] == '\t' || tag[start-1] == '\n' || tag[start-1] == '\r') {
				start--
			}
			result := append([]byte{}, tag[:start]...)
			return append(result, tag[location[1]:]...)
		}
	}
	return tag
}

func reviewCopy(original []byte, include bool) ([]byte, int, []string, error) {
	archive, err := zip.NewReader(bytes.NewReader(original), int64(len(original)))
	if err != nil {
		return nil, 0, nil, fmt.Errorf("invalid PPTX: %w", err)
	}
	ordered, err := presentationSlides(archive)
	if err != nil {
		return nil, 0, nil, err
	}
	referenced := map[string]bool{}
	for _, name := range ordered {
		referenced[name] = true
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	total := len(ordered)
	hiddenParts := map[string]bool{}
	for _, file := range archive.File {
		if !referenced[file.Name] {
			if err = writer.Copy(file); err != nil {
				return nil, 0, nil, err
			}
			continue
		}
		input, e := file.Open()
		if e != nil {
			return nil, 0, nil, e
		}
		data, e := io.ReadAll(input)
		input.Close()
		if e != nil {
			return nil, 0, nil, e
		}
		decoder := xml.NewDecoder(bytes.NewReader(data))
		var start xml.StartElement
		rootEnd := 0
		for {
			token, e := decoder.Token()
			if e != nil {
				return nil, 0, nil, fmt.Errorf("invalid slide XML %s: %w", file.Name, e)
			}
			if node, ok := token.(xml.StartElement); ok {
				start = node
				rootEnd = int(decoder.InputOffset())
				break
			}
		}
		if start.Name.Local != "sld" {
			return nil, 0, nil, fmt.Errorf("invalid slide root: %s", file.Name)
		}
		isHidden := false
		for _, a := range start.Attr {
			if a.Name.Local == "show" && a.Name.Space == "" && (a.Value == "0" || a.Value == "false") {
				isHidden = true
			}
		}
		if isHidden {
			hiddenParts[file.Name] = true
		}
		if !include || !isHidden {
			if err = writer.Copy(file); err != nil {
				return nil, 0, nil, err
			}
			continue
		}
		rootStart := bytes.LastIndexByte(data[:rootEnd], '<')
		changed := append([]byte{}, data[:rootStart]...)
		changed = append(changed, removeShow(data[rootStart:rootEnd])...)
		changed = append(changed, data[rootEnd:]...)
		header := file.FileHeader
		header.CRC32 = 0
		header.CompressedSize64 = 0
		header.UncompressedSize64 = 0
		destination, e := writer.CreateHeader(&header)
		if e != nil {
			return nil, 0, nil, e
		}
		if _, e = destination.Write(changed); e != nil {
			return nil, 0, nil, e
		}
	}
	hidden := []string{}
	for _, name := range ordered {
		if hiddenParts[name] {
			hidden = append(hidden, name)
		}
	}
	if err = writer.Close(); err != nil {
		return nil, 0, nil, err
	}
	if !include || len(hidden) == 0 {
		return original, total, hidden, nil
	}
	return buffer.Bytes(), total, hidden, nil
}
