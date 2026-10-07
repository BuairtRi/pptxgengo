GO ?= go
# The full short suite includes pinned catalog/index verification; slower
# Macs need a separate package ceiling from the small, selected race workloads.
FAST_TEST_TIMEOUT ?= 5m
INTEGRATION_TEST_TIMEOUT ?= 10m
RACE_TEST_TIMEOUT ?= 150s
FULL_RACE_TEST_TIMEOUT ?= 10m
HEADLESS_TEST_ENV = PPTXGENGO_NATIVE_LIVE_OUT= PPTXGENGO_ROUNDTRIP_PREPARE_OUT= PPTXGENGO_ROUNDTRIP_FIXTURE= PPTXGENGO_ROUNDTRIP_SAVED_AS= PPTXGENGO_ROUNDTRIP_EDITED= PPTXGENGO_ROUNDTRIP_VERIFY_OUT= PPTXGENGO_ROUNDTRIP_WINDOWS_OUT= PPTXGENGO_SEARCH_BENCH_OUT= PPTXGENGO_INSTALL_PROCESS_OUT= PPTXGENGO_NATIVE_DIAGRAM_FIXTURE= PPTXGENGO_NATIVE_DIAGRAM_EDITED= PPTXGENGO_NATIVE_DIAGRAM_VERIFY_OUT= PPTXGENGO_EDITABLE_LIST_PILOT_OUT= PPTXGENGO_NATIVE_COMPONENT_DEMO_OUT= PPTXGENGO_NATIVE_COMPONENT_DEMO_VERIFY_ROOT=

.PHONY: build test test-pr test-race test-integration test-race-full test-native test-roundtrip-prepare test-roundtrip-verify test-roundtrip-windows test-search-performance test-installation-process test-diagram-verify

build:
	CGO_ENABLED=0 $(GO) build -trimpath -o bin/ ./cmd/pptxgengo ./cmd/pptxdesign ./cmd/wmdsdocs

# Everyday checks omit only explicitly marked exhaustive/private-asset tests.
test:
	$(HEADLESS_TEST_ENV) $(GO) test -short -count=1 -timeout=$(FAST_TEST_TIMEOUT) ./...

# Focused PR units plus compilation of all packages. No race detector, model
# download, source-deck rendering, actual installers, or Office qualification.
test-pr:
	$(HEADLESS_TEST_ENV) $(GO) test -short -run '^$$' -timeout=$(FAST_TEST_TIMEOUT) ./...
	$(HEADLESS_TEST_ENV) $(GO) test -short -count=1 -timeout=150s ./pptx ./internal/installstate ./internal/modelpackage ./internal/localembed ./internal/finishedslide ./internal/powershellenv ./scripts/cmd/ci-lint

# Race the source-mutation guard, native-worker cleanup paths, library cache,
# and the library's concurrent presentation serialization regressions.
test-race:
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) ./internal/installstate
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) ./internal/modelpackage ./scripts/cmd/release-ci ./scripts/cmd/search-benchmark
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) ./internal/finishedslide
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run 'TestSourceMutationsShareGuard|TestConcurrentSourceMutationGuard|TestSectionMutationCommentsAndAtomicity|TestFinishedSlide|TestObservedDependency|TestNative|TestTypedCardNative|TestTypedCardField' -skip '^(TestNativeEditability|TestNativeRoundTrip|TestFinishedSlideClaims)' ./internal/deckproject
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run '^TestFinishedSlideClaims' ./internal/deckproject
	$(HEADLESS_TEST_ENV) $(GO) test -race -short -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run '^TestPortable' ./internal/deckproject
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run '^TestReconcile' ./internal/deckproject
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run '^TestNativeEditability' ./internal/deckproject
	$(HEADLESS_TEST_ENV) $(GO) test -race -short -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run '^TestNativeRoundTrip' ./internal/deckproject
	$(HEADLESS_TEST_ENV) $(GO) test -race -short -count=1 -timeout=$(RACE_TEST_TIMEOUT) ./internal/nativeexport
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run 'TestConcurrentWriteRace|TestConcurrentAddChartRace|TestNativeConnector' ./pptx
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run '^TestNativeEditing' ./internal/wmdesign
	@for check in ConcurrentColdAndWarm CloneIsolation FingerprintDependencies BoundedEvictionAndErrors DriftAfterWarm; do \
		$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run "^TestLibraryCatalogCache$$check$$" ./internal/wmdesign || exit $$?; \
	done
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run '^TestKeywordConcurrentReadOnlyQueries$$' ./internal/wmdesign

# Includes catalog-wide and relocated-deck checks. Requires registered private
# branding assets at WMDS_BRANDING_ROOT (or ~/Documents/branding).
test-integration:
	$(HEADLESS_TEST_ENV) $(GO) test -count=1 -timeout=$(INTEGRATION_TEST_TIMEOUT) ./...

# Opt-in exhaustive race pass. The per-package timeout is a hard ceiling.
test-race-full:
	$(HEADLESS_TEST_ENV) $(GO) test -race -count=1 -timeout=$(FULL_RACE_TEST_TIMEOUT) ./...

# Live PowerPoint is separate from hermetic tests and requires a new output path.
test-native:
	test -n "$$PPTXGENGO_NATIVE_LIVE_OUT" || (echo 'set PPTXGENGO_NATIVE_LIVE_OUT to a new output directory' >&2; exit 1)
	test ! -e "$$PPTXGENGO_NATIVE_LIVE_OUT" || (echo 'PPTXGENGO_NATIVE_LIVE_OUT must not already exist' >&2; exit 1)
	$(GO) test -count=1 -timeout=2m -run '^TestNativeLiveSmoke$$' ./internal/nativeexport

# Only the Windows action opens Office. Explicit path checks prevent a skipped
# opt-in test from reporting a successful Make target.
test-roundtrip-prepare:
	test -n "$$PPTXGENGO_ROUNDTRIP_PREPARE_OUT" || (echo 'set PPTXGENGO_ROUNDTRIP_PREPARE_OUT to a new directory' >&2; exit 1)
	$(GO) test -count=1 -timeout=2m -run '^TestNativeRoundTripPrepare$$' ./internal/deckproject

test-roundtrip-verify:
	test -n "$$PPTXGENGO_ROUNDTRIP_FIXTURE" -a -n "$$PPTXGENGO_ROUNDTRIP_SAVED_AS" -a -n "$$PPTXGENGO_ROUNDTRIP_EDITED" -a -n "$$PPTXGENGO_ROUNDTRIP_VERIFY_OUT" || (echo 'set the four round-trip verification paths; see docs/native-roundtrip.md' >&2; exit 1)
	$(GO) test -count=1 -timeout=3m -run '^TestNativeRoundTripVerify$$' ./internal/deckproject

test-roundtrip-windows:
	test -n "$$PPTXGENGO_ROUNDTRIP_WINDOWS_OUT" || (echo 'set PPTXGENGO_ROUNDTRIP_WINDOWS_OUT to a new directory on an interactive Windows desktop' >&2; exit 1)
	$(GO) test -count=1 -timeout=3m -run '^TestNativeRoundTripLiveWindows$$' ./internal/deckproject

# Offline diagnostics never open Office; fixture/model preparation is untimed.
test-search-performance:
	test -n "$$PPTXGENGO_EMBED_MODEL_DIR" -a -n "$$PPTXGENGO_SEARCH_BENCH_OUT" || (echo 'set existing model directory and new benchmark output directory; see docs/search-performance.md' >&2; exit 1)
	test ! -e "$$PPTXGENGO_SEARCH_BENCH_OUT" || (echo 'benchmark output must not already exist' >&2; exit 1)
	$(GO) test -count=1 -timeout=7m -run '^TestPinnedLibrarySearchPerformance$$' -v ./internal/wmdesign

test-installation-process:
	test -n "$$PPTXGENGO_INSTALL_PROCESS_OUT" || (echo 'set PPTXGENGO_INSTALL_PROCESS_OUT to a new qualification directory; see docs/installation.md' >&2; exit 1)
	test ! -e "$$PPTXGENGO_INSTALL_PROCESS_OUT" || (echo 'qualification output must not already exist' >&2; exit 1)
	$(GO) test -count=1 -timeout=8m -run '^TestInstallationRealToolProcesses$$' -v ./internal/installstate

# Read-only native diagram evidence; opening/moving/editing remains a desktop task.
test-diagram-verify:
	test -n "$$PPTXGENGO_NATIVE_DIAGRAM_FIXTURE" -a -n "$$PPTXGENGO_NATIVE_DIAGRAM_EDITED" -a -n "$$PPTXGENGO_NATIVE_DIAGRAM_VERIFY_OUT" || (echo 'set all three diagram fixture/edited/new evidence paths; see docs/native-editing-pilot.md' >&2; exit 1)
	test ! -e "$$PPTXGENGO_NATIVE_DIAGRAM_VERIFY_OUT" || (echo 'diagram evidence destination must be new' >&2; exit 1)
	$(GO) test -count=1 -timeout=3m -run '^TestNativeEditabilityDiagramSuppliedCopy$$' -v ./internal/deckproject
