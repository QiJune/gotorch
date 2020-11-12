package recordio

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wangkuiyi/recordio"
)

func createTestData(fileNum, recordNum int) (string, map[string]int) {
	testData := &ImageRecord{
		Image: []byte{1, 2},
		Label: 0,
	}
	b, _ := testData.Encode()
	dir, _ := ioutil.TempDir("", "testdata")

	records := map[string]int{}
	for i := 0; i < fileNum; i++ {
		f1, _ := ioutil.TempFile(dir, FilePrefix+strconv.Itoa(i))
		w1 := recordio.NewWriter(f1, -1, -1)
		for i := 0; i < recordNum; i++ {
			w1.Write(b)
		}
		w1.Close()
		records[f1.Name()] = recordNum
	}
	return dir, records
}

func TestReaderFromFile(t *testing.T) {
	dir, records := createTestData(10, 16)
	defer os.RemoveAll(dir)

	files := []string{}
	totalRecords := 0
	for file := range records {
		files = append(files, file)
		totalRecords += records[file]
	}

	r, e := NewReaderFromFile(files)
	defer r.Close()
	assert.Nil(t, e)

	i := 0
	for {
		ir, _ := r.Next()
		if ir == nil {
			break
		}
		i++
	}
	assert.Equal(t, totalRecords, i)
}

func TestReaderFromTask(t *testing.T) {
	dir, _ := createTestData(10, 16)
	defer os.RemoveAll(dir)

	tm := NewTaskManager(dir, 16, 2)
	t0 := tm.AssginTask(0, 0)
	r, e := NewReaderFromTask(t0)
	defer r.Close()
	assert.Nil(t, e)

	totalRecords := 0
	for _, task := range t0 {
		for _, shard := range task.Shards {
			totalRecords += (shard.End - shard.Start)
		}
	}

	i := 0
	for {
		ir, _ := r.Next()
		if ir == nil {
			break
		}
		i++
	}
	assert.Equal(t, totalRecords, i)
}
