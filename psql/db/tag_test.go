package db

import (
	"context"
	"testing"

	"github.com/bwen19/blog/util"
	"github.com/stretchr/testify/require"
)

func createRandomTag(t *testing.T) Tag {
	length := util.RandomInt(3, 10)
	name := util.RandomString(int(length))

	tag, err := testStore.CreateTag(context.Background(), name)
	require.NoError(t, err)
	require.NotEmpty(t, tag)

	require.Equal(t, name, tag.Name)
	require.NotZero(t, tag.ID)

	return tag
}
