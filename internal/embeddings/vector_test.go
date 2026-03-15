package embeddings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeDeserialize_RoundTrip(t *testing.T) {
	original := []float32{1.0, -2.5, 3.14159, 0.0, -0.001}
	serialized := SerializeVector(original)
	require.NotNil(t, serialized)
	assert.Len(t, serialized, len(original)*4)

	deserialized := DeserializeVector(serialized)
	require.Len(t, deserialized, len(original))

	for i := range original {
		assert.Equal(t, original[i], deserialized[i], "mismatch at index %d", i)
	}
}

func TestSerializeDeserialize_384Dim(t *testing.T) {
	original := make([]float32, 384)
	for i := range original {
		original[i] = float32(i) * 0.01
	}

	serialized := SerializeVector(original)
	assert.Len(t, serialized, 384*4) // 1536 bytes

	deserialized := DeserializeVector(serialized)
	require.Len(t, deserialized, 384)

	for i := range original {
		assert.Equal(t, original[i], deserialized[i])
	}
}

func TestSerializeVector_Empty(t *testing.T) {
	assert.Nil(t, SerializeVector(nil))
	assert.Nil(t, SerializeVector([]float32{}))
}

func TestDeserializeVector_Empty(t *testing.T) {
	assert.Nil(t, DeserializeVector(nil))
	assert.Nil(t, DeserializeVector([]byte{}))
}

func TestDeserializeVector_InvalidLength(t *testing.T) {
	// 5 bytes is not divisible by 4
	assert.Nil(t, DeserializeVector([]byte{1, 2, 3, 4, 5}))
}

func TestSerializeDeserialize_SpecialValues(t *testing.T) {
	original := []float32{0.0, -0.0, 1e-38, 1e38}
	deserialized := DeserializeVector(SerializeVector(original))
	require.Len(t, deserialized, len(original))
	for i := range original {
		assert.Equal(t, original[i], deserialized[i])
	}
}
