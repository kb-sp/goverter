// Code generated by github.com/kb-sp/goverter, DO NOT EDIT.
//go:build !goverter

package generated

import useunderlyingtypemethods "github.com/kb-sp/goverter/example/use-underlying-type-methods"

type ConverterImpl struct{}

func (c *ConverterImpl) Convert(source useunderlyingtypemethods.Input) useunderlyingtypemethods.Output {
	var exampleOutput useunderlyingtypemethods.Output
	exampleOutput.ID = useunderlyingtypemethods.OutputID(useunderlyingtypemethods.ConvertUnderlying(int(source.ID)))
	return exampleOutput
}
