package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

type Info struct {
	Version string
	Commit  string
	Date    string
}

func Resolve(v string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if v != "" && v != "dev" {
		return strings.TrimPrefix(v, "v")
	}
	if readBuildInfo == nil {
		readBuildInfo = debug.ReadBuildInfo
	}
	info, ok := readBuildInfo()
	if !ok || info == nil {
		return "dev"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}

func (i Info) Format(readBuildInfo func() (*debug.BuildInfo, bool)) string {
	c := i.Commit
	if c == "" {
		c = "unknown"
	}
	d := i.Date
	if d == "" {
		d = "unknown"
	}
	return fmt.Sprintf("%s (commit %s, built at %s)", Resolve(i.Version, readBuildInfo), c, d)
}
