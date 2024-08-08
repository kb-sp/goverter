//go:generate go run -cover -covermode=atomic github.com/kb-sp/goverter/cmd/goverter gen ./...
//go:generate go run -cover -covermode=atomic github.com/kb-sp/goverter/cmd/goverter gen -cwd ./wrap-errors-using ./
//go:generate go run -cover -covermode=atomic github.com/kb-sp/goverter/cmd/goverter gen -cwd ./protobuf ./
//go:generate go run -C ./enum/transform-custom -cover -covermode=atomic -coverpkg "github.com/kb-sp/goverter/...,goverter/example/..." ./goverter gen ./
package example
