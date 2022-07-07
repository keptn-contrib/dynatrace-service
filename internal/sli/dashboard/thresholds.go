package dashboard

import (
	"golang.org/x/exp/slices"
)

var passColors = []string{
	"#006613",
	"#1f7e1e",
	"#5ead35",
	"#7dc540",
	"#9cd575",
	"#e8f9dc",
	"#048855",
	"#009e60",
	"#2ab06f",
	"#54c27d",
	"#99dea8",
	"#e1f7dc",
}

var warnColors = []string{
	"#ef651f",
	"#fd8232",
	"#ffa86c",
	"#ffd0ab",
	"#c9a000",
	"#e6be00",
	"#f5d30f",
	"#ffe11c",
	"#ffee7c",
	"#fff9d5",
}

var failColors = []string{
	"#93060e",
	"#ab0c17",
	"#c41425",
	"#dc172a",
	"#f28289",
	"#ffeaea",
}

func isPassColor(color string) bool {
	return slices.Contains(passColors, color)
}

func isWarnColor(color string) bool {
	return slices.Contains(warnColors, color)
}

func isFailColor(color string) bool {
	return slices.Contains(failColors, color)
}
