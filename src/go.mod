// this module contains two modules from the Go standard library:
// * `internal/fmtsort` in directory `fmtsort`. It is a dependency of `text/template`, and modules with `internal` in their path cannot be published.
// * `text/template`
module github.com/sonarsource/go/src

go 1.23.4

require (
	github.com/stretchr/testify v1.8.4
  golang.org/x/crypto v0.31.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	golang.org/x/crypto v0.31.0
)
