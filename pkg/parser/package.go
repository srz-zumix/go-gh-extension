package parser

import (
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// PackageRef holds the parsed result of a package reference specification.
type PackageRef struct {
	Repository repository.Repository
	Package    string
}

// ParsePackageRef splits a package reference string into host, owner, and package name.
// The first segment is treated as a host if it contains '.'; otherwise it is the owner.
// Supported formats:
//   - "owner"              -> PackageRef{"", owner, defaultPackage}
//   - "owner/pkg"          -> PackageRef{"", owner, pkg}
//   - "owner/scope/pkg"    -> PackageRef{"", owner, scope/pkg}
//   - "host/owner"         -> PackageRef{host, owner, defaultPackage}
//   - "host/owner/pkg"     -> PackageRef{host, owner, pkg}
//   - "host/owner/scope/pkg" -> PackageRef{host, owner, scope/pkg}
func ParsePackageRef(s, defaultPackage string) (PackageRef, error) {
	if s == "" {
		return PackageRef{}, fmt.Errorf("package reference is required")
	}

	firstSlash := strings.Index(s, "/")
	if firstSlash == -1 {
		// "owner" only
		return PackageRef{Repository: repository.Repository{Owner: s}, Package: defaultPackage}, nil
	}

	first := s[:firstSlash]
	rest := s[firstSlash+1:]

	if first == "" {
		return PackageRef{}, fmt.Errorf("invalid package reference %q: owner cannot be empty", s)
	}
	if strings.Contains(first, ".") {
		// first segment is a host
		if rest == "" {
			return PackageRef{}, fmt.Errorf("invalid package reference %q: owner cannot be empty", s)
		}
		nextSlash := strings.Index(rest, "/")
		if nextSlash == -1 {
			// "host/owner"
			return PackageRef{Repository: repository.Repository{Host: first, Owner: rest}, Package: defaultPackage}, nil
		}
		owner := rest[:nextSlash]
		pkg := rest[nextSlash+1:]
		if owner == "" {
			return PackageRef{}, fmt.Errorf("invalid package reference %q: owner cannot be empty", s)
		}
		if pkg == "" {
			return PackageRef{}, fmt.Errorf("invalid package reference %q: package name cannot be empty", s)
		}
		return PackageRef{Repository: repository.Repository{Host: first, Owner: owner}, Package: pkg}, nil
	}

	// first segment is an owner; rest (including any slashes) is the package name
	pkg := rest
	if pkg == "" {
		pkg = defaultPackage
	}
	return PackageRef{Repository: repository.Repository{Owner: first}, Package: pkg}, nil
}
