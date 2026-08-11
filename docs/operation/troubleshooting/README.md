# 疑难问题排查

记录 bk-nodemgr 运行过程中不易从表面错误判断根因的故障案例。每个案例包含现象、根因、确认方法和处理原则。

## [Redis Cluster 跨环境消费 Machinery 任务](redis_cluster_machinery_queue_isolation.md)

两套 bk-nodemgr 环境共用 Redis Cluster 时，因 `redis.db` 无法隔离且 Machinery 队列同名，导致任务被另一套环境消费并报 `oper_inst_data not found`。
