package builder

import (
	"github.com/dave/jennifer/jen"
	"github.com/jmattheis/goverter/xtype"
)

/*
	This applies special sauce to PB []foo -> *[]foo translations:

		var pModelsVendorContactUpdateList []*models.VendorContactUpdate
		if (*source).Contacts != nil {
			pModelsVendorContactUpdateList = make([]*models.VendorContactUpdate, len((*source).Contacts))
			for j := 0; j < len((*source).Contacts); j++ {
				pModelsVendorContactUpdateList[j] = c.VendorContactUpdateFromPB((*source).Contacts[j])
			}
		}
		// <UPDATE_HANDLING>
		pPModelsContractTermsList := pModelsContractTermsList
		pModelsContractTermsList2 := &pPModelsContractTermsList
		if len((*source).ContractTerms) == 0 {
			pModelsContractTermsList2 = nil
		} else {
			if len(pModelsContractTermsList) == 1 {
				var isZero3 bool
				pModelsv1ContractTerms := reflect.ValueOf(*pModelsContractTermsList[0])
				isZero3 = isZero3 || pModelsv1ContractTerms.IsZero()
				pModelsv1ContractTerms2 := reflect.ValueOf(pModelsContractTermsList[0])
				isZero3 = isZero3 || pModelsv1ContractTerms2.IsZero()
				if isZero3 {
					pPModelsContractTermsList2 := pModelsContractTermsList[:0]
					pModelsContractTermsList2 = &pPModelsContractTermsList2
				}
			}
		}
		// </UPDATE_HANDLING>
		modelsVendorUpdate.Contacts = pModelsVendorContactUpdateList2

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
	// emptyPtrValue := ctx.Name(source.ListInner.ID())

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
	)
	codes = append(codes,
		jen.If(jen.Id(isZero)).Block(fixCode...),
	)

	stmt = append(stmt,
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
