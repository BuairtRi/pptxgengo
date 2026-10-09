package nativeexport

import (
	"bytes"
	"fmt"
	"io"
)

// phaseForwarder retains helper stderr for the returned command error, while
// forwarding only recognized, complete phase markers to the owning worker's
// stderr. Other helper output may contain customer paths/content and is not
// streamed. No filesystem access or app operation is required for reporting.
type phaseForwarder struct {
	output  bytes.Buffer
	pending []byte
	target  io.Writer
}

func (w *phaseForwarder) Write(data []byte) (int, error) {
	_, _ = w.output.Write(data)
	w.pending = append(w.pending, data...)
	for {
		end := bytes.IndexByte(w.pending, '\n')
		if end < 0 {
			// An ordinary long line is not a progress protocol frame.
			if len(w.pending) > 4096 {
				w.pending = nil
			}
			break
		}
		line := string(w.pending[:end])
		w.pending = w.pending[end+1:]
		if phase := lastHelperPhase(line); phase != "" {
			// Reporting failure must not turn a successful export into a failure.
			_, _ = fmt.Fprintf(w.target, "native_helper_entered=%s;\n", phase)
		}
	}
	return len(data), nil
}

func helperObservationSummary(stderr string) string {
	matches := nativeObservedPhasePattern.FindAllStringSubmatch(stderr, -1)
	if len(matches) == 0 {
		return ""
	}
	return "native_last_helper_phase=" + matches[len(matches)-1][1] + "; "
}
