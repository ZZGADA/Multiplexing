package code

import (
	"Multiplexing_/src/entity/bo"
	"fmt"
	"sync"
)

func TestConcrete() {
	goFuncNum := 1000
	var wg sync.WaitGroup
	wg.Add(goFuncNum)

	for i := 0; i < goFuncNum; i++ {

		go func() {
			defer wg.Done()
			affectRows := bo.FuFileBOMapperImpl.UpdateFileName("test", 62)
			if affectRows == 0 {
				fmt.Println("something went wrong", i)
			}
		}()

	}

	wg.Wait()
}
