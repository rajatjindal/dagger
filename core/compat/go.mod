module github.com/dagger/dagger/core/compat

go 1.23.0

replace github.com/dagger/dagger => ../../

require (
	github.com/dagger/dagger v0.0.0-00010101000000-000000000000
	golang.org/x/mod v0.20.0
)

require (
	github.com/ettle/strcase v0.2.1-0.20230114185658-e5db6a6becf3 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
)
