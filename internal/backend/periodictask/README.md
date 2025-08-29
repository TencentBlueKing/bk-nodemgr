# periodictask

## 设计意图

1. 实现非业务层的系统周期性任务调度

## 功能边界

1. 此包负责：
    - backend服务的周期任务
2. 此包不负责：
    - 业务层的周期任务，业务逻辑请交由schedule workflow处理
