package nativeexport

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

const taskOwnershipMarker = "stable-files-owned"
const taskIdentityFile = "presentation-identity.pptx"

// PowerPoint-visible files live directly in the stable authorization folder.
// Random task subdirectories scope app grants too narrowly to qualify the next
// render. Scripts and rasterization intermediates remain private per task.
func taskPresentationPaths(root, taskID string) (string, string) {
	return filepath.Join(root, taskID+".pptx"), filepath.Join(root, taskID+".pdf")
}

func taskExportArguments(script, pptx, pdf, taskID string, seconds int) []string {
	return []string{script, pptx, pdf, taskID + ".pptx", fmt.Sprint(seconds), "export", filepath.Join(filepath.Dir(script), taskIdentityFile)}
}

func taskCloseArguments(script, pptx, pdf, taskID string, seconds int, stable bool) []string {
	args := []string{script, pptx, pdf, taskID + ".pptx", fmt.Sprint(seconds), "close"}
	if stable {
		args = append(args, filepath.Join(filepath.Dir(script), taskIdentityFile))
	}
	return args
}

func writeTaskPresentation(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create exact task presentation: %w", err)
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	if err != nil {
		_ = os.Remove(path)
		return err
	}
	if closeErr != nil {
		_ = os.Remove(path)
	}
	return closeErr
}

// Acquire names before calling PowerPoint. Existing files and symlinks are
// collisions, never export destinations or exact-task cleanup candidates.
func acquireTaskFiles(root, taskID string, data []byte) (string, string, error) {
	pptx, pdf := taskPresentationPaths(root, taskID)
	work := filepath.Join(root, taskID)
	if err := writeTaskPresentation(pptx, data); err != nil {
		return "", "", err
	}
	pdfFile, err := os.OpenFile(pdf, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		_ = os.Remove(pptx)
		return "", "", fmt.Errorf("reserve exact task PDF: %w", err)
	}
	if err = pdfFile.Close(); err == nil {
		err = os.Link(pptx, filepath.Join(work, taskIdentityFile))
	}
	if err == nil {
		err = os.WriteFile(filepath.Join(work, taskOwnershipMarker), []byte(taskID), 0600)
	}
	if err != nil {
		_ = os.Remove(pptx)
		_ = os.Remove(pdf)
		return "", "", err
	}
	return pptx, pdf, nil
}

func stableTaskOwned(root, taskID string) (bool, error) {
	marker := filepath.Join(root, taskID, taskOwnershipMarker)
	info, err := os.Lstat(marker)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("task ownership metadata is not a regular file")
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		return false, err
	}
	if !bytes.Equal(data, []byte(taskID)) {
		return false, fmt.Errorf("task ownership metadata does not match task identity")
	}
	return true, nil
}

func exactTaskPresentation(root, taskID string) (string, error) {
	pptx, _ := taskPresentationPaths(root, taskID)
	info, err := os.Lstat(pptx)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("task presentation is not a regular file")
	}
	identity, err := os.Lstat(filepath.Join(root, taskID, taskIdentityFile))
	if err != nil {
		return "", err
	}
	if !identity.Mode().IsRegular() || !os.SameFile(info, identity) {
		return "", fmt.Errorf("task presentation inode changed; exact close is unconfirmed")
	}
	return pptx, nil
}

func validateTaskExport(root, taskID string) error {
	if _, err := exactTaskPresentation(root, taskID); err != nil {
		return err
	}
	_, pdf := taskPresentationPaths(root, taskID)
	info, err := os.Lstat(pdf)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() != 0 {
		return fmt.Errorf("task PDF reservation changed before export")
	}
	return nil
}

func removeTaskFiles(root, taskID string) error {
	owned, err := stableTaskOwned(root, taskID)
	if err != nil {
		return err
	}
	pptx, pdf := taskPresentationPaths(root, taskID)
	if owned {
		if _, err = exactTaskPresentation(root, taskID); err != nil && !os.IsNotExist(err) {
			return err
		}
		if info, statErr := os.Lstat(pdf); statErr == nil && !info.Mode().IsRegular() {
			return fmt.Errorf("task PDF is not a regular file; retained exact task files")
		} else if statErr != nil && !os.IsNotExist(statErr) {
			return statErr
		}
		for _, path := range []string{pptx, pdf} {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return os.RemoveAll(filepath.Join(root, taskID))
}
