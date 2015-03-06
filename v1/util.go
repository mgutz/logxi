package log

import (
	"bytes"
	"strings"
)

func expandTabs(s string, tabLen int) string {
	if s == "" {
		return s
	}
	parts := strings.Split(s, "\t")
	var buf bytes.Buffer
	for _, part := range parts {
		buf.WriteString(part)
		buf.WriteString(strings.Repeat(" ", 4-len(part)%4))
	}
	return buf.String()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
