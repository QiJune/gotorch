package recordio

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	dir, _ := createTestData(10, 16)
	defer os.RemoveAll(dir)

	tm := NewTaskManager(dir, 16, 2)
	t0 := tm.AssginTask(0, 0)
	t1 := tm.AssginTask(0, 1)

	assert.Equal(t, 10/2, len(t0))
	assert.Equal(t, 10/2, len(t1))
}
