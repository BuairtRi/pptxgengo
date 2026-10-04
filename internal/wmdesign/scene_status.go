package wmdesign

type sceneStatusSpec struct{ Label, Ink string }

var sceneStatuses = map[string]sceneStatusSpec{
	"on": {"On track", "kpi.on"}, "risk": {"At risk", "kpi.risk"}, "off": {"Off track", "kpi.off"},
	"done": {"Done", "#070154"}, "progress": {"In progress", "#0047FF"},
	"notstarted": {"Not started", "#FFFFFF"}, "blocked": {"Blocked", "#F52C00"},
	"pass": {"Pass", "#1DD566"}, "fail": {"Fail", "#F52C00"},
}
