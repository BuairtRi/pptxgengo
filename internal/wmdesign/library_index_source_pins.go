package wmdesign

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Frozen candidates may share unchanged source files with an older registered
// bundle. This exception applies only to exact inventory source pins; generic
// bundle, gallery and legacy resources still use indexRelative.
type indexSourcePinResolver struct {
	root    string
	files   map[string]string
	targets map[string]map[string]string
}

func newIndexSourcePinResolver(source *Source) *indexSourcePinResolver {
	resolver := &indexSourcePinResolver{root: source.Root, files: map[string]string{}, targets: map[string]map[string]string{}}
	for _, file := range source.Files {
		resolver.files[file.Path] = file.SHA256
	}
	return resolver
}

func (resolver *indexSourcePinResolver) resolve(pin LibraryIndexPin) (string, error) {
	if expected, known := resolver.files[pin.Path]; !known || expected != pin.SHA256 {
		return "", fmt.Errorf("index.unverified_source_pin: %s", pin.Path)
	}
	path, err := indexRelative(resolver.root, pin.Path)
	if err == nil || !strings.HasPrefix(err.Error(), "index.symlink_resource:") {
		return path, err
	}
	// Preserve rejection unless the final target has this exact relative path
	// in a registered bundle's hash-pinned inventory. EvalSymlinks also handles
	// candidate chains such as V7 -> V6 -> V5 and relocated bundle trees.
	target, resolveErr := filepath.EvalSymlinks(filepath.Join(resolver.root, filepath.FromSlash(pin.Path)))
	if resolveErr != nil {
		return "", resolveErr
	}
	suffix := string(filepath.Separator) + filepath.Join("source", filepath.FromSlash(pin.Path))
	if !strings.HasSuffix(target, suffix) {
		return "", err
	}
	root := strings.TrimSuffix(target, suffix)
	files, known := resolver.targets[root]
	if !known {
		files = registeredIndexSourceFiles(root)
		resolver.targets[root] = files
	}
	if files[pin.Path] != pin.SHA256 {
		return "", err
	}
	info, statErr := os.Lstat(target)
	if statErr != nil {
		return "", statErr
	}
	if !info.Mode().IsRegular() {
		return "", err
	}
	hash, hashErr := indexFileDigest(target)
	if hashErr != nil {
		return "", hashErr
	}
	if hash != pin.SHA256 {
		return "", fmt.Errorf("index.stale_input: source/%s", pin.Path)
	}
	return target, nil
}

// Attest the target metadata using the same registered manifest and inventory
// hashes as Load. Only the requested source bytes need reading here; Load has
// already validated the current bundle and all of its declared source inputs.
func registeredIndexSourceFiles(root string) map[string]string {
	bundlePath, err := indexRelative(root, "bundle.json")
	if err != nil {
		return nil
	}
	bundle, err := os.ReadFile(bundlePath)
	if err != nil {
		return nil
	}
	pin, known := sourcePins[indexDigest(bundle)]
	if !known {
		return nil
	}
	inventoryPath, err := indexRelative(root, "inventory.json")
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(inventoryPath)
	if err != nil || indexDigest(data) != pin.Inventory {
		return nil
	}
	var inventory Inventory
	if json.Unmarshal(data, &inventory) != nil || (inventory.Revision != "" && inventory.Revision != pin.Revision) {
		return nil
	}
	files := make(map[string]string, len(inventory.Sources))
	for _, file := range inventory.Sources {
		files[file.Path] = file.SHA256
	}
	return files
}
