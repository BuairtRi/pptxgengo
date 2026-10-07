package installstate

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type UninstallReport struct {
	Selection string `json:"selection"`
	Skill     string `json:"skill"`
	Retained  string `json:"retained"`
	Fonts     string `json:"fonts"`
}

// Uninstall deactivates only owned bindings, restores the original user skill
// when unchanged, and retains releases/backups. User projects and fonts are
// never removed. Edited skills and unowned launchers remain in place.
func (c Config) Uninstall(keepSkill bool) (UninstallReport, error) {
	r := UninstallReport{Selection: "inactive", Skill: "not_managed", Retained: "immutable releases, original skill backups and authored projects", Fonts: "not_changed"}
	e := c.mutate(func(c Config) error {
		before, e := loadState(c.Root)
		if e != nil {
			return e
		}
		after := State{Schema: before.Schema, OriginalSkills: map[string]OriginalSkill{}}
		for path, original := range before.OriginalSkills {
			after.OriginalSkills[path] = original
		}
		delete(after.OriginalSkills, c.SkillDir)
		if keepSkill {
			after.OriginalSkills = nil
		}
		tx := transaction{Schema: "pptxgengo.activation/v1", Before: before, After: after}
		if original, ok := before.OriginalSkills[c.SkillDir]; ok && !keepSkill {
			current, e := optionalSkillHash(c.SkillDir)
			if e != nil {
				return e
			}
			if current != "" && current != original.ManagedHash {
				r.Skill = "user_modifications_preserved"
			} else if current == "" {
				r.Skill = "already_absent"
			} else {
				id := token()
				tx.SkillPath = c.SkillDir
				tx.SkillStage = c.SkillDir + ".stage-" + id
				tx.SkillBackup = c.SkillDir + ".backup-" + id
				tx.BeforeSkillHash = current
				tx.AfterSkillHash = original.Hash
				defer func() {
					if _, e := os.Lstat(filepath.Join(c.Root, "pending.json")); os.IsNotExist(e) {
						os.RemoveAll(tx.SkillStage)
					}
				}()
				if original.Hash != "" {
					if !strings.HasPrefix(original.Backup, c.SkillDir+".backup-") {
						return fmt.Errorf("original skill backup is outside the owned backup path")
					}
					h, e := optionalSkillHash(original.Backup)
					if e != nil {
						return e
					}
					if h != original.Hash {
						return fmt.Errorf("original user skill backup changed; preserved")
					}
					info, e := os.Lstat(original.Backup)
					if e != nil {
						return e
					}
					if info.Mode()&os.ModeSymlink != 0 {
						target, e := os.Readlink(original.Backup)
						if e != nil {
							return e
						}
						if e = os.Symlink(target, tx.SkillStage); e != nil {
							return e
						}
					} else if e = copyTree(original.Backup, tx.SkillStage); e != nil {
						return e
					}
					h, e = optionalSkillHash(tx.SkillStage)
					if e != nil || h != original.Hash {
						return fmt.Errorf("original skill changed while staging: %v", e)
					}
					r.Skill = "original_user_skill_restored"
				} else {
					r.Skill = "owned_skill_removed; last installed copy retained as backup"
				}
				defer func() {
					if _, e := os.Lstat(filepath.Join(c.Root, "pending.json")); os.IsNotExist(e) {
						os.RemoveAll(tx.SkillStage)
					}
				}()
			}
		} else if keepSkill {
			r.Skill = "preserved_by_request"
		}
		// Multiple custom skill locations cannot be silently forgotten.
		for path := range before.OriginalSkills {
			if path != c.SkillDir && !keepSkill {
				return fmt.Errorf("another skill location is recorded: %s; use --keep-skill to preserve all skills", path)
			}
		}
		if runtime.GOOS == "windows" {
			tx.BeforePath, e = userPath()
			if e != nil {
				return e
			}
			tx.AfterPath = tx.BeforePath
			parts := []string{}
			for _, part := range strings.Split(tx.BeforePath.Value, ";") {
				if !strings.EqualFold(filepath.Clean(part), filepath.Join(c.Root, "bin")) {
					parts = append(parts, part)
				}
			}
			tx.AfterPath.Value = strings.Join(parts, ";")
			tx.ChangePath = tx.BeforePath != tx.AfterPath
		} else {
			for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
				path := filepath.Join(c.BinDir, tool)
				target, e := os.Readlink(path)
				if os.IsNotExist(e) {
					continue
				}
				if e != nil {
					continue
				}
				if target == filepath.Join(c.Root, "bin", tool) {
					old := target
					tx.Links = append(tx.Links, LinkChange{Path: path, Before: &old})
				}
			}
		}
		return c.commit(tx)
	})
	return r, e
}
