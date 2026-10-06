package deckproject

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Migration changes only the toolchain pin. Copy, template definitions, build
// history and native acceptance remain owned by their existing workflows.
type Migration struct {
	Schema             string                            `json:"schema"`
	Status             string                            `json:"status"`
	From               string                            `json:"from"`
	To                 string                            `json:"to"`
	Backup             string                            `json:"backup,omitempty"`
	Fit                string                            `json:"fit"`
	Native             string                            `json:"native"`
	NextAction         string                            `json:"next_action"`
	DensityAdjustments []wmdesign.SlideDensityAdjustment `json:"density_adjustments,omitempty"`
}

func Migrate(p *Project, bundle, engine string, dryRun bool) (Migration, error) {
	r := Migration{Schema: "pptxgengo.project-migration.v1", Native: "review_required", NextAction: "Run project build, then native render and review before using the updated deck."}
	// Share the existing mutation/build guards; a migration must not change pins
	// underneath either operation. Dry runs take the same guards temporarily.
	for _, name := range []string{".deck-source-mutation.lock", ".project-build.lock"} {
		guard, err := SafePath(p.Root, name)
		if err != nil {
			return r, err
		}
		if err = writeExclusive(guard, []byte("project migrate\n"), 0600); err != nil {
			return r, fmt.Errorf("project busy: %s: %w", name, err)
		}
		defer os.Remove(guard)
	}
	old, oldBytes, err := ReadLock(p)
	if err != nil {
		return r, err
	}
	r.From = old.BundleRevision
	target, err := makeLock(bundle, engine)
	if err != nil {
		return r, fmt.Errorf("migration target unavailable; original lock unchanged: %w", err)
	}
	r.To = target.BundleRevision
	if err = ValidateEditorial(p); err != nil {
		return r, fmt.Errorf("migration blocked; original lock unchanged: %w", err)
	}
	c, err := Compile(p, bundle, engine)
	if err != nil {
		return r, fmt.Errorf("migration blocked by target template compatibility; original lock unchanged: %w", err)
	}
	_, report, err := wmdesign.BuildWithEngineAndAssets(bundle, "", c.Document, engine, c.Assets)
	if err != nil {
		return r, fmt.Errorf("migration blocked by target fit; original lock unchanged: %w", err)
	}
	r.Fit = "compiler_checks_passed_arbitrary_content_unqualified"
	r.DensityAdjustments = report.DensityAdjustments
	current, err := Load(p.SourcePath)
	if err != nil || current.SourceHash() != p.SourceHash() {
		return r, fmt.Errorf("project source changed during migration; original lock unchanged")
	}
	_, currentBytes, err := ReadLock(p)
	if err != nil || !bytes.Equal(currentBytes, oldBytes) {
		return r, fmt.Errorf("project lock changed during migration; no replacement performed")
	}
	if reflect.DeepEqual(old, target) {
		r.Status = "already_current"
		return r, nil
	}
	if dryRun {
		r.Status = "ready_to_migrate"
		return r, nil
	}
	path, err := SafePath(p.Root, p.Document.Toolchain.Lockfile)
	if err != nil {
		return r, err
	}
	stat, err := os.Stat(path)
	if err != nil {
		return r, err
	}
	r.Backup = filepath.ToSlash(p.Document.Toolchain.Lockfile) + ".pre-migrate-" + digest(oldBytes)[:12]
	backup, err := SafePath(p.Root, r.Backup)
	if err != nil {
		return r, err
	}
	if existing, e := os.ReadFile(backup); e == nil {
		if !bytes.Equal(existing, oldBytes) {
			return r, fmt.Errorf("migration backup collision: %s; original lock unchanged", r.Backup)
		}
	} else if !os.IsNotExist(e) {
		return r, e
	} else if err = writeExclusive(backup, oldBytes, 0444); err != nil {
		return r, err
	}
	temporary := path + ".migrate-" + nonce()
	defer os.Remove(temporary)
	if err = writeExclusive(temporary, canonical(target), stat.Mode().Perm()); err != nil {
		return r, err
	}
	if err = os.Rename(temporary, path); err != nil {
		return r, fmt.Errorf("migration could not replace lock; original retained: %w", err)
	}
	r.Status = "migrated"
	return r, nil
}
