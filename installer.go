package installer

import (
	_ "embed"
)

//go:embed go_version.conf
var GoVersion string

type Mode int

const (
	Script Mode = iota // download + exec external script
	GoInstall          // go install source@version
	Binary             // download GitHub release
)

type Tool struct {
	Mode      Mode
	Name      string
	Source    string // module path (GoInstall) or GitHub repo (Binary)
	Version   string // pinned in tool definition
	Required  bool   // true = always install; false = prompt user
	DependsOn string // "" = no dependency; "git" = requires git installed first
}

var Tools = []Tool{
	{Mode: GoInstall, Name: "tinygoinstall", Source: "github.com/tinywasm/tinygo/cmd/tinygoinstall", Version: "0.40.1", Required: true},
	{Mode: Binary, Name: "tinywasm-cli", Source: "https://github.com/tinywasm/tinywasm", Version: "0.1.0", Required: false},
	{Mode: Binary, Name: "tinywasm-server", Source: "https://github.com/tinywasm/tinywasm", Version: "0.1.0", Required: false, DependsOn: "tinywasm-cli"},
}

type Installer struct{}

func New() *Installer {
	return &Installer{}
}
