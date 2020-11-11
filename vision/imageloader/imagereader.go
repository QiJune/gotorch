package imageloader

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/wangkuiyi/gotorch/tool/recordio"
	"github.com/wangkuiyi/gotorch/tool/tgz"
	"gocv.io/x/gocv"
)

// ImageReader struct
type ImageReader struct {
	colorSpace string
	// tgz
	tgzReader *tgz.Reader
	vocab     map[string]int
	// recordio
	recordioReader *recordio.Reader
}

// NewTgzImageReader creates an ImageReader from a tgz file
func NewTgzImageReader(fn string, vocab map[string]int, colorSpace string) (*ImageReader, error) {
	r, e := tgz.OpenFile(fn)
	if e != nil {
		return nil, e
	}

	return &ImageReader{
		colorSpace: colorSpace,
		tgzReader:  r,
		vocab:      vocab,
	}, nil
}

// NewRecordIOImageReader creates an ImageReader from recordio files
func NewRecordIOImageReader(fn string, recordsPerShard, rank, size int, seed int64, colorSpace string) (*ImageReader, error) {
	tm := recordio.NewTaskManager(fn, recordsPerShard, size)
	tasks := tm.AssginTask(seed, rank)
	r, e := recordio.NewReaderFromTask(tasks)
	if e != nil {
		return nil, e
	}

	return &ImageReader{
		colorSpace:     colorSpace,
		recordioReader: r,
	}, nil
}

// ReadSample returns a pair of <*gocv.Mat, int>
func (ir *ImageReader) ReadSample() (*gocv.Mat, int, error) {
	if ir.tgzReader != nil {
		hdr, err := ir.tgzReader.Next()
		if err != nil {
			return nil, -1, err
		}

		if !hdr.FileInfo().Mode().IsRegular() {
			return nil, -1, nil
		}

		classStr := filepath.Base(filepath.Dir(hdr.Name))
		label := ir.vocab[classStr]
		buffer := make([]byte, hdr.Size)
		io.ReadFull(ir.tgzReader, buffer)

		m, err := decodeImage(buffer, ir.colorSpace)
		if err != nil {
			return nil, -1, err
		}
		if m.Empty() {
			return nil, -1, fmt.Errorf("read invalid image content")
		}
		return &m, label, nil
	} else if ir.recordioReader != nil {
		record, err := ir.recordioReader.Next()
		if err != nil {
			return nil, -1, err
		}
		label := record.Label
		m, err := decodeImage(record.Image, ir.colorSpace)
		if err != nil {
			return nil, -1, err
		}
		if m.Empty() {
			return nil, -1, fmt.Errorf("read invalid image content")
		}
		return &m, label, nil
	} else {
		return nil, -1, fmt.Errorf("reader not initialized")
	}
}

func decodeImage(buffer []byte, colorSpace string) (gocv.Mat, error) {
	var m gocv.Mat
	var e error
	if colorSpace == "rgb" {
		m, e = gocv.IMDecode(buffer, gocv.IMReadColor)
		gocv.CvtColor(m, &m, gocv.ColorBGRToRGB)
	} else if colorSpace == "gray" {
		m, e = gocv.IMDecode(buffer, gocv.IMReadGrayScale)
	} else {
		return m, fmt.Errorf("Cannot read image with color space %v", colorSpace)
	}
	return m, e
}
