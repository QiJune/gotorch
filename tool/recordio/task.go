package recordio

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/wangkuiyi/recordio"
)

// FilePrefix when converting into recordio format
const FilePrefix string = "data-"

// Shard struct
type Shard struct {
	Name  string
	Start int
	End   int
}

// Task struct
type Task struct {
	Shards []Shard
}

// TaskManager struct
type TaskManager struct {
	Dataset         string
	RecordsPerShard int
	Size            int
	mTasks          []Task
	nTasks          []Task
}

// NewTaskManager creates a TaskManager instance
func NewTaskManager(dataset string, recordsPerShard, size int) *TaskManager {
	tm := &TaskManager{
		Dataset:         dataset,
		RecordsPerShard: recordsPerShard,
		Size:            size,
		mTasks:          []Task{},
		nTasks:          []Task{},
	}
	tm.createTask()
	return tm
}

func (tm *TaskManager) createTask() error {
	filenames, err := filepath.Glob(fmt.Sprintf("%s/%s*", tm.Dataset, FilePrefix))
	if err != nil {
		return err
	}

	// 1. check the last recordio file
	fullShardNum := len(filenames)
	f, err := os.Open(filenames[len(filenames)-1])
	if err != nil {
		return err
	}
	idx, err := recordio.LoadIndex(f)
	if err != nil {
		return err
	}
	lastNum := idx.NumRecords()
	if lastNum != tm.RecordsPerShard {
		fullShardNum--
	}

	// 2. divide full shard files
	m := fullShardNum / tm.Size
	n := fullShardNum % tm.Size

	// 3. create task from the first m*size shard files
	for i := 0; i < m*tm.Size; i++ {
		t := Task{Shards: []Shard{
			Shard{Name: filenames[i], Start: 0, End: tm.RecordsPerShard}}}
		tm.mTasks = append(tm.mTasks, t)
	}

	// 4. TODO(qijun) create task from the last n shard files
	recordsNum := n * tm.RecordsPerShard
	if lastNum != tm.RecordsPerShard {
		recordsNum += lastNum
	}

	return nil
}

// AssginTask for a training process giving the rank id
func (tm *TaskManager) AssginTask(seed int64, rank int) []Task {
	rand.Seed(seed)
	rand.Shuffle(len(tm.mTasks), func(i, j int) { tm.mTasks[i], tm.mTasks[j] = tm.mTasks[j], tm.mTasks[i] })

	res := []Task{}
	for i := 0; i < len(tm.mTasks); i++ {
		if i%tm.Size == rank {
			res = append(res, tm.mTasks[i])
		}
	}
	return res
}
