package installstate

import (
	"os"
	"strings"
	"testing"
)

// Set PPTXGENGO_TEST_TEMPLATE_PACKAGE_ROOT to verify a locally assembled
// templates-only CLI archive with real target binaries and build evidence.
func TestStagedTemplateOnlyArchiveVerify(t *testing.T) {
	root := os.Getenv("PPTXGENGO_TEST_TEMPLATE_PACKAGE_ROOT")
	if root == "" { t.Skip("staged template package integration not requested") }
	p, err := Verify(root)
	if err != nil { t.Fatal(err) }
	if p.Kind != "cli-only" || p.Bundle != "v11" || p.SourceRevision != "wmds-library.v11" || len(p.Files) < 2500 {
		t.Fatalf("template-only package metadata incomplete: kind=%s bundle=%s source=%s files=%d",p.Kind,p.Bundle,p.SourceRevision,len(p.Files))
	}
	for _, name := range []string{"library/wm-design-system/v11/library.sqlite", "library/wm-design-system/v11/fonts/IBMPlexSans-Regular.ttf", "browsing/template-library.pptx", "browsing/native-editing-coverage.json"} {
		if _, ok := p.Files[name]; !ok { t.Fatalf("verified package omitted %s",name) }
	}
	for name := range p.Files { if strings.Contains(name,"reusable-slides") || strings.Contains(name,"catalog/assets/") { t.Fatalf("deferred payload present: %s",name) } }
}
