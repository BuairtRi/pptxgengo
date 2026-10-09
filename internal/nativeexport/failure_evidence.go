package nativeexport

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FailureEvidence describes observations, not a diagnosis of macOS privacy
// settings or the calling agent's sandbox. In particular -9074 is not evidence
// of a denied permission, and a basic Apple event is not proof of file access.
type FailureEvidence struct {
	SchemaVersion   int    `json:"schema_version"`
	RecordedAt      string `json:"recorded_at"`
	TaskID          string `json:"task_id,omitempty"`
	Operation       string `json:"operation"`
	Phase           string `json:"phase"`
	LastHelperPhase string `json:"last_helper_phase,omitempty"`
	Layer           string `json:"layer"`
	Classification  string `json:"classification"`
	ErrorCode       int    `json:"error_code,omitempty"`
	CauseConfirmed  bool   `json:"cause_confirmed"`
	Message         string `json:"message"`
}

var nativePhasePattern = regexp.MustCompile(`native_phase=(prepare_identity|open_document|identify_document|export_pdf|close_document);`)
var nativeEnteredPhasePattern = regexp.MustCompile(`native_phase_entered=(prepare_identity|open_document|identify_document|export_pdf|close_document);`)
var nativeObservedPhasePattern = regexp.MustCompile(`(?m)^native_helper_entered=(prepare_identity|open_document|identify_document|export_pdf|close_document);$`)
var nativeLastPhasePattern = regexp.MustCompile(`native_last_helper_phase=(prepare_identity|open_document|identify_document|export_pdf|close_document);`)

// An entry marker proves only that the helper reached this phase. It does not
// identify the failing call, establish completion, or diagnose a permission.
func lastHelperPhase(message string) string {
	matches := nativeEnteredPhasePattern.FindAllStringSubmatch(message, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1][1]
}

var nativeErrorCodePattern = regexp.MustCompile(`\((-?[0-9]+)\)`)
var nativeExplicitErrorCodePattern = regexp.MustCompile(`native_error_code=(-?[0-9]+);`)

func classifyNativeFailure(message, taskID string) FailureEvidence {
	evidence := FailureEvidence{SchemaVersion: 1, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), TaskID: taskID, Operation: "native_render", Phase: "unknown", Layer: "unknown", Classification: "unclassified", Message: message}
	// The parent snapshots primary-operation progress before cleanup. Prefer
	// its first snapshot so later cleanup errors cannot replace that evidence.
	if match := nativeLastPhasePattern.FindStringSubmatch(message); len(match) == 2 {
		evidence.LastHelperPhase = match[1]
	}
	if match := nativePhasePattern.FindStringSubmatch(message); len(match) == 2 {
		evidence.Phase = match[1]
		evidence.Layer = "powerpoint_document"
	}
	// Prefer the script's explicit code; legacy helpers use parenthesized codes.
	if match := nativeExplicitErrorCodePattern.FindStringSubmatch(message); len(match) == 2 {
		evidence.ErrorCode, _ = strconv.Atoi(match[1])
	} else if matches := nativeErrorCodePattern.FindAllStringSubmatch(message, -1); len(matches) > 0 {
		evidence.ErrorCode, _ = strconv.Atoi(matches[len(matches)-1][1])
	}
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "-1743") || strings.Contains(lower, "not authorized") || strings.Contains(lower, "automation denied"):
		evidence.Layer, evidence.Classification = "apple_events", "automation_denied"
	case strings.Contains(lower, "-10827"):
		evidence.Layer, evidence.Classification = "application_dispatch", "application_dispatch_failed"
	case evidence.Phase == "prepare_identity" && strings.Contains(lower, "file_access_denied"):
		evidence.Layer, evidence.Classification = "filesystem_access", "filesystem_permission_denied"
	case strings.Contains(lower, "file_access_denied"):
		evidence.Layer, evidence.Classification = "powerpoint_file_access", "file_access_blocked"
	case strings.Contains(lower, "permission denied"):
		evidence.Layer, evidence.Classification = "filesystem_access", "filesystem_permission_denied"
	case strings.Contains(lower, "identity_ambiguous") || strings.Contains(lower, "identity_changed"):
		evidence.Layer, evidence.Classification = "document_identity", "identity_unconfirmed"
	case strings.Contains(lower, "open_identity_timeout"):
		evidence.Classification = "document_open_or_identity_timeout"
	case strings.Contains(lower, "deadline exceeded") || strings.Contains(lower, "-1712"):
		evidence.Classification = "deadline_exceeded"
	case evidence.Phase == "open_document":
		evidence.Classification = "document_open_failed"
	case evidence.Phase == "identify_document":
		evidence.Classification = "document_identity_failed"
	case evidence.Phase == "export_pdf":
		evidence.Classification = "pdf_export_failed"
	case evidence.Phase == "close_document":
		evidence.Classification = "document_close_failed"
	}
	return evidence
}
