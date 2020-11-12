# AllReduce Training Demo

In this demo, we train a MLP using MNIST datset.


## Build the Sample

Please follow the
[CONTRIBUTING.md](https://github.com/wangkuiyi/gotorch/blob/develop/CONTRIBUTING.md)
guide to build.

```bash
go install ./...
```

## Prepare the Data

```bash
cd $DATA_HOME
git clone https://github.com/myleott/mnist_png
cd mnist_png
tar zxvf mnist_png.tar.gz
export MNIST=$DATA_HOME/mnist_png/mnist_png
```

There are two directories, training and testing. 

We first create a label file for the dataset,

```
vim $MNIST/label.txt
```

and write the following content.

```txt
0
1
2
3
4
5
6
7
8
9
```

Then, we convert the training data into RecordIO format.

```bash
$GOPATH/bin/image-recordio-gen -label=$MNIST/label.txt -dataset=$MNIST/training -output=$MNIST/train_record -recordsPerShard=1500
```

The 60000 samples in the training directory are converted into 40 RecordIO shards in `train_record` directory. Each RecordIO shard contains 1500 samples.

## Launch Training Processes

We launch 2 training processes in a single node.

```bash
$GOPATH/bin/launch -nprocPerNode=2 -masterAddr=127.0.0.1 -masterPort=11111 -trainingCmd="$GOPATH/bin/allreduce -dataset=$MNIST/train_record -recordsPerShard=1500"
```

You could find two log files for each training process.

```bash
cat 0.log
2020/11/12 10:53:23 epoch: 0, step: 0, loss: 2.371462
2020/11/12 10:53:24 epoch: 0, step: 100, loss: 0.376211
2020/11/12 10:53:26 epoch: 0, step: 200, loss: 0.265234
2020/11/12 10:53:27 epoch: 0, step: 300, loss: 0.242690
2020/11/12 10:53:28 epoch: 0, step: 400, loss: 0.302058
2020/11/12 10:53:29 epoch: 1, step: 500, loss: 0.214308
2020/11/12 10:53:31 epoch: 1, step: 600, loss: 0.335965
2020/11/12 10:53:33 epoch: 1, step: 700, loss: 0.189105
2020/11/12 10:53:35 epoch: 1, step: 800, loss: 0.317435
2020/11/12 10:53:37 epoch: 1, step: 900, loss: 0.250716
```



