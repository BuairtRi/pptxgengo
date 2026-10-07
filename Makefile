GO ?= go
# The full short suite includes pinned catalog/index verification; slower hosted
# Macs need a separate package ceiling from the small, selected race workloads.
FAST_TEST_TIMEOUT ?= 5m
INTEGRATION_TEST_TIMEOUT ?= 10m
RACE_TEST_TIMEOUT ?= 150s
FULL_RACE_TEST_TIMEOUT ?= 10m

.PHONY: build test test-race test-integration test-race-full test-native

build:
	CGO_ENABLED=0 $(GO) build -trimpath -o bin/ ./cmd/pptxgengo ./cmd/pptxdesign ./cmd/wmdsdocs

# Everyday checks omit only explicitly marked exhaustive/private-asset tests.
test:
	PPTXGENGO_NATIVE_LIVE_OUT= $(GO) test -short -count=1 -timeout=$(FAST_TEST_TIMEOUT) ./...

# Race the source-mutation guard, native-worker cleanup paths, library cache,
# and the library's concurrent presentation serialization regressions.
test-race:
	$(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) ./internal/installstate
	$(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run 'TestSourceMutationsShareGuard|TestConcurrentSourceMutationGuard|TestSectionMutationCommentsAndAtomicity|TestFinishedSlide|TestObservedDependency|TestNative|TestTypedCardNative|TestTypedCardField|TestReconcile' ./internal/deckproject
	PPTXGENGO_NATIVE_LIVE_OUT= $(GO) test -race -short -count=1 -timeout=$(RACE_TEST_TIMEOUT) ./internal/nativeexport
	$(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run 'TestConcurrentWriteRace|TestConcurrentAddChartRace' ./pptx
	@for check in ConcurrentColdAndWarm CloneIsolation FingerprintDependencies BoundedEvictionAndErrors DriftAfterWarm; do \
		$(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run "^TestLibraryCatalogCache$$check$$" ./internal/wmdesign || exit $$?; \
	done
	$(GO) test -race -count=1 -timeout=$(RACE_TEST_TIMEOUT) -run '^TestKeywordConcurrentReadOnlyQueries$$' ./internal/wmdesign

# Includes catalog-wide and relocated-deck checks. Requires registered private
# branding assets at WMDS_BRANDING_ROOT (or ~/Documents/branding).
test-integration:
	PPTXGENGO_NATIVE_LIVE_OUT= $(GO) test -count=1 -timeout=$(INTEGRATION_TEST_TIMEOUT) ./...

# Opt-in exhaustive race pass. The per-package timeout is a hard ceiling.
test-race-full:
	PPTXGENGO_NATIVE_LIVE_OUT= $(GO) test -race -count=1 -timeout=$(FULL_RACE_TEST_TIMEOUT) ./...

# Live PowerPoint is separate from hermetic tests and requires a new output path.
test-native:
	test -n "$$PPTXGENGO_NATIVE_LIVE_OUT" || (echo 'set PPTXGENGO_NATIVE_LIVE_OUT to a new output directory' >&2; exit 1)
	test ! -e "$$PPTXGENGO_NATIVE_LIVE_OUT" || (echo 'PPTXGENGO_NATIVE_LIVE_OUT must not already exist' >&2; exit 1)
	$(GO) test -count=1 -timeout=2m -run '^TestNativeLiveSmoke$$' ./internal/nativeexport
