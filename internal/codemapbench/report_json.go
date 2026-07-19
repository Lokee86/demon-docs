package codemapbench

import (
	"bytes"
	"encoding/json"
	"io"
)

// ReportSchemaVersion identifies the stable JSON benchmark report shape.
const ReportSchemaVersion = 1

type jsonReportEnvelope struct {
	SchemaVersion int `json:"schema_version"`
	Report
}

// MarshalJSONReport returns a canonical, indented JSON benchmark report.
func MarshalJSONReport(report Report) ([]byte, error) {
	var output bytes.Buffer
	if err := WriteJSONReport(&output, report); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// WriteJSONReport writes a canonical, versioned JSON benchmark report.
func WriteJSONReport(writer io.Writer, report Report) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonReportEnvelope{
		SchemaVersion: ReportSchemaVersion,
		Report:        canonicalReport(report),
	})
}
