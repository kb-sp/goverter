// Code generated by github.com/kb-sp/goverter, DO NOT EDIT.
//go:build !goverter

package generated

import default1 "github.com/kb-sp/goverter/example/default"

type ConverterImpl struct{}

func (c *ConverterImpl) Convert(source *default1.Input) *default1.Output {
	pExampleOutput := default1.NewOutput()
	if source != nil {
		var exampleOutput default1.Output
		exampleOutput.Age = (*source).Age
		if (*source).Name != nil {
			xstring := *(*source).Name
			exampleOutput.Name = &xstring
		}
		pExampleOutput = &exampleOutput
	}
	return pExampleOutput
}
