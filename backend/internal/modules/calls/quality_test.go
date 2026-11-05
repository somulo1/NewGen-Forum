package calls

import (
	"math"
	"testing"
)

func TestQualityReportValid(t *testing.T) {
	good := QualityReport{Mode: "sfu", Samples: 10, WeakSamples: 2, RTTMs: 80, LossPct: 1.5, JitterMs: 12, JoinMs: 1800}
	if !good.valid() {
		t.Fatal("a normal report was refused")
	}
	for name, change := range map[string]func(*QualityReport){
		"unknown mode":           func(q *QualityReport) { q.Mode = "other" },
		"loss over 100":          func(q *QualityReport) { q.LossPct = 101 },
		"negative rtt":           func(q *QualityReport) { q.RTTMs = -1 },
		"not a number":           func(q *QualityReport) { q.JitterMs = math.NaN() },
		"more weak than samples": func(q *QualityReport) { q.WeakSamples = 11 },
	} {
		bad := good
		change(&bad)
		if bad.valid() {
			t.Errorf("%s was accepted", name)
		}
	}
}

func TestRecordQualityAddsUp(t *testing.T) {
	g := QualityGauges()
	before, loss := g["hub_call_quality_mesh_reports_total"](), g["hub_call_quality_mesh_loss_pct_sum"]()
	recordQuality(QualityReport{Mode: "mesh", Samples: 3, LossPct: 2.5})
	if g["hub_call_quality_mesh_reports_total"]() != before+1 || g["hub_call_quality_mesh_loss_pct_sum"]() != loss+2.5 {
		t.Fatal("the totals did not move")
	}
}
