module github.com/mrz1836/go-fortress/magefiles

go 1.25.0

require (
	github.com/magefile/mage v1.16.1
	github.com/mrz1836/go-fortress v0.0.0
	go.yaml.in/yaml/v4 v4.0.0-rc.3 // indirect; pinned: actionlint v1.7.11 requires rc.3 (rc.4 changed yaml error API, removing yaml.ParserError)
)

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.21 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/owenrumney/go-sarif/v3 v3.3.0 // indirect
	github.com/rhysd/actionlint v1.7.11 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)

replace (
	github.com/mrz1836/go-fortress => ../
	go.yaml.in/yaml/v4 => go.yaml.in/yaml/v4 v4.0.0-rc.3
)
