package commands

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCache_AddAndGetMin(t *testing.T) {
	cache := NewRawGPCache()

	// Add values to the cache
	cache.Add(big.NewInt(5))
	cache.Add(big.NewInt(3))
	cache.Add(big.NewInt(7))

	// Check the minimum value
	minRGP, err := cache.GetMin()
	require.NoError(t, err)

	expectedMin := big.NewInt(3)
	require.Equal(t, minRGP.Int64(), expectedMin.Int64())
}

func TestCache_EmptyCache(t *testing.T) {
	cache := NewRawGPCache()

	// Try to get the minimum value from an empty cache
	_, err := cache.GetMin()
	require.Error(t, err)
}

func TestCache_OldValuesCleanup(t *testing.T) {
	cache := NewRawGPCache()

	// Add an old value
	oldValue := big.NewInt(10)
	cache.values[time.Now().Add(-6*time.Minute)] = oldValue

	// Add a recent value
	recentValue := big.NewInt(5)
	cache.Add(recentValue)

	// Check the minimum value
	minRGP, err := cache.GetMin()
	require.NoError(t, err)

	expectedMin := big.NewInt(5)
	require.Equal(t, minRGP.Int64(), expectedMin.Int64())

	// Ensure the old value is removed
	require.Equal(t, 1, len(cache.values))
}

func TestCache_MultipleValues(t *testing.T) {
	cache := NewRawGPCache()

	// Add multiple values
	cache.Add(big.NewInt(8))
	cache.Add(big.NewInt(6))
	cache.Add(big.NewInt(4))
	cache.Add(big.NewInt(9))

	// Check the minimum value
	minRGP, err := cache.GetMin()
	require.NoError(t, err)

	expectedMin := big.NewInt(4)
	require.Equal(t, expectedMin.Int64(), minRGP.Int64())

}

func TestCache_ValuesExactlyAtLimit(t *testing.T) {
	cache := NewRawGPCache()

	// Add values exactly at the limit of 5 minutes
	cache.values[time.Now().Add(-5*time.Minute)] = big.NewInt(8)
	cache.values[time.Now().Add(-5*time.Minute+1*time.Second)] = big.NewInt(4)
	cache.values[time.Now().Add(-4*time.Minute)] = big.NewInt(2)

	// Check the minimum value
	minRGP, err := cache.GetMin()
	require.NoError(t, err)

	expectedMin := big.NewInt(2)
	require.Equal(t, expectedMin.Int64(), minRGP.Int64())
}
