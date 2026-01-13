// src/internal/gec/debugger.go
package gec

import (
	"gec-demo/src/internal/print"
)

func ViewGibbs(gibb_scores []GibbResults) {
	print.Debug("Gibb Scores:")
	for _, gibbScr := range gibb_scores {
		print.Debug("{%v, %v, %v}", gibbScr.Index, gibbScr.Length, gibbScr.Score)
	}
}

func ViewMisspells(misspells []Misspell) {
	print.Debug("Misspells:")
	for _, miss := range misspells {
		print.Debug("{%v, %v, %v, %q},", miss.Index, miss.Length, miss.Type, miss.Suggestions)
	}
}
