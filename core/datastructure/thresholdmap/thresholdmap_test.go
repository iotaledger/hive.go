package thresholdmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	thresholdMap := New(LowerThresholdMode)

	thresholdMap.Set(1, 2)
	assert.Equal(t, thresholdMap.Size(), 1)
	thresholdMap.Set(1, 3)
	assert.Equal(t, thresholdMap.Size(), 1)
	thresholdMap.Set(3, 2)
	assert.Equal(t, thresholdMap.Size(), 2)
	thresholdMap.Delete(3)
	assert.Equal(t, 1, thresholdMap.Size())
	thresholdMap.Set(7, 8)
	assert.Equal(t, thresholdMap.Size(), 2)
	thresholdMap.Set(9, 10)
	assert.Equal(t, thresholdMap.Size(), 3)

	thresholdMap.ForEach(func(node *Element) bool {
		switch node.Key() {
		case 1:
			assert.Equal(t, 3, node.Value())
		case 7:
			assert.Equal(t, 8, node.Value())
		case 9:
			assert.Equal(t, 10, node.Value())
		default:
			t.Fail()
		}

		return true
	})
}
