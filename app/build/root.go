package build

import "fmt"

var (
	Name string
	Version string
	Stamp string
	Hash string
	Email string
	Host string
	Type string
)

func ShortVersionString() string {
	if ReleaseBuild() {
		return fmt.Sprintf("%s v%s", Name, Version)
	}
	return fmt.Sprintf("%s v%s (%s)", Name, Version, Type)
}

func LongVersionString() string {
	return fmt.Sprintf("%s v%s (hash: %s, stamp: %s)", Name, Version, Hash, Stamp)
}

func PrettyVersionString() string {
	return fmt.Sprintf("%s v%s\n" +
		"  Commit Hash    : %s\n" +
		"  Build Type     : %s\n" +
		"  Build Date     : %s\n" +
		"  Build Host     : %s\n" +
		"  Builder Email  : %s\n",
		Name, Version, Hash, Type, Stamp, Host, Email)
}

func ReleaseBuild() bool {
	return Type == "release"
}