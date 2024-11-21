package example

import (
	"github.com/kb-sp/goverter/example/enum/map-panic/input"
	"github.com/kb-sp/goverter/example/enum/map-panic/output"
)

// goverter:converter
// goverter:enum:unknown @panic
type Converter interface {
	// goverter:enum:map Gray @panic
	Convert(input.Color) output.Color
}
