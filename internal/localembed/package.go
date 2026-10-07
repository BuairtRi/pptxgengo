package localembed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/modelpackage"
)

type PackageReport = modelpackage.Report

// PreparePackage is an explicit maintenance operation. It either downloads the
// pinned public model with TLS or copies an already verified offline directory.
// A staged directory is published without replacing an existing destination.
func PreparePackage(ctx context.Context, out, from string, download bool) (PackageReport, error) {
	report := PackageReport{Schema: "pptxgengo.offline-embedding-package.v1", Identity: PinnedIdentity(), Source: "https://huggingface.co/" + ModelID + "/tree/" + ModelRevision, License: "Apache-2.0"}
	if out == "" || (download && from != "") || (!download && from == "") {
		return report, fmt.Errorf("embedding.package_options: require --out NEW-DIR and exactly one of --download or --from DIR")
	}
	dest, err := filepath.Abs(out)
	if err != nil {
		return report, err
	}
	report.Directory = dest
	if _, err = os.Lstat(dest); !os.IsNotExist(err) {
		return report, fmt.Errorf("embedding.output_exists: %s", dest)
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return report, err
	}
	stage, err := os.MkdirTemp(filepath.Dir(dest), ".model-stage-")
	if err != nil {
		return report, err
	}
	defer os.RemoveAll(stage)
	client := &http.Client{Timeout: 5 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 || req.URL.Scheme != "https" {
			return fmt.Errorf("embedding.unsafe_redirect")
		}
		return nil
	}}
	for _, a := range Artifacts() {
		if err = ctx.Err(); err != nil {
			return report, err
		}
		var data []byte
		if from != "" {
			data, err = pinnedBytes(from, a)
		} else {
			data, err = downloadArtifact(ctx, client, a)
		}
		if err != nil {
			return report, err
		}
		if err = writeSynced(filepath.Join(stage, a.File), data); err != nil {
			return report, err
		}
	}
	// Verify every staged artifact, including exact bytes actually published.
	for _, a := range Artifacts() {
		if _, err = pinnedBytes(stage, a); err != nil {
			return report, err
		}
	}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return report, err
	}
	for name, data := range map[string][]byte{"manifest.json": append(b, '\n'), "LICENSE": modelpackage.License(), "README.txt": modelpackage.Readme()} {
		if err = writeSynced(filepath.Join(stage, name), data); err != nil {
			return report, err
		}
	}
	if err = publishDirectory(stage, dest); err != nil {
		return report, fmt.Errorf("embedding.publish_failed: %w", err)
	}
	return report, nil
}

func downloadArtifact(ctx context.Context, client *http.Client, a Artifact) ([]byte, error) {
	url := "https://huggingface.co/" + ModelID + "/resolve/" + ModelRevision + "/" + a.Source
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "pptxgengo-offline-model-maintenance")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding.download_status: %s: HTTP %d", a.File, resp.StatusCode)
	}
	if resp.ContentLength > a.Bytes {
		return nil, fmt.Errorf("embedding.download_size: %s", a.File)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, a.Bytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) != a.Bytes || digest(data) != a.SHA256 {
		return nil, fmt.Errorf("embedding.download_hash_mismatch: %s", a.File)
	}
	return data, nil
}

func writeSynced(path string, data []byte) (err error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if e := f.Close(); err == nil {
			err = e
		}
	}()
	if _, err = f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}

// WriteSnapshot publishes a complete regular file without replacing any file,
// directory or symlink already at out. A hard link is atomic on supported local
// filesystems; an unsupported filesystem returns an error without clobbering.
func WriteSnapshot(out string, data []byte) error {
	if strings.TrimSpace(out) == "" {
		return fmt.Errorf("embedding.output_required")
	}
	dest, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(dest), ".embedding-stage-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Link(name, dest)
}
