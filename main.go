package main

import (
	"Multiplexing_/code"
	"fmt"
	"time"
)

func main() {
	fmt.Println("start jenkins job test ======> 睡眠5秒 准备执行Multiplexing方法")
	fmt.Println("======> 测试github回调函数")
	time.Sleep(5 * time.Second)
	code.Multiplexing()

	fmt.Println("code.Multiplexing() ======> 执行结束")

}
