module github.com/Bios-Marcel/yagcl-json/test

go 1.18

require (
	github.com/Bios-Marcel/yagcl v0.0.0-20220629111241-e7cd1cc91730
	github.com/Bios-Marcel/yagcl-json v0.0.0 // No version required
	github.com/stretchr/testify v1.8.0
)

replace github.com/Bios-Marcel/yagcl-json => ../

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
