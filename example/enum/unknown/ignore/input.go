package example

import (
	"github.com/kb-sp/goverter/example/enum/unknown/input"
	"github.com/kb-sp/goverter/example/enum/unknown/output"
)

// goverter:converter
// goverter:enum:unknown @ignore
type Converter interface {
	Convert(input.Color) output.Color
}
