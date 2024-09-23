package code

import (
	"Multiplexing_/src/entity/bo"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	PoolSize   int64          // 协程池
	WorkerSize int64          // 工作者数量
	ch         chan int       // 通信信号量
	chWorker   chan int       // 工作者处理
	wg         sync.WaitGroup // wg协程阻塞
)

func init() {
	PoolSize = 32
	WorkerSize = 0
	ch = make(chan int)
	chWorker = make(chan int)
	wg = sync.WaitGroup{}
}

func handleWorkerSize() {
	for {
		select {
		case <-chWorker:
			WorkerSize--
		}
	}
}

func Multiplexing() {
	log.Printf("协程池大小为: %d", PoolSize)
	// 开启异步线程监听 对chWorker 进行唯一处理
	go handleWorkerSize()

	for {
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
		// 直接返回
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
