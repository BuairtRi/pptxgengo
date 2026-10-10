// pptxgengo dispatches to the tools in a frozen local release.
package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var version = "dev"
var releaseIdentity = "dev"

const currentBundle = "v12"

var tools = map[string]string{
	"design": "pptxdesign",
	"docs":   "wmdsdocs",
}

var retiredRoutes = map[string]string{
	"template":  "`pptxgengo design template` for foundation templates or `pptxgengo design project` for maintained decks",
	"compose":   "`pptxgengo design build` for foundation documents or `pptxgengo design project build` for maintained decks",
	"scene":     "`pptxgengo design build` for foundation documents",
	"component": "`pptxgengo design library-find --kinds component` to discover components",
	"lib":       "`pptxgengo design library-find` to search the design library",
	"anchor":    "no supported replacement command; use `pptxgengo design project` for deck authoring",
	"diff":      "no supported replacement command",
	"adapt":     "no supported replacement command",
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: pptxgengo <design> <command> [flags]")
	fmt.Fprintln(os.Stderr, "       pptxgengo catalog [--templates|--design-system|--assets] [--print|--open]")
	fmt.Fprintln(os.Stderr, "       pptxgengo paths")
	fmt.Fprintln(os.Stderr, "       pptxgengo docs [--addr localhost:8787]")
	fmt.Fprintln(os.Stderr, "       pptxgengo package --binaries DIR --version VERSION --out NEW.zip [--arch amd64|arm64] [--without-photos]")
	fmt.Fprintln(os.Stderr, "       pptxgengo installation install|rollback|recover|doctor|uninstall [flags]")
	fmt.Fprintln(os.Stderr, "       pptxgengo --version")
}

func releaseRoot() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", err
	}
	return filepath.Dir(filepath.Dir(resolved)), nil
}

func hasRoot(args []string) bool {
	return hasFlag(args, "--root")
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			return true
		}
	}
	return false
}

// Resolve the packaged documentation independently of the caller's directory.
func docsArgs(root string, input []string) []string {
	args := append([]string{}, input...)
	if !hasFlag(args, "--dir") && !hasFlag(args, "-dir") {
		args = append(args, "--dir", filepath.Join(root, "wmds-docs", "site"))
	}
	return args
}

func publishedBundle(root string) (string, error) {
	path := filepath.Join(root, "release", "default-bundle.txt")
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return currentBundle, nil
	}
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(b))
	if value != currentBundle {
		return "", fmt.Errorf("invalid published bundle in %s: %q", path, value)
	}
	return value, nil
}

func designArgs(root string, input []string) ([]string, error) {
	args := append([]string{}, input...)
	if len(args) == 0 {
		return args, nil
	}
	command := args[0]
	if command == "library-model" || command == "render" || command == "render-doctor" || command == "render-native-worker" || command == "source-inventory" || command == "asset-gallery" {
		return args, nil
	}
	if command != "project" && !hasFlag(args, "--bundle") {
		bundle, err := publishedBundle(root)
		if err != nil {
			return nil, err
		}
		args = append(args, "--bundle", filepath.Join(root, "library", "wm-design-system", bundle))
	}
	indexRead := command == "library-embed" || command == "library-find" || command == "library-inspect" || command == "library-preview"
	if indexRead && !hasFlag(args, "--index") {
		bundlePath := ""
		for i, arg := range args {
			if arg == "--bundle" && i+1 < len(args) {
				bundlePath = args[i+1]
			} else if strings.HasPrefix(arg, "--bundle=") {
				bundlePath = strings.TrimPrefix(arg, "--bundle=")
			}
		}
		args = append(args, "--index", filepath.Join(bundlePath, "library.sqlite"))
		if !hasFlag(args, "--gallery") {
			args = append(args, "--gallery", filepath.Join(bundlePath, "catalog"))
		}
	}
	engineAllowed := command != "library-embed" && command != "library-index" && command != "library-inspect" && command != "library-preview" && command != "project" && command != "library-authoring"
	if engineAllowed && !hasFlag(args, "--engine") {
		args = append(args, "--engine", "wmds-go-foundation.v2")
	}
	return args, nil
}

func runCatalog(root string, args []string) error {
	open := false
	assets := false
	seenPage, seenAction := false, false
	for _, arg := range args {
		switch arg {
		case "--templates", "--design-system", "--assets":
			if seenPage {
				return fmt.Errorf("choose one catalog gallery")
			}
			seenPage = true
			assets = arg == "--assets"
		case "--components":
			return fmt.Errorf("component discovery is available through: pptxgengo design library-find --kinds component")
		case "--open", "--print":
			if seenAction {
				return fmt.Errorf("choose --open or --print")
			}
			seenAction = true
			open = arg == "--open"
		default:
			return fmt.Errorf("usage: pptxgengo catalog [--templates|--design-system|--assets] [--print|--open]")
		}
	}
	bundle, err := publishedBundle(root)
	if err != nil {
		return err
	}
	path := filepath.Join(root, "library", "wm-design-system", bundle, "catalog", "design-system.html")
	if assets {
		path = filepath.Join(root, "library", "wm-design-system", bundle, "catalog", "assets", "index.html")
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("catalog unavailable: %w", err)
	}
	if !open {
		fmt.Println(path)
		return nil
	}
	name, args := catalogOpenCommand(path, runtime.GOOS)
	return exec.Command(name, args...).Run()
}

// Pass paths as arguments, never through cmd.exe or a shell. Catalog filenames
// may contain spaces, ampersands or other shell metacharacters.
func catalogOpenCommand(path, platform string) (string, []string) {
	switch platform {
	case "windows":
		uri := &url.URL{Scheme: "file", Path: "/" + strings.ReplaceAll(path, `\`, "/")}
		if strings.HasPrefix(uri.Path, "///") {
			parts := strings.SplitN(strings.TrimPrefix(uri.Path, "///"), "/", 2)
			uri.Host = parts[0]
			uri.Path = "/"
			if len(parts) == 2 {
				uri.Path += parts[1]
			}
		}
		return "rundll32.exe", []string{"url.dll,FileProtocolHandler", uri.String()}
	case "darwin":
		return "open", []string{path}
	default:
		return "xdg-open", []string{path}
	}
}

func toolFilename(tool, platform string) string {
	if platform == "windows" {
		return tool + ".exe"
	}
	return tool
}

func run() error {
	if len(os.Args) > 1 && os.Args[1] == "--installation-api" {
		fmt.Println("pptxgengo.installation-api/v1")
		return nil
	}
	if len(os.Args) > 1 && os.Args[1] == "installation" {
		if handled, err := runManagedInstallation(os.Args[1:]); handled || err != nil {
			return err
		}
		return runInstallation(os.Args[2:])
	}
	if handled, err := runManagedLauncher(os.Args[1:]); handled || err != nil {
		return err
	}
	if len(os.Args) < 2 {
		usage()
		return fmt.Errorf("missing command")
	}
	name := os.Args[1]
	if name == "--build-info" {
		fmt.Println(releaseIdentity)
		return nil
	}
	if next, retired := retiredRoutes[name]; retired {
		return fmt.Errorf("pptxgengo %s was removed; %s", name, next)
	}
	if name == "--version" || name == "version" {
		fmt.Println(version)
		return nil
	}
	if name == "--help" || name == "help" {
		usage()
		return nil
	}
	root, err := releaseRoot()
	if err != nil {
		return err
	}
	if name == "catalog" {
		return runCatalog(root, os.Args[2:])
	}
	if name == "package" {
		return runPackage(root, os.Args[2:])
	}
	if name == "paths" {
		if len(os.Args) != 2 {
			return fmt.Errorf("usage: pptxgengo paths")
		}
		bundle, err := publishedBundle(root)
		if err != nil {
			return err
		}
		paths := map[string]string{
			"root":                    root,
			"library":                 filepath.Join(root, "library"),
			"scripts":                 filepath.Join(root, "scripts"),
			"catalog":                 filepath.Join(root, "library", "wm-design-system", bundle, "catalog", "design-system.html"),
			"catalog_design_system":   filepath.Join(root, "library", "wm-design-system", bundle, "catalog", "design-system.html"),
			"design_system_" + bundle: filepath.Join(root, "library", "wm-design-system", bundle),
			"design_system_default":   filepath.Join(root, "library", "wm-design-system", bundle),
			"design_index":            filepath.Join(root, "library", "wm-design-system", bundle, "library.sqlite"),
			"project_example":         filepath.Join(root, "examples", "deck-project"),
			"skill":                   filepath.Join(root, "skills", "west-monroe-presentations", "SKILL.md"),
			"design_docs":             filepath.Join(root, "wmds-docs", "site"),
			"design_docs_source":      filepath.Join(root, "wmds-docs", "site", "SOURCE.json"),
		}
		return json.NewEncoder(os.Stdout).Encode(paths)
	}
	tool := tools[name]
	if tool == "" {
		for _, full := range tools {
			if name == full {
				tool = full
				break
			}
		}
	}
	if tool == "" {
		usage()
		return fmt.Errorf("unknown tool %q", name)
	}
	path := filepath.Join(root, "bin", toolFilename(tool, runtime.GOOS))
	args := append([]string{}, os.Args[2:]...)
	if tool == "wmdsdocs" {
		args = docsArgs(root, args)
	}
	if tool == "pptxdesign" && len(args) > 0 {
		args, err = designArgs(root, args)
		if err != nil {
			return err
		}
	}
	if (tool == "pptxtemplate" || tool == "pptxlib" || tool == "pptxadapt") && !hasRoot(args) && len(args) > 0 && !(tool == "pptxtemplate" && (args[0] == "apply-accent" || args[0] == "apply-gauge")) {
		args = append(args, "--root", root)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), "PPTXGENGO_RELEASE_ROOT="+root)
	if err := cmd.Run(); err != nil {
		if status, ok := err.(*exec.ExitError); ok {
			os.Exit(status.ExitCode())
		}
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "pptxgengo:", err)
		os.Exit(1)
	}
}
