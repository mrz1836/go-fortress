module github.com/mrz1836/go-fortress

go 1.25.0

require (
	github.com/owenrumney/go-sarif/v3 v3.3.1
	github.com/rhysd/actionlint v1.7.12
	github.com/stretchr/testify v1.12.1
	go.yaml.in/yaml/v4 v4.0.0-rc.3 // pinned: actionlint v1.7.10 incompatible with rc.4 (yaml.ParserError API change)
)

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mattn/go-runewidth v0.0.28 // indirect
	github.com/mattn/go-shellwords v1.0.14 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace go.yaml.in/yaml/v4 => go.yaml.in/yaml/v4 v4.0.0-rc.3
