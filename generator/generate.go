package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/kb-sp/goverter/builder"
	"github.com/kb-sp/goverter/config"
	"github.com/kb-sp/goverter/namer"
)

// Config the generate config.
type Config struct {
	WorkingDir      string
	BuildConstraint string
}

// BuildSteps that'll used for generation.
var BuildSteps = []builder.Builder{
	&builder.UseUnderlyingTypeMethods{},
	&builder.SkipCopy{},
	&builder.Enum{},
	&builder.BasicTargetPointerRule{},
	&builder.Pointer{},
	&builder.SourcePointer{},
	&builder.TargetListPointer{},
	&builder.TargetPointer{},
	&builder.Basic{},
	&builder.Struct{},
	&builder.List{},
	&builder.Map{},
}

// Generate generates a jen.File containing converters.
func Generate(converters []*config.Converter, c Config) (map[string][]byte, error) {
	manager := &fileManager{Files: map[string]*managedFile{}}

	for _, converter := range converters {
		jenFile, n, err := manager.Get(converter, c)
		if err != nil {
			return nil, err
		}

		if err := generateConverter(converter, c, jenFile, n); err != nil {
			return nil, err
		}
	}

	return manager.renderFiles()
}

func generateConverter(converter *config.Converter, c Config, f *jen.File, n *namer.Namer) error {
	gen := setupGenerator(converter, n)

	if err := validateMethods(gen.lookup); err != nil {
		return err
	}

	if err := gen.buildMethods(f); err != nil {
		return err
	}
	return nil
}
