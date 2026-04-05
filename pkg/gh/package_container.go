package gh

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// ContainerRegistry returns the container registry host for the given GitHub host.
// For github.com, it returns "ghcr.io".
// For GitHub Enterprise Server, it returns "containers.HOSTNAME".
func ContainerRegistry(host string) string {
	if host == "" || host == defaultHost {
		return "ghcr.io"
	}
	return "containers." + host
}

// ContainerImageBase returns the base image path for an OCI image using a repository definition.
// Owner and package are lowercased to comply with the OCI Distribution Spec.
func ContainerImageBase(repo repository.Repository, pkg string) string {
	host := repo.Host
	if host == "" {
		host = defaultHost
	}
	return ContainerRegistry(host) + "/" + strings.ToLower(repo.Owner) + "/" + strings.ToLower(pkg)
}

// DockerRegistry returns the legacy Docker Package Registry host for the given GitHub host.
// For github.com, it returns "docker.pkg.github.com".
// For GitHub Enterprise Server, it returns "docker.pkg.HOSTNAME".
func DockerRegistry(host string) string {
	if host == "" || host == defaultHost {
		return "docker.pkg.github.com"
	}
	return "docker.pkg." + host
}

// DockerImageBase returns the base image path for a legacy Docker Package Registry image.
// Owner, repository name, and package are lowercased to comply with the OCI Distribution Spec.
func DockerImageBase(repo repository.Repository, pkg string) string {
	host := repo.Host
	if host == "" {
		host = defaultHost
	}
	return DockerRegistry(host) + "/" + strings.ToLower(repo.Owner) + "/" + strings.ToLower(repo.Name) + "/" + strings.ToLower(pkg)
}
