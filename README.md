# 6.824
6.824

> 日志处理
> https://blog.josejg.com/debugging-pretty/

  pip install typer
  pip install rich 


## raft

#### 2A : 领导人选举 

>1、 git 提交 `* b75ced3 (HEAD -> master, origin/master) just pass 2A-1` 通过 2A 的第一个测试,有微小的概率通过第二和第三个
>
>2、 rpc 配合goroutine、wg、chan ,进行选举超时处理，进行心跳超时处理。（不然rpc会阻塞）。
> 当前情况，有 2/40 `2A-2`会失败, 2/40 `2A-3`会失败。连续测试40次。
>
> `````go
> VERBOSE=1 go test -run  2A   |  ./dslogs -c 3 
> VERBOSE=0 go test -run  2A
> VERBOSE=1 go test -run  2A  > output.log 
> ./dslogs -c 3 output.log
> ``````
> 
> 通过2A的全部测试


#### 2B : 实现领导者和追随者代码以附加新的日志条目

