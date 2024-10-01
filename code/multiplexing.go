package code

import (
	"Multiplexing_/src/entity/bo"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	PoolSize      int64          // 协程池
	WorkerSize    int64          // 工作者数量
	ch            chan int       // 通信信号量
	chWorker      chan int       // 工作者处理
	wg            sync.WaitGroup // wg协程阻塞
	iteratorNum   int64          // 限制迭代次数
	iteratorEndCh chan int       // 迭代终止监听
	ifEnd         bool
)

const endNum = 3

func init() {
	PoolSize = 32
	WorkerSize = 0
	ch = make(chan int)
	chWorker = make(chan int)
	wg = sync.WaitGroup{}
	iteratorNum = 0
	ifEnd = false
	iteratorEndCh = make(chan int)
}

func handleWorkerSize() {
	defer func() {
		fmt.Println("handleWorkerSize 协程结束")
	}()
	// iteratorEndCh 用于结束 handleWorkerSize 协程
	for {
		select {
		case <-chWorker:
			WorkerSize--
		case <-iteratorEndCh:
			return
		}
	}
}

func Multiplexing() {
	defer func() {
		close(ch)
		close(chWorker)
		close(iteratorEndCh)
		fmt.Println("关闭全部 channel 通道")
	}()

	log.Printf("协程池大小为: %d", PoolSize)
	// 开启异步线程监听 对chWorker 进行唯一处理
	go handleWorkerSize()

	for {
		if ifEnd {
			return
		}
		wg.Add(1)
		run()
		wg.Wait()
	}
}

func run() {
	defer wg.Done()
	log.Println("进入入方法查询")
	fmt.Printf("当前工作者的个数是： %d\n", WorkerSize)
	if WorkerSize >= PoolSize {
		log.Println("工作者过多 休息等待一下")
		time.Sleep(3 * time.Second)
		iteratorNum++
		// 直接返回
		return
	}

	// 终止结束
	if iteratorNum >= endNum {
		fmt.Println("Multiplexing 多路复用监听主线程结束")
		ifEnd = true
		iteratorEndCh <- 1 // 发送终止信号
		return
	}

	// 开启协程 ch阻塞等待
	WorkerSize++
	go handle()
	fmt.Printf("正在阻塞\n")
	<-ch
}

func handle() {
	defer func() {
		// 结束之后向chWorker发送消息用于结束worker数量
		chWorker <- 1
	}()

	file := bo.FuFileBOMapperImpl.GetOneFile("057ad0d6-6064-11ef-9d45-7e15eb14963e")
	fmt.Printf("resutl is %#v\n", file)
	fmt.Println("----------------------------------------------------------------")
	ch <- 1
	// 模拟后序耗时业务
	time.Sleep(10 * time.Second)
}
