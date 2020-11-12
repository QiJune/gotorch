package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	torch "github.com/wangkuiyi/gotorch"
	F "github.com/wangkuiyi/gotorch/nn/functional"
	"github.com/wangkuiyi/gotorch/vision/imageloader"
	"github.com/wangkuiyi/gotorch/vision/models"
	"github.com/wangkuiyi/gotorch/vision/transforms"
)

var masterAddr = flag.String("masterAddr", "127.0.0.1", "The address of master node(rank 0)")
var masterPort = flag.Int("masterPort", 11111, "The port of master node")
var rank = flag.Int("rank", 0, "The rank of the current process")
var size = flag.Int("size", 1, "The size of the processes")
var dataset = flag.String("dataset", "", "The training dataset")

func getGrads(params []torch.Tensor) (grads []torch.Tensor) {
	for _, p := range params {
		grads = append(grads, p.Grad())
	}
	return
}

func mnistLoader(fn string, recordsPerShard, rank, size int, seed int64) *imageloader.ImageLoader {
	trans := transforms.Compose(transforms.ToTensor(), transforms.Normalize([]float32{0.1307}, []float32{0.3081}))
	ir, e := imageloader.NewRecordIOImageReader(fn, recordsPerShard, rank, size, seed, "gray")
	if e != nil {
		panic(e)
	}
	loader, e := imageloader.New(ir, trans, 64, 64, time.Now().UnixNano(), torch.IsCUDAAvailable())
	if e != nil {
		panic(e)
	}
	return loader
}

func main() {
	flag.Parse()

	f, err := os.OpenFile(fmt.Sprintf("%d.log", *rank), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	ts := torch.NewTCPStore(*masterAddr, int64(*masterPort), int64(*size), *rank == 0)
	defer ts.Close()
	pg := torch.NewProcessGroupGloo(ts, int64(*rank), int64(*size))
	defer pg.Close()

	net := models.MLP()
	opt := torch.SGD(0.01, 0.5, 0, 0, false)
	params := net.Parameters()
	opt.AddParameters(params)
	for _, p := range params {
		pg.Broadcast([]torch.Tensor{p})
	}
	defer torch.FinishGC()

	epochs := 2
	step := 0

	for epoch := 0; epoch < epochs; epoch++ {
		trainLoader := mnistLoader(*dataset, 1500, *rank, *size, int64(epoch))
		for trainLoader.Scan() {
			data, label := trainLoader.Minibatch()

			opt.ZeroGrad()
			pred := net.Forward(data)
			loss := F.NllLoss(pred, label, torch.Tensor{}, -100, "mean")
			loss.Backward()

			if step%100 == 0 {
				log.Printf("epoch: %d, step: %d, loss: %f\n", epoch, step, loss.Item())
			}

			grads := getGrads(params)
			pg.AllReduceCoalesced(grads)

			opt.Step()
			step++
		}
	}
}
