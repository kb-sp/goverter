// Code generated by github.com/kb-sp/goverter, DO NOT EDIT.
//go:build !goverter

package generated

import (
	"fmt"
	input "github.com/kb-sp/goverter/example/enum/map/input"
	output "github.com/kb-sp/goverter/example/enum/map/output"
)

type ConverterImpl struct{}

func (c *ConverterImpl) Convert(source input.Color) output.Color {
	var outputColor output.Color
	switch source {
	case input.Gray:
		outputColor = output.Grey
	case input.Green:
		outputColor = output.Green
	default:
		panic(fmt.Sprintf("unexpected enum element: %v", source))
	}
	return outputColor
}
