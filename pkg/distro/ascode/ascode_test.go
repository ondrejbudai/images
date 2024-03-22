package ascode

import (
	"os"
	"testing"

	"github.com/osbuild/images/internal/common"
	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/distro"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_asCodeImageType_Manifest(t *testing.T) {
	d, err := newDistro("../../../defs", "fedorang-39")
	require.NoError(t, err)

	a, err := d.GetArch("x86_64")
	require.NoError(t, err)

	it, err := a.GetImageType("disk")
	require.NoError(t, err)

	bp := blueprint.Blueprint{
		Packages: []blueprint.Package{{Name: "0ad"}},
		Customizations: &blueprint.Customizations{
			Hostname: common.ToPtr("smurf.machine"),
		},
	}

	m, _, err := it.Manifest(&bp, distro.ImageOptions{}, nil, 0)
	require.NoError(t, err)

	man, err := m.Serialize(nil, nil, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, man)

	// yolo
	err = os.WriteFile("output.json", man, 0644)
	assert.NoError(t, err)
}
