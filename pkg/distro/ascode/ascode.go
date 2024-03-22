package ascode

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/rpmmd"
	"gopkg.in/yaml.v3"
)

type asCodeImageType struct {
	NameField             string   `yaml:"name"`
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
		mppManifest: it.ManifestPath,
		exports:     it.ExportsField,
		bp:          bp,
	}, nil, nil
}

type asCodeManifest struct {
	mppManifest string
	exports     []string
	bp          *blueprint.Blueprint
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

	cmd := exec.Command("osbuild-mpp", "-I", dir, m.mppManifest, "-")
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
	return m.exports
}

func Load(path string) (*asCodeImageType, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read imgdef file: %w", err)
	}

	var it asCodeImageType

	err = yaml.Unmarshal(f, &it)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal imgdef file: %w", err)
	}

	return &it, nil
}
