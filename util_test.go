package minsexp

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTraverseLists(t *testing.T) {
	sexpStr := "(+ (a (b) c ()))"
	sexp, e := ReadFully(sexpStr)
	require.Nil(t, e)
	require.NotNil(t, sexp)

	expectedListCounts := []int{2, 4, 1, 0}
	var actualListCounts []int
	fn := func(list []interface{}) error {
		actualListCounts = append(actualListCounts, len(list))
		return nil
	}
	e = TraverseLists(sexp, fn)
	require.Nil(t, e)
	require.Equal(t, expectedListCounts, actualListCounts)
}
