package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListChildrenWithParams(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolName)
	require.NoError(t, err)
	defer ioctx.Destroy()

	name := "parent"
	img, err := Create(ioctx, name, testImageSize, testImageOrder, 1)
	assert.NoError(t, err)
	defer img.Remove()

	img, err = OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	defer img.Close()

	snapName := "snap01"
	snapshot, err := img.CreateSnapshot(snapName)
	assert.NoError(t, err)

	err = snapshot.Protect()
	assert.NoError(t, err)

	// create an image context with the parent+snapshot
	snapImg, err := OpenImage(ioctx, name, snapName)
	assert.NoError(t, err)
	defer snapImg.Close()

	// ensure no children prior to clone
	result, err := snapImg.ListChildrenWithParams()
	assert.NoError(t, err)
	assert.Equal(t, len(result), 0, "List should be empty before cloning")

	//create first child Image
	childImageName := "childImage"
	clone, err := img.Clone(snapName, ioctx, childImageName, 1, testImageOrder)
	assert.NoError(t, err)
	defer func() {
		childImg, err := OpenImage(ioctx, childImageName, NoSnapshot)
		if err == nil {
			childImg.Remove()
		}
	}()

	result, err = snapImg.ListChildrenWithParams()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(result), 1, "List should contain one child image")

	cloneID, err := clone.GetId()
	assert.NoError(t, err, "Failed to get ImageID for first child")

	//validate child image properties
	assert.Equal(t, childImageName, result[0].ImageName)
	assert.Equal(t, cloneID, result[0].ImageID)
	assert.Equal(t, ioctx.GetPoolID(), result[0].PoolID)
	assert.Equal(t, poolName, result[0].PoolName)
	assert.Equal(t, result[0].PoolNamespace, "")
	assert.False(t, result[0].Trash, "Newly cloned image should not be in trash")

}
