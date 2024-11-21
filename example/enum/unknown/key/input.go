package example

import (
	"github.com/kb-sp/goverter/example/enum/unknown/key/input"
	"github.com/kb-sp/goverter/example/enum/unknown/key/output"
)

// goverter:converter
type Converter interface {
	// goverter:enum:unknown Unknown
	Convert(input.Color) output.Color
}
