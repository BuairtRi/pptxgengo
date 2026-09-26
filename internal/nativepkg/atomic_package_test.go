package nativepkg

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractFailureLeavesNoProject(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source.pptx")
	writeFixturePPTX(t, source)
	parts, err := readZip(source)
	if err != nil {
		t.Fatal(err)
	}
	parts["ppt/_rels/presentation.xml.rels"] = []byte(relationships(
		`rId2|http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide|slides/slide2.xml`,
	))
	broken := filepath.Join(tmp, "broken.pptx")
	writeZipParts(t, broken, parts)

	project := filepath.Join(tmp, "project")
	err = Extract(broken, project, []int{1})
	if err == nil || !strings.Contains(err.Error(), "missing slide relation") {
		t.Fatalf("Extract error = %v, want missing slide relation", err)
	}
	if _, err := os.Stat(project); !os.IsNotExist(err) {
		t.Errorf("failed extraction left project directory %s: %v", project, err)
	}
	if temporary, err := filepath.Glob(filepath.Join(tmp, ".project.tmp-*")); err != nil || len(temporary) != 0 {
		t.Errorf("failed extraction left staging directory: %v, err=%v", temporary, err)
	}
}

func TestBuildRejectsExistingReportWithoutOverwrite(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source.pptx")
	writeFixturePPTX(t, source)
	project := filepath.Join(tmp, "project")
	if err := Extract(source, project, []int{1}); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "rebuilt.pptx")
	reportPath := out + ".build.json"
	const previousReport = "keep this report exactly\n"
	if err := os.WriteFile(reportPath, []byte(previousReport), 0644); err != nil {
		t.Fatal(err)
	}
	err := Build(project, out, nil, false)
	if err == nil || !strings.Contains(err.Error(), "build report already exists") {
		t.Fatalf("Build error = %v, want existing report error", err)
	}
	got, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != previousReport {
		t.Errorf("existing report was changed: got %q, want %q", got, previousReport)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("Build wrote PPTX despite existing report: %v", err)
	}
}

func writeZipParts(t *testing.T, file string, parts map[string][]byte) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range parts {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}
