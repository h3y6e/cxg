package cmd

import "runtime/debug"

func Version(injected string) string {
	moduleVersion := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		moduleVersion = info.Main.Version
	}

	return resolveVersion(injected, moduleVersion)
}

func resolveVersion(injected string, moduleVersion string) string {
	if injected != "" && injected != "dev" {
		return injected
	}
	if moduleVersion != "" && moduleVersion != "(devel)" {
		return moduleVersion
	}

	return "dev"
}
