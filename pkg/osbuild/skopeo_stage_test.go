package osbuild

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSkopeoStage(t *testing.T) {
	// use the same input for all tests since inputs are orthogonal to the destination type
	skopeoInput := ContainersInput{
		inputCommon: inputCommon{
			Type:   "org.osbuild.containers",
			Origin: InputOriginSource,
		},
		References: map[string]ContainersInputSourceRef{
			"sha256:f0c6094df5b84d59e039fe661914a4760c21933a167c4ebd5a0d43fcc83f9b3a": {
				Name: "registry.example.com/linux:latest",
			},
		},
	}
	input := SkopeoStageInputs{
		Images: skopeoInput,
	}

	cases := []struct {
		name                 string
		fn                   func() *Stage
		expectedStageOptions *SkopeoStageOptions
	}{
		{
			name: "containers-storage",
			fn: func() *Stage {
				return NewSkopeoStageWithContainersStorage("/var/lib/containers", skopeoInput, nil)
			},
			expectedStageOptions: &SkopeoStageOptions{
				Destination: &SkopeoDestinationContainersStorage{
					Type:        "containers-storage",
					StoragePath: "/var/lib/containers",
				},
			},
		},
		{
			name: "oci",
			fn: func() *Stage {
				return NewSkopeoStageWithOCI("/container", skopeoInput, nil)
			},
			expectedStageOptions: &SkopeoStageOptions{
				Destination: &SkopeoDestinationOCI{
					Type: "oci",
					Path: "/container",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stage := c.fn()
			assert.Equal(t, "org.osbuild.skopeo", stage.Type)
			assert.Equal(t, input, stage.Inputs)
			assert.Equal(t, c.expectedStageOptions, stage.Options)
		})
	}
}
