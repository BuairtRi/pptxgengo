package deckproject

import (
	"os"
	"time"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func commitSourceSections(p *Project, change SectionChange, changes map[string][]byte, verifyDivider bool, bundle, engine string) (SectionChange, error) {
	change.BeforeSHA256 = p.SourceHash()
	change.Decision = "decisions/" + change.Operation + "-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	decision, err := SafePath(p.Root, change.Decision)
	if err != nil {
		return change, err
	}
	_, err = commitSourceChanges(p, changes, func(candidate *Project) error {
		if verifyDivider {
			if engine == "" {
				engine = wmdesign.CandidateEngine
			}
			var compiled Compilation
			var err error
			if _, _, lockErr := ReadLock(p); lockErr == nil {
				compiled, err = Check(candidate, bundle, engine)
			} else if os.IsNotExist(lockErr) {
				compiled, err = Compile(candidate, bundle, engine)
			} else {
				err = lockErr
			}
			if err != nil {
				return err
			}
			if _, _, err := wmdesign.BuildWithEngineAndAssets(bundle, "", compiled.Document, engine, compiled.Assets); err != nil {
				return err
			}
		}
		change.AfterSHA256, change.Sections = candidate.SourceHash(), ListSections(candidate)
		return writeJSON(decision, change)
	})
	if err != nil {
		os.Remove(decision)
	}
	return change, err
}
