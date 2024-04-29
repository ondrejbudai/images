package osbuild

type SkopeoDestination interface {
	isSkopeoDestination()
}

type SkopeoDestinationContainersStorage struct {
	Type          string `json:"type"`
	StoragePath   string `json:"storage-path,omitempty"`
	StorageDriver string `json:"storage-driver,omitempty"`
}

func (SkopeoDestinationContainersStorage) isSkopeoDestination() {}

type SkopeoDestinationOCI struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

func (SkopeoDestinationOCI) isSkopeoDestination() {}

type SkopeoDestinationOCIArchive struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

func (SkopeoDestinationOCIArchive) isSkopeoDestination() {}

type SkopeoDestinationDir struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

func (SkopeoDestinationDir) isSkopeoDestination() {}

type SkopeoStageOptions struct {
	Destination SkopeoDestination `json:"destination"`
}

func (o SkopeoStageOptions) isStageOptions() {}

type SkopeoStageInputs struct {
	Images        ContainersInput `json:"images"`
	ManifestLists *FilesInput     `json:"manifest-lists,omitempty"`
}

func (SkopeoStageInputs) isStageInputs() {}

func newSkopeoStage(images ContainersInput, manifests *FilesInput, destination SkopeoDestination) *Stage {

	inputs := SkopeoStageInputs{
		Images:        images,
		ManifestLists: manifests,
	}

	return &Stage{
		Type: "org.osbuild.skopeo",
		Options: &SkopeoStageOptions{
			Destination: destination,
		},
		Inputs: inputs,
	}
}

func NewSkopeoStageWithContainersStorage(path string, images ContainersInput, manifests *FilesInput) *Stage {
	return newSkopeoStage(images, manifests, &SkopeoDestinationContainersStorage{
		Type:        "containers-storage",
		StoragePath: path,
	})
}

func NewSkopeoStageWithOCI(path string, images ContainersInput, manifests *FilesInput) *Stage {
	return newSkopeoStage(images, manifests, &SkopeoDestinationOCI{
		Type: "oci",
		Path: path,
	})
}

func NewSkopeoStageWithOCIArchive(path string, images ContainersInput, manifests *FilesInput) *Stage {
	return newSkopeoStage(images, manifests, &SkopeoDestinationOCIArchive{
		Type: "oci-archive",
		Path: path,
	})
}

func NewSkopeoStageWithDir(path string, images ContainersInput, manifests *FilesInput) *Stage {
	return newSkopeoStage(images, manifests, &SkopeoDestinationDir{
		Type: "dir",
		Path: path,
	})
}
