package builder

import (
	"github.com/dave/jennifer/jen"
	"github.com/kb-sp/goverter/xtype"
)

/*
	This applies special sauce to PB []foo -> *[]foo translations:

	// <UPDATE_HANDLING>
	pPModelsCostTermUpdateList := pModelsCostTermUpdateList
	pModelsCostTermUpdateList2 := &pPModelsCostTermUpdateList
	if len((*source).CostTerms) == 0 {
			pModelsCostTermUpdateList2 = nil
	} else {
			if len(pModelsCostTermUpdateList) == 1 {
					var isZero bool
					pModelsv1CostTermUpdate := reflect.ValueOf(*pModelsCostTermUpdateList[0])
					isZero = isZero || pModelsv1CostTermUpdate.IsZero()
					pModelsv1CostTermUpdate2 := reflect.ValueOf(pModelsCostTermUpdateList[0])
					isZero = isZero || pModelsv1CostTermUpdate2.IsZero()
					if isZero {
							pPModelsCostTermUpdateList2 := pModelsCostTermUpdateList[:0]
							pModelsCostTermUpdateList2 = &pPModelsCostTermUpdateList2
					}
			}
	}
	// </UPDATE_HANDLING>

*/

// TargetListPointer handles array / slice types, where only the target is a pointer.
type TargetListPointer struct{}

// Matches returns true, if the builder can create handle the given types.
func (*TargetListPointer) Matches(_ *MethodContext, source, target *xtype.Type) bool {
	return !source.Pointer &&
		target.Pointer && target.PointerInner.List &&
		(target.PointerInner.ListInner.Basic || target.PointerInner.ListInner.Pointer ||
			target.PointerInner.ListInner.List || target.PointerInner.ListInner.Struct)
}

// Build creates conversion source code for the given source and target type.
func (*TargetListPointer) Build(gen Generator, ctx *MethodContext, sourceID *xtype.JenID, source, target *xtype.Type, path ErrorPath) ([]jen.Code, *xtype.JenID, *Error) {
	ctx.SetErrorTargetVar(jen.Nil())

	stmt, id, err := gen.Build(ctx, sourceID, source, target.PointerInner, path)
	if err != nil {
		return nil, nil, err.Lift(&Path{
			SourceID:   "*",
			SourceType: source.String,
			TargetID:   "*",
			TargetType: target.PointerInner.String,
		})
	}

	sourceValue := ctx.Name(target.ID())
	targetPtr := ctx.Name(target.PointerInner.ID())
	emptyList := ctx.Name(target.ID())

	fixCode := []jen.Code{
		jen.Id(emptyList).Op(":=").Add(id.Code).Index(jen.Empty(), jen.Lit(0)),
		jen.Id(targetPtr).Op("=").Op("&").Id(emptyList),
	}

	codes := make([]jen.Code, 0)
	isZero := ctx.Name("isZero")
	codes = append(codes,
		jen.Var().Id(isZero).Bool(),
	)

	if source.ListInner.Pointer {
		// If the Inner type is a pointer, first add a nil check.
		rValue := ctx.Name(source.ListInner.ID())
		codes = append(codes,
			jen.Id(rValue).Op(":=").Qual("reflect", "ValueOf").Call(jen.Op("*").Add(id.Code.Clone().Index(jen.Lit(0)))),
			jen.Id(isZero).Op("=").Id(isZero).Op("||").Id(rValue).Dot("IsZero").Call(),
		)
	}

	rValue := ctx.Name(source.ListInner.ID())
	codes = append(codes,
		jen.Id(rValue).Op(":=").Qual("reflect", "ValueOf").Call(id.Code.Clone().Index(jen.Lit(0))),
		jen.Id(isZero).Op("=").Id(isZero).Op("||").Id(rValue).Dot("IsZero").Call(),
		jen.If(jen.Id(isZero)).Block(fixCode...),
	)

	stmt = append(stmt,
		// TODO(kb): Remove the Comments.
		jen.Comment("<UPDATE_HANDLING>"),
		jen.Id(sourceValue).Op(":=").Add(id.Code),
		jen.Id(targetPtr).Op(":=").Op("&").Id(sourceValue),
		jen.If(jen.Len(sourceID.Code).Op("==").Lit(0)).Block(
			jen.Id(targetPtr).Op("=").Nil(),
		).Else().Block(
			jen.If(jen.Len(id.Code).Op("==").Lit(1)).Block(codes...),
		),
		jen.Comment("</UPDATE_HANDLING>"),
	)

	return stmt, xtype.VariableID(jen.Id(targetPtr)), nil
}

func (tlp *TargetListPointer) Assign(gen Generator, ctx *MethodContext, assignTo *AssignTo, sourceID *xtype.JenID, source, target *xtype.Type, path ErrorPath) ([]jen.Code, *Error) {
	return AssignByBuild(tlp, gen, ctx, assignTo, sourceID, source, target, path)
}
