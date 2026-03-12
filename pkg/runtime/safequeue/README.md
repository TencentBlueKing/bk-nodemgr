# safequeue

## 设计意图
提供泛型的并发安全 FIFO 队列，解决 channel 容量固定、满时阻塞的问题，支持动态增长且不阻塞写入方。

## 功能边界
1. 此包负责：
    - 线程安全的入队（Enqueue）、出队（Dequeue）、查看队首（Peek）
    - 队列长度查询和空队列判断
2. 此包不负责：
    - 持久化、分布式队列
    - 消费者阻塞等待（无元素时立即返回，不阻塞）

## 设计考量
1. 基于 `sync.Mutex` + slice 实现，写入时自动扩容，不会像 buffered channel 那样在队列满时阻塞生产者
2. 使用 Go 泛型（`SafeQueue[T any]`），可存放任意类型元素，无需类型断言
3. 零值即可用，无需构造函数

## 使用限制
1. 出队和 Peek 在队列为空时返回 `(zero, false)`，调用方需检查第二个返回值
2. 不提供阻塞等待语义，如需消费者阻塞等待新元素，应使用 channel

## 使用示例
```go
q := &safequeue.SafeQueue[int]{}

q.Enqueue(1)
q.Enqueue(2)

v, ok := q.Dequeue() // v == 1, ok == true
v, ok = q.Peek()     // v == 2, ok == true（不移除）
fmt.Println(q.Len()) // 1
```

## 演进方向
1. 按需增加批量入队/出队方法
2. 按需增加容量上限和淘汰策略
