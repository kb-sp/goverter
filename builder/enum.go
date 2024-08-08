package builder

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/kb-sp/goverter/config"
	"github.com/kb-sp/goverter/enum"
	"github.com/kb-sp/goverter/xtype"
)

// Basic handles basic data types.
type Enum struct{}

// Matches returns true, if the builder can create handle the given types.
func (*Enum) Matches(ctx *MethodContext, source, target *xtype.Type) bool {
	return isEnum(ctx, source, target)
}

func isEnum(ctx *MethodContext, source, target *xtype.Type) bool {
	return ctx.Conf.Enum.Enabled &&
		source.Enum(&ctx.Conf.Enum).OK &&
		target.Enum(&ctx.Conf.Enum).OK
}

// Build creates conversion source code for the given source and target type.
func (*Enum) Build(gen Generator, ctx *MethodContext, sourceID *xtype.JenID, source, target *xtype.Type, path ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	stmt, nameVar, err := buildTargetVar(gen, ctx, sourceID, source, target, path)
	if err != nil {
		return nil, nil, err
	}

	var cases []jen.Code

	targetEnum := target.Enum(&ctx.Conf.Enum)
	sourceEnum := source.Enum(&ctx.Conf.Enum)

	definedKeys := ctx.DefinedEnumFields(target)

	transformerMapping, err := executeTransformers(ctx.Conf.EnumMapping.Transformers, source, target, sourceEnum, targetEnum)
	if err != nil {
		return nil, nil, err
	}

	sourceTargetMapping := map[interface{}]enumMapping{}
	for _, sourceName := range sourceEnum.SortedMembers() {
		delete(definedKeys, sourceName)

		targetName, ok := ctx.Conf.EnumMapping.Map[sourceName]
		if !ok {
			targetName, ok = transformerMapping[sourceName]
		}

		if !ok {
			targetName = sourceName
		}

		sourceQual := jen.Qual(source.NamedType.Obj().Pkg().Path(), sourceName)
		body, err := caseAction(gen, ctx, nameVar, target, targetEnum, targetName, sourceID, path)
		if err != nil {
			return nil, nil, err.Lift(&Path{
				SourceType: fmtEnumValue(sourceEnum, sourceName),
				SourceID:   sourceName,
				Prefix:     ".",
				TargetID:   targetName,
				TargetType: "???",
			})
		}

		sourceValue := sourceEnum.Members[sourceName]
		if previous, ok := sourceTargetMapping[sourceValue]; ok {
			if enumTargetMismatches(previous, targetEnum, targetName) {
				return nil, nil, enumTargetMismatchError(targetEnum, sourceName, targetName, previous, sourceValue).Lift(&Path{
					SourceType: fmtEnumValue(sourceEnum, sourceName),
					SourceID:   sourceName,
					Prefix:     ".",
					TargetID:   targetName,
					TargetType: fmtEnumValue(targetEnum, targetName),
				})
			} else {
				cases = append(cases, jen.Comment(fmt.Sprintf("Skipped %s -> %s because it duplicates %s -> %s",
					fmtEnumValue(sourceEnum, sourceName), fmtEnumValue(targetEnum, targetName),
					fmtEnumValue(sourceEnum, previous.Source), fmtEnumValue(targetEnum, previous.Target))))
			}
		} else {
			sourceTargetMapping[sourceValue] = enumMapping{Source: sourceName, Target: targetName}
			cases = append(cases, jen.Case(sourceQual).Add(body))
		}
	}

	enumUnknown := ctx.Conf.Common.Enum.Unknown
	if enumUnknown == "" {
		return nil, nil, NewError("Enum detected but enum:unknown is not configured.\nSee https://goverter.jmattheis.de/guide/enum")
	}

	body, err := caseAction(gen, ctx, nameVar, target, targetEnum, enumUnknown, sourceID, path)
	if err != nil {
		return nil, nil, err.Lift(&Path{
			SourceID:   "@enum:unknown",
			Prefix:     ".",
			TargetID:   enumUnknown,
			TargetType: "???",
		})
	}
	cases = append(cases, jen.Default().Add(body))

	for name := range definedKeys {
		return nil, nil, NewError(fmt.Sprintf("Configured enum value %s does not exist on\n    %s", name, source.String)).
			Lift(&Path{
				Prefix:     ".",
				SourceID:   name,
				SourceType: "???",
			})
	}

	stmt = append(stmt, jen.Switch(sourceID.Code).Block(cases...))
	return stmt, xtype.VariableID(nameVar), nil
}

func (s *Enum) Assign(gen Generator, ctx *MethodContext, assignTo *AssignTo, sourceID *xtype.JenID, source, target *xtype.Type, path ErrorPath) ([]jen.Code, *Error) {
	return AssignByBuild(s, gen, ctx, assignTo, sourceID, source, target, path)
}

func caseAction(gen Generator, ctx *MethodContext, nameVar *jen.Statement, target *xtype.Type, targetEnum *xtype.Enum, targetName string, sourceID *xtype.JenID, errPath ErrorPath) (jen.Code, *Error) {
	if config.IsEnumAction(targetName) {
		switch targetName {
		case config.EnumActionIgnore:
			return jen.Comment("ignored"), nil
		case config.EnumActionPanic:
			return jen.Panic(jen.Qual("fmt", "Sprintf").Call(jen.Lit("unexpected enum element: %v"), sourceID.Code.Clone())), nil
		case config.EnumActionError:
			errStmt := jen.Qual("fmt", "Errorf").Call(jen.Lit("unexpected enum element: %v"), sourceID.Code.Clone())
			code, ok := gen.ReturnError(ctx, errPath, errStmt)
			if !ok {
				return nil, NewError(fmt.Sprintf("Cannot return %s because the explicitly defined conversion method doesn't return an error.", config.EnumActionError))
			}
			return code, nil
		default:
			return nil, NewError(fmt.Sprintf("invalid target %q", targetName))
		}
	}
	_, ok := targetEnum.Members[targetName]
	if !ok {
		return nil, NewError(fmt.Sprintf("Enum %s does not exist on\n    %s\n\nSee https://goverter.jmattheis.de/guide/enum", targetName, target.String))
	}

	targetQual := jen.Qual(target.NamedType.Obj().Pkg().Path(), targetName)
	return nameVar.Clone().Op("=").Add(targetQual), nil
}

func executeTransformers(transformers []config.ConfiguredTransformer, source, target *xtype.Type, sourceEnum, targetEnum *xtype.Enum) (map[string]string, *Error) {
	transformerMapping := map[string]string{}
	for _, t := range transformers {
		m, err := t.Transformer(enum.TransformContext{
			Source: enum.Enum{Type: source.NamedType, Members: sourceEnum.Members},
			Target: enum.Enum{Type: target.NamedType, Members: targetEnum.Members},
			Config: t.Config,
		})
		if err != nil {
			return nil, NewError(fmt.Sprintf("error executing transformer %q with config %q: %s", t.Name, t.Config, err))
		}
		if len(m) == 0 {
			return nil, NewError(fmt.Sprintf("transformer %q with config %q did not return any mapped values. Is there an configuration error?", t.Name, t.Config))
		}
		for key, value := range m {
			transformerMapping[key] = value
		}
	}
	return transformerMapping, nil
}

func enumTargetMismatches(previous enumMapping, targetEnum *xtype.Enum, targetName string) bool {
	if !config.IsEnumAction(targetName) && !config.IsEnumAction(previous.Target) {
		return targetEnum.Members[previous.Target] != targetEnum.Members[targetName]
	}
	return targetName != previous.Target
}

func enumTargetMismatchError(targetEnum *xtype.Enum, sourceName, targetName string, previous enumMapping, sourceValue interface{}) *Error {
	return NewError(fmt.Sprintf(`Detected multiple enum source members with the same value but different target values/actions.
    %s(%v) -> %s
    %s(%v) -> %s

Explicitly define what the correct mapping is. E.g. by adding
    goverter:enum:map %s %s
    goverter:enum:map %s %s

See https://goverter.jmattheis.de/guide/enum#mapping-enum-keys`,
		previous.Source, sourceValue, fmtEnumValue(targetEnum, previous.Target),
		sourceName, sourceValue, fmtEnumValue(targetEnum, targetName),
		previous.Source, previous.Target,
		sourceName, previous.Target))
}

func fmtEnumValue(targetEnum *xtype.Enum, targetName string) string {
	if config.IsEnumAction(targetName) {
		return fmt.Sprintf("%s(action)", targetName)
	}
	return fmt.Sprintf("%s(%v)", targetName, targetEnum.Members[targetName])
}

type enumMapping struct {
	Target string
	Source string
}
