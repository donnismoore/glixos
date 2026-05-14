// Package ui contains terminal output helpers.
package ui

import (
	"fmt"
	"io"
	"strings"
)

// Table prints a simple aligned text table to w.
//
//	headers: column titles
//	rows:    rows of cells; each row should have len(headers) cells
func Table(w io.Writer, headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, r := range rows {
		for i, c := range r {
			if i < len(widths) && len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}

	writeRow := func(cells []string) {
		var parts []string
		for i, c := range cells {
			if i >= len(widths) {
				continue
			}
			if i == len(widths)-1 {
				parts = append(parts, c)
			} else {
				parts = append(parts, c+strings.Repeat(" ", widths[i]-len(c)))
			}
		}
		fmt.Fprintln(w, strings.Join(parts, "  "))
	}

	writeRow(headers)
	sep := make([]string, len(headers))
	for i, h := range headers {
		_ = h
		sep[i] = strings.Repeat("-", widths[i])
	}
	writeRow(sep)
	for _, r := range rows {
		writeRow(r)
	}
}
