package ascode

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/rpmmd"
	"gopkg.in/yaml.v3"
)

// helper
func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

type asCodeDistro struct {
	basedir    string
	name       string
	imageTypes []asCodeImageType
}

// Returns the name of the distro.
func (d *asCodeDistro) Name() string {
	return d.name
}

// Returns the release version of the distro. This is used in repo
// files on the host system and required for the subscription support.
func (d *asCodeDistro) Releasever() string {
	s := strings.Split(d.name, "-")

	return s[len(s)-1]
}

// Returns the module platform id of the distro. This is used by DNF
// for modularity support.
func (d *asCodeDistro) ModulePlatformID() string {
	// TODO: This is a hack, we should have a way to specify this somewhere
	return "platform:f40"
}

// Returns the ostree reference template
func (d *asCodeDistro) OSTreeRef() string {
	panic("not implemented") // TODO: Implement
}

// Returns a sorted list of the names of the architectures this distro
// supports.
func (d *asCodeDistro) ListArches() []string {
	var arches []string

	for _, it := range d.imageTypes {
		arches = append(arches, it.Architecture)
	}

	return removeDuplicate(arches)
}

// Returns an object representing the given architecture as support
// by this distro.
func (d *asCodeDistro) GetArch(arch string) (distro.Arch, error) {
	return &asCodeArch{
		name: arch,
		d:    d,
	}, nil
}

type asCodeArch struct {
	name string
	d    *asCodeDistro
}

// Returns the name of the architecture.
func (a *asCodeArch) Name() string {
	return a.name
}

// Returns a sorted list of the names of the image types this architecture
// supports.
func (a *asCodeArch) ListImageTypes() []string {
	var imageTypes []string

	for _, it := range a.d.imageTypes {
		if it.Architecture == a.name {
			imageTypes = append(imageTypes, it.NameField)
		}
	}

	return imageTypes
}

// Returns an object representing a given image format for this architecture,
// on this distro.
func (a *asCodeArch) GetImageType(imageType string) (distro.ImageType, error) {
	for _, it := range a.d.imageTypes {
		fmt.Printf("%s, %s", it.NameField, it.Architecture)
		if it.Architecture == a.name && it.NameField == imageType {
			return &it, nil
		}
	}

	return nil, fmt.Errorf("image type not found")
}

// Returns the parent distro
func (a *asCodeArch) Distro() distro.Distro {
	return a.d
}

// TODO: We should separate the data model from the actual interface implementation
type asCodeImageType struct {
	NameField             string   `yaml:"name"`
	Architecture          string   `yaml:"architecture"`
	FilenameField         string   `yaml:"filename"`
	MIMETypeField         string   `yaml:"mime_type"`
	OSTreeRefField        string   `yaml:"ostree_ref"`
	SizeField             uint64   `yaml:"size"`
	PartitionTypeField    string   `yaml:"partition_type"`
	BootModeField         string   `yaml:"boot_mode"`
	BuildPipelinesField   []string `yaml:"build_pipelines"`
	PayloadPipelinesField []string `yaml:"payload_pipelines"`
	ExportsField          []string `yaml:"exports"`
	ManifestPath          string   `yaml:"manifest"`

	distro *asCodeDistro
}

func (it *asCodeImageType) Name() string {
	return it.NameField
}

func (it *asCodeImageType) Arch() distro.Arch {
	panic("not implemented") // TODO: Implement
}

func (it *asCodeImageType) Filename() string {
	return it.FilenameField
}

func (it *asCodeImageType) MIMEType() string {
	return it.MIMETypeField
}

func (it *asCodeImageType) OSTreeRef() string {
	return it.OSTreeRefField
}

func (it *asCodeImageType) Size(size uint64) uint64 {
	panic("not implemented") // TODO: Implement
}

func (it *asCodeImageType) PartitionType() string {
	return it.PartitionTypeField
}

func (it *asCodeImageType) BootMode() distro.BootMode {
	switch it.BootModeField {
	case "hybrid":
		return distro.BOOT_HYBRID
	case "legacy":
		return distro.BOOT_LEGACY
	case "uefi":
		return distro.BOOT_UEFI
	}

	panic("unknown boot mode, handle me nicely! " + it.BootModeField)
}

func (it *asCodeImageType) BuildPipelines() []string {
	return it.BuildPipelinesField
}

func (it *asCodeImageType) PayloadPipelines() []string {
	return it.PayloadPipelinesField
}

func (it *asCodeImageType) PayloadPackageSets() []string {
	panic("not implemented") // TODO: Implement
}

func (it *asCodeImageType) PackageSetsChains() map[string][]string {
	panic("not implemented") // TODO: Implement
}

func (it *asCodeImageType) Exports() []string {
	return it.ExportsField
}

func (it *asCodeImageType) Manifest(bp *blueprint.Blueprint, options distro.ImageOptions, repos []rpmmd.RepoConfig, seed int64) (distro.ManifestInterface, []string, error) {
	return &asCodeManifest{
		bp: bp,
		it: it,
	}, nil, nil
}

type asCodeManifest struct {
	bp *blueprint.Blueprint
	it *asCodeImageType
}

func (m *asCodeManifest) GetPackageSetChains() map[string][]rpmmd.PackageSet {
	return nil
}

func (m *asCodeManifest) GetContainerSourceSpecs() map[string][]container.SourceSpec {
	return nil
}

func (m *asCodeManifest) GetOSTreeSourceSpecs() map[string][]ostree.SourceSpec {
	return nil
}

func (m *asCodeManifest) Serialize(packageSets map[string][]rpmmd.PackageSpec, containerSpecs map[string][]container.Spec, ostreeCommits map[string][]ostree.CommitSpec) (manifest.OSBuildManifest, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("could not create temporary directory: %w", err)
	}
	defer os.RemoveAll(dir) // TODO: handle error

	var customizations struct {
		Version string
		Vars    struct {
			ExtraPackages []string `yaml:"extra_packages"`
			Hostname      string
		} `yaml:"mpp-vars"`
	}

	customizations.Version = "2"
	customizations.Vars.ExtraPackages = m.bp.GetPackages()
	if hostname := m.bp.Customizations.GetHostname(); hostname != nil {
		customizations.Vars.Hostname = *hostname
	}

	custBlob, err := yaml.Marshal(customizations)
	if err != nil {
		return nil, fmt.Errorf("could not marshal customizations: %w", err)
	}

	err = os.WriteFile(path.Join(dir, "customizations.ipp.yaml"), custBlob, 0444)
	if err != nil {
		return nil, fmt.Errorf("could not write customizations: %w", err)
	}

	cmd := exec.Command("osbuild-mpp", "-I", dir, "-I", path.Join(m.it.distro.basedir, m.it.distro.name), path.Join(m.it.distro.basedir, m.it.distro.name, m.it.ManifestPath), "-")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("could not run osbuild-mpp: %w\nstdout: %s", err, stdout.String())
	}

	return stdout.Bytes(), nil
}

func (m *asCodeManifest) GetCheckpoints() []string {
	return nil
}

func (m *asCodeManifest) GetExports() []string {
	return m.it.ExportsField
}

func (d *asCodeDistro) load(itName string) (*asCodeImageType, error) {
	f, err := os.ReadFile(path.Join(d.basedir, d.name, itName))
	if err != nil {
		return nil, fmt.Errorf("could not read imgdef file: %w", err)
	}

	var it asCodeImageType

	err = yaml.Unmarshal(f, &it)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal imgdef file: %w", err)
	}

	it.distro = d

	return &it, nil
}

func newDistro(basedir, name string) (*asCodeDistro, error) {
	matches, err := filepath.Glob(path.Join(basedir, name, "*.imgdef.yaml"))
	if err != nil {
		return nil, err
	}

	// sanity check
	if len(matches) == 0 {
		return nil, fmt.Errorf("no image types found")
	}

	d := &asCodeDistro{
		basedir: basedir,
		name:    name,
	}

	for _, match := range matches {
		typ, err := d.load(path.Base(match))
		if err != nil {
			return nil, err
		}

		d.imageTypes = append(d.imageTypes, *typ)
	}

	return d, nil
}

func DistroFactory(basedir string) func(string) distro.Distro {
	return func(idStr string) distro.Distro {
		files, err := os.ReadDir(basedir)
		if err != nil {
			panic(err) // TODO: handle error properly
		}

		for _, file := range files {
			if file.IsDir() {
				if idStr == file.Name() {
					d, err := newDistro(basedir, idStr)
					if err != nil {
						panic(err) // TODO: handle error properly
					}
					return d
				}
			}
		}

		return nil
	}

}
