package calls

import (
	"math"
	"net/http"
	"sync"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

// QualityReport is one device's summary of a call that just ended: numbers only, nothing about who or what.
type QualityReport struct {
	Mode        string  `json:"mode"`
	Samples     int     `json:"samples"`
	RTTMs       float64 `json:"rtt_ms"`
	LossPct     float64 `json:"loss_pct"`
	JitterMs    float64 `json:"jitter_ms"`
	WeakSamples int     `json:"weak_samples"`
	Reconnects  int     `json:"reconnects"`
	JoinMs      float64 `json:"join_ms"`
}

// valid refuses anything a real browser could not have measured, so a bad client cannot wreck the averages.
func (q QualityReport) valid() bool {
	if q.Mode != "mesh" && q.Mode != "sfu" {
		return false
	}
	in := func(v, top float64) bool { return !math.IsNaN(v) && v >= 0 && v <= top }
	return q.Samples >= 0 && q.Samples <= 100000 && q.WeakSamples >= 0 && q.WeakSamples <= q.Samples &&
		q.Reconnects >= 0 && q.Reconnects <= 1000 && in(q.RTTMs, 60000) && in(q.LossPct, 100) && in(q.JitterMs, 60000) && in(q.JoinMs, 600000)
}

type qualityTotals struct{ reports, samples, weak, reconnects, rtt, loss, jitter, join, joins float64 }

var (
	qualityMu sync.Mutex
	quality   = map[string]*qualityTotals{"mesh": {}, "sfu": {}}
)

func recordQuality(q QualityReport) {
	qualityMu.Lock()
	defer qualityMu.Unlock()
	t := quality[q.Mode]
	t.reports++
	t.samples += float64(q.Samples)
	t.weak += float64(q.WeakSamples)
	t.reconnects += float64(q.Reconnects)
	t.rtt += q.RTTMs
	t.loss += q.LossPct
	t.jitter += q.JitterMs
	if q.JoinMs > 0 {
		t.join += q.JoinMs
		t.joins++
	}
}

// QualityGauges are the running totals for /metrics. Divide a sum by reports for an average per call, e.g.
// rate(hub_call_quality_sfu_loss_pct_sum[1h]) / rate(hub_call_quality_sfu_reports_total[1h]).
func QualityGauges() map[string]func() float64 {
	out := map[string]func() float64{}
	for mode := range quality {
		mode := mode
		add := func(name string, read func(*qualityTotals) float64) {
			out["hub_call_quality_"+mode+"_"+name] = func() float64 {
				qualityMu.Lock()
				defer qualityMu.Unlock()
				return read(quality[mode])
			}
		}
		add("reports_total", func(t *qualityTotals) float64 { return t.reports })
		add("samples_total", func(t *qualityTotals) float64 { return t.samples })
		add("weak_samples_total", func(t *qualityTotals) float64 { return t.weak })
		add("reconnects_total", func(t *qualityTotals) float64 { return t.reconnects })
		add("rtt_ms_sum", func(t *qualityTotals) float64 { return t.rtt })
		add("loss_pct_sum", func(t *qualityTotals) float64 { return t.loss })
		add("jitter_ms_sum", func(t *qualityTotals) float64 { return t.jitter })
		add("join_ms_sum", func(t *qualityTotals) float64 { return t.join })
		add("joins_total", func(t *qualityTotals) float64 { return t.joins })
	}
	return out
}

func handleQuality(w http.ResponseWriter, r *http.Request, _ *tenant.Principal) error {
	var in QualityReport
	if err := httpx.Decode(r, &in, 1<<10); err != nil {
		return err
	}
	if !in.valid() {
		return httpx.Bad("Those call numbers do not look right.")
	}
	recordQuality(in)
	w.WriteHeader(204)
	return nil
}
