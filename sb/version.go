package sb

import "runtime/debug"

// modulePath is the module that provides this library, used to find its
// version in the build info embedded in the running binary.
const modulePath = "github.com/unfoldingWord/go-rc2sb"

// moduleVersion is resolved once at package init from the binary's build info.
var moduleVersion = resolveModuleVersion()

// resolveModuleVersion returns the go-rc2sb module version embedded in the
// binary. When go-rc2sb is the main module (e.g. the rc2sb CLI), Go 1.24+
// stamps the version from the VCS tag at build time; when it is a dependency
// of another module, the version comes from the importer's go.mod. Builds
// without version information (go run, go test, untagged source) report
// "(devel)".
func resolveModuleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	if info.Main.Path == modulePath && info.Main.Version != "" {
		return info.Main.Version
	}
	for _, dep := range info.Deps {
		if dep.Path != modulePath {
			continue
		}
		if dep.Replace != nil && dep.Replace.Version != "" {
			return dep.Replace.Version
		}
		if dep.Version != "" {
			return dep.Version
		}
	}
	return "(devel)"
}
