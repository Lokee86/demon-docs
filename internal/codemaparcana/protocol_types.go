package codemaparcana

import "context"

type queryClient interface {
	query(context.Context, map[string]any, any) error
	Close() error
}

type nodeList struct {
	Count     int            `json:"count"`
	Returned  int            `json:"returned"`
	Truncated bool           `json:"truncated"`
	Nodes     []protocolNode `json:"nodes"`
}

type protocolNode struct {
	NodeID        uint32        `json:"node_id"`
	Key           string        `json:"key"`
	Identity      string        `json:"identity"`
	ContentID     string        `json:"content_id"`
	Kind          string        `json:"kind"`
	Path          string        `json:"path"`
	Name          string        `json:"name"`
	QualifiedName string        `json:"qualified_name"`
	Span          *protocolSpan `json:"span"`
}

type protocolSpan struct {
	Path        string `json:"path"`
	StartLine   uint32 `json:"start_line"`
	StartColumn uint32 `json:"start_column"`
	EndLine     uint32 `json:"end_line"`
	EndColumn   uint32 `json:"end_column"`
}
