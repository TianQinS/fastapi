# FastApi
[![GoDoc](https://godoc.org/github.com/TianQinS/fastapi?status.svg)](https://godoc.org/github.com/TianQinS/fastapi)

**简介**

高性能，容错好。
>1. 协程池封装，基于通道(不活跃调用场景)和基于无锁队列(高负载场景)。
>2. timer模块封装，提供了不同精度的延迟函数实现，crontab设置格式同unix crontab命令字符串。


**基础模块**

---------------------------------------
  * [basic module](#fastapi)
	* [Timer](#timer)
	* [Post](#post)
---------------------------------------

### Timer

The way to use timer is as follows:

```go
package main
import (
	"fmt"
	"time"
	"github.com/TianQinS/fastapi/timer"
)

func main() {
	timer.SetQcTime("2019-03-19 23:01:40")
	t1 := timer.AddTimer(time.Second*5, func(d int) {
		fmt.Println(d)
	}, 1)
	t2 := timer.AddCrontab("1-50/2 * * * *", func() {
		fmt.Printf("crontab %s\n", timer.GetQcTime())
		t1.Cancel()
	})
	time.Sleep(time.Minute * 5)
	t2.Cancel()
}
```

- In general, it's commonly used functions includes AddCallback, AddTimer, CallOut and Cancel.

### Post

Post execute jobs by group in asynchronous task queue.

```go
package main

import (
	"fmt"

	"github.com/TianQinS/fastapi/post"
)

func main(){
	p := post.GPost
	
	// aysnc job with chan.
	p.PutJob("test", func(i int) {
		fmt.Println(i)
	}, 2)
	// aysnc job with lock-free queue.
	p.PutQueue(func(i int){
		fmt.Println(i)
	}, 2)
	// aysnc job with strict mode.
	p.PutQueueStrict(func(args ...interface{}){
		fmt.Println(args[0].(int))
	}, 2)
}
```

- It is right that functions should be called in different mode based on the load.