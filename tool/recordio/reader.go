package recordio

import (
	"bytes"
	"encoding/gob"
	"io"
	"os"

	"github.com/wangkuiyi/recordio"
)

// Reader struct
type Reader struct {
	Files      []*os.File
	Scanners   []*recordio.Scanner
	CurScanner int
}

// NewReaderFromFile returns a Reader from files
func NewReaderFromFile(fns []string) (*Reader, error) {
	r := &Reader{
		Files:    []*os.File{},
		Scanners: []*recordio.Scanner{},
	}
	for _, fn := range fns {
		f, err := os.Open(fn)
		if err != nil {
			return nil, err
		}
		idx, err := recordio.LoadIndex(f)
		if err != nil {
			return nil, err
		}
		s := recordio.NewScanner(f, idx, 0, idx.NumRecords())
		r.Scanners = append(r.Scanners, s)
		r.Files = append(r.Files, f)
	}
	return r, nil
}

// NewReaderFromTask returns a Reader from tasks
func NewReaderFromTask(tasks []Task) (*Reader, error) {
	r := &Reader{
		Files:    []*os.File{},
		Scanners: []*recordio.Scanner{},
	}
	for _, task := range tasks {
		for _, shard := range task.Shards {
			f, err := os.Open(shard.Name)
			if err != nil {
				return nil, err
			}
			idx, err := recordio.LoadIndex(f)
			if err != nil {
				return nil, err
			}
			s := recordio.NewScanner(f, idx, shard.Start, shard.End-shard.Start)
			r.Scanners = append(r.Scanners, s)
			r.Files = append(r.Files, f)
		}
	}
	return r, nil
}

// Next returns the next image record
func (r *Reader) Next() (*ImageRecord, error) {
	success := r.Scanners[r.CurScanner].Scan()
	if !success {
		r.CurScanner++
		if r.CurScanner == len(r.Scanners) {
			return nil, io.EOF
		}
		return r.Next()
	}
	ir := &ImageRecord{}
	err := gob.NewDecoder(bytes.NewBuffer(r.Scanners[r.CurScanner].Record())).Decode(&ir)
	if err != nil {
		return nil, err
	}
	return ir, nil
}

// Close the reader
func (r *Reader) Close() {
	for _, fs := range r.Files {
		fs.Close()
	}
}
