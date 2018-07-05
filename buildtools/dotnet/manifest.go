package dotnet

import (
	"strings"

	"github.com/fossas/fossa-cli/files"
)

type Manifest struct {
	PropertyGroup []PropertyGroup
	ItemGroup     []ItemGroup
}

func (m *Manifest) Name() string {
	for _, propertyGroup := range m.PropertyGroup {
		if propertyGroup.RootNamespace != "" {
			return propertyGroup.RootNamespace
		}
	}
	return ""
}

func (m *Manifest) Version() string {
	for _, propertyGroup := range m.PropertyGroup {
		if propertyGroup.Version != "" {
			return propertyGroup.Version
		}
	}
	return ""
}

func (m *Manifest) Imports() []Reference {
	var refs []Reference
	for _, itemGroup := range m.ItemGroup {
		refs = append(refs, itemGroup.ProjectReference...)
		refs = append(refs, itemGroup.PackageReference...)
	}
	return refs
}

func (m *Manifest) Packages() []Reference {
	var pkgs []Reference
	for _, itemGroup := range m.ItemGroup {
		pkgs = append(pkgs, itemGroup.PackageReference...)
	}
	return pkgs
}

func (m *Manifest) Projects() []Reference {
	var projects []Reference
	for _, itemGroup := range m.ItemGroup {
		projects = append(projects, itemGroup.ProjectReference...)
	}
	return projects
}

type PropertyGroup struct {
	RootNamespace string
	Version       string
}

type ItemGroup struct {
	Reference        []Reference
	PackageReference []Reference
	ProjectReference []Reference
}

type Reference struct {
	Include string `xml:",attr"`
	Version string `xml:",attr"`
}

type NuSpec struct {
	Metadata Metadata
}

type Metadata struct {
	ID      string
	Version string
}

type Lockfile struct {
	Version int
	Targets map[string]map[string]Target

	resolved map[string]resolved
}

type resolved struct {
	Version string
	Imports map[string]string
}

func ReadLockfile(filename string) (Lockfile, error) {
	var lockfile Lockfile
	err := files.ReadJSON(&lockfile, filename)
	if err != nil {
		return Lockfile{}, err
	}

	lockfile.resolved = make(map[string]resolved)
	for _, deps := range lockfile.Targets {
		for key, dep := range deps {
			sections := strings.Split(key, "/")
			name := sections[0]
			version := sections[1]
			lockfile.resolved[name] = resolved{
				Version: version,
				Imports: dep.Dependencies,
			}
		}
	}

	return lockfile, nil
}

func (l *Lockfile) Resolve(pkg string) string {
	return l.resolved[pkg].Version
}

func (l *Lockfile) Imports(pkg string) map[string]string {
	return l.resolved[pkg].Imports
}

type Target struct {
	Type         string
	Dependencies map[string]string
}
