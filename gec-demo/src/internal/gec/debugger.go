// src/internal/gec/debugger.go
package gec

import (
	"gec-demo/src/internal/print"
)

func ViewMisspells(misspells []Misspell) {
	print.Debug("Misspells:")
	for _, miss := range misspells {
		print.Debug("{%v, %v, %v, %q},", miss.Index, miss.Length, miss.Category, miss.Suggestions)
	}
}
