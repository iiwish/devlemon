package reporter

import (
	"encoding/json"
	"io"

	"devlemon/internal/model"
)

type JSONReporter struct {
	w io.Writer
}

func NewJSONReporter(w io.Writer) *JSONReporter {
	return &JSONReporter{w: w}
}

// Render 将 ScanReport 序列化为标准 JSON，专供后续原生 SwiftUI 前端调用消费
func (r *JSONReporter) Render(report *model.ScanReport) error {
	encoder := json.NewEncoder(r.w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
