# FastApi
[![GoDoc](https://godoc.org/github.com/TianQinS/fastapi?status.svg)](https://godoc.org/github.com/TianQinS/fastapi)

**简介**

高性能，协程池，在线更新。
>1. post：协程池，基于通道(不活跃调用场景)和基于无锁队列(高负载场景)。
>2. timer：提供了不同精度的延迟函数实现，crontab设置格式同unix crontab命令字符串，降低内置计时器堆负载。
>3. hotfix：提供对register functions(结合commserv模块)和public variable的在线更新功能。


**基础模块**

---------------------------------------
  * [basic module](#fastapi)
	* [Timer](#timer)
	* [Post](#post)
	* [Hotfix](#hotfix)
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
		fmt.Printf("Timer every 5 second data=%d\n", d)
	}, 1)
	t2 := timer.AddCrontab("1-50/2 * * * *", func() {
		fmt.Printf("crontab %s is fired\n", timer.GetQcTime())
		t1.Cancel()
	})
	timer.CallOut(3, func(d int){
		fmt.Printf("CallOut after 3 second data=%d\n", d)
	}, 1)
	time.Sleep(time.Minute * 5)
	t2.Cancel()
}
```

- In general, it's commonly used functions includes AddCallback, AddTimer, CallOut and Cancel.
- It's not just that a timer construction of min-heap such as time's Tick, but class the tasks by time interval adapt to different mission scenes. 

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

### Hotfix

Hotfix evaluate the script in the context of interpreter and change variables in runtime context with parameter passing.

```go
// forTest is used for scripts's sample2.
func (rpc *RPCModule) forTest(arg1, arg2 string) string {
	return fmt.Sprintf("args: %s, %s.", arg1, arg2)
}

func (rpc *RPCModule) OnInit(topic string, qSize uint64) {
	rpc.topic = topic
	rpc.BaseModule.Init(qSize)
	// for test script.
	rpc.RegisterRpc("test", rpc.forTest)
}
```
- A rpc module written by golang with a test function registered for hotfix testing.

```python
if __name__ == "__main__":
	u"""for hotfix testing."""
	rpc = RDB(**config.REDIS_CONFIG)
	rpc.register("callback", rpc.test)
	rpc.call("rpc", "test", "callback", "arg1", "arg2")
	time.sleep(0.1)
	rpc.patch("""package patch

import (
	"fmt"
	"github.com/TianQinS/commserv/event"
	"github.com/TianQinS/commserv/module"
)

func Process(ev *event.EventMgr) error {
	if mod := ev.GetModule("rpc"); mod != nil {
		rmod := mod.(*module.RPCModule)
		rmod.RegisterRpc("test", func(arg1, arg2 string) string {
			return arg1 + arg2 + arg1 + arg2
		})
	}
	fmt.Println("Patch process finished.")
	return nil
}
	
""")
	rpc.call("rpc", "test", "callback", "arg1", "arg2")
```

- A remote call with callback written by python.
- A call function call server's rpc module with arguments and callback. 
- A patch function send a simple script to server and update rpc module's register functions.
