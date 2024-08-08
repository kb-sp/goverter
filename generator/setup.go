package generator

import (
	"github.com/kb-sp/goverter/config"
	"github.com/kb-sp/goverter/method"
	"github.com/kb-sp/goverter/namer"
	"github.com/kb-sp/goverter/xtype"
)

func setupGenerator(converter *config.Converter, n *namer.Namer) *generator {
	extend := map[xtype.Signature]*method.Definition{}
	for _, def := range converter.Extend {
		extend[def.Signature()] = def
	}

	lookup := map[xtype.Signature]*generatedMethod{}
	for _, method := range converter.Methods {
		lookup[method.Definition.Signature()] = &generatedMethod{
			Method:   method,
			Dirty:    true,
			Explicit: true,
		}
	}

	gen := generator{
		namer:  n,
		conf:   converter,
		lookup: lookup,
		extend: extend,
	}

	return &gen
}
