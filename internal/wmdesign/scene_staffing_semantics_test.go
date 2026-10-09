package wmdesign

import (
	"encoding/json"
	"math"
	"testing"
)

func TestStaffingSemanticIndependentLineAboveFilledBand(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, smooth := range []string{"0", "0.07"} {
		raw := json.RawMessage(`{"type":"teamcurve","x":0,"y":0,"w":500,"h":200,"max":10,"lineBasis":"independent","curve":"monotone","smooth":` + smooth + `,"series":[{"name":"Team","values":[6,6],"labelAt":false},{"name":"Target","values":[8,8],"style":"line","labelAt":false},{"name":"Partner","values":[2,2],"labelAt":false}]}`)
		p, ok, e := r.planIntakeCurveScene("curve", raw, SceneContext{Surface: "light"})
		if e != nil || !ok {
			t.Fatalf("%s %v", smooth, e)
		}
		line := p.Items[1].Shape
		if line == nil || len(line.Props.Points) < 1 {
			t.Fatal("missing line")
		}
		y := line.Props.Points[0].Y.Val * 72
		if math.Abs(y-29.2) > 1e-6 {
			t.Fatalf("independent line included stack base: %.6f", y)
		}
		if p.Items[2].Shape == nil {
			t.Fatal("line dropped later band")
		}
	}
}
