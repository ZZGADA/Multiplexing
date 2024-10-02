package code

import "fmt"

type Home struct {
	Address string
}

type Person struct {
	Name string
	Age  int
	Home Home
}

func newPerson1() Person {
	person1 := Person{
		Name: "Jim",
		Age:  18,
		Home: Home{Address: "TAL"},
	}

	// 打印结构体对象的内存地址
	fmt.Printf("结构体对象 person1 的内存地址: %p\n", &person1)
	fmt.Printf("结构体对象 person1.Home 的内存地址: %p\n", &person1.Home)
	return person1
}

func newPerson2() *Person {
	person2 := &Person{
		Name: "Jim",
		Age:  18,
		Home: Home{Address: "TAL"},
	}

	// 打印结构体对象的内存地址
	fmt.Printf("结构体对象 person2 的内存地址: %p\n", person2)
	fmt.Printf("结构体对象 person2.Home 的内存地址: %p\n", &person2.Home)
	return person2
}

/*
  - 对于结构体 如果是返回对象 是对对象进行复制(深拷贝) 返回指针 则是地址不变
    结构体对象 person1 的内存地址: 0x140000c9ec0
    结构体对象 person1 的内存地址 from newPerson1: 0x140000c9ea8
    结构体对象 person2 的内存地址: 0x140000c9ed8
    结构体对象 person2 的内存地址 from newPerson2: 0x140000c9ed8
*/
func testStructureReturn() {
	person1 := newPerson1()
	fmt.Printf("结构体对象 person1 的内存地址 from newPerson1: %p\n", &person1)
	fmt.Printf("结构体对象 person1.Home 的内存地址: %p\n", &person1.Home)

	person2 := newPerson2()
	fmt.Printf("结构体对象 person2 的内存地址 from newPerson2: %p\n", person2)
	fmt.Printf("结构体对象 person2.Home 的内存地址: %p\n", &person2.Home)
}
