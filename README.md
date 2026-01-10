# GoCache

一个用 Go 语言实现的高性能键值存储系统，兼容 RESP 协议。

## 当前状态

**迭代 1: MVP 最小可用版本** - 进行中

### 已完成功能 ✅

- [x] 项目基础设施（配置系统、日志系统）
- [ ] RESP 协议解析器
- [ ] 并发字典（分片锁）
- [ ] String 数据结构
- [ ] 数据库引擎
- [ ] TCP 服务器
- [ ] AOF 持久化

### 性能目标

- QPS ≥ 50,000 (MVP)
- P99 延迟 < 5ms (MVP)
- 最终目标 QPS ≥ 100,000, P99 < 1ms

## 快速开始

### 安装

```bash
git clone https://github.com/wangbo/gocache.git
cd gocache
go mod download
```

### 配置

编辑 `gocache.conf` 配置文件：

```conf
bind 127.0.0.1
port 6379
databases 16
appendonly yes
appendfsync everysec
```

### 运行

```bash
# 构建
go build -o gocache

# 运行
./gocache
```

### 测试连接

使用 redis-cli 或 telnet 连接：

```bash
# 使用 redis-cli
redis-cli -p 6379

# 或使用 telnet
telnet 127.0.0.1 6379
```

### 示例命令

```
SET mykey "Hello GoCache"
GET mykey
DEL mykey
```

## 项目结构

```
gocache/
├── cmd/                    # 命令行入口
├── config/                 # 配置管理
├── logger/                 # 日志系统
├── protocol/               # RESP 协议
├── dict/                   # 并发字典
├── datastruct/             # 数据结构
├── database/               # 数据库引擎
├── net/                    # 网络层
├── aof/                    # AOF 持久化
├── test/                   # 测试
├── data/                   # 数据文件目录
├── go.mod
├── go.sum
├── gocache.conf            # 配置文件
└── README.md
```

## 开发

### 运行测试

```bash
# 运行所有测试
go test ./... -v

# 运行测试并检查覆盖率
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# 运行竞态检测
go test ./... -race
```

### 代码检查

```bash
# 格式化代码
gofmt -l .

# 静态分析
go vet ./...
```

## 配置说明

### 核心配置

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| bind | 127.0.0.1 | 监听地址 |
| port | 6379 | 监听端口 |
| databases | 16 | 数据库数量 |
| maxclients | 10000 | 最大客户端连接数 |
| timeout | 0 | 客户端空闲超时（秒） |

### 持久化配置

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| appendonly | no | 是否启用 AOF |
| appendfilename | appendonly.aof | AOF 文件名 |
| appendfsync | everysec | AOF 刷盘策略 (always/everysec/no) |
| dbfilename | dump.rdb | RDB 文件名 |

### 日志配置

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| loglevel | info | 日志级别 (debug/info/warn/error) |
| logfile | "" | 日志文件（空字符串表示标准输出） |

## 迭代计划

详见 [迭代开发计划](/Users/wangbo/.claude/plans/enumerated-dazzling-panda.md)

### 迭代 1: MVP 最小可用版本（当前）
- TCP 服务器和 RESP 协议
- String 数据结构
- 并发安全存储
- 基础 TTL
- AOF 持久化

### 迭代 2: 数据结构增强
- List/Set/Hash/ZSet 数据结构

### 迭代 3: 过期与淘汰策略
- 时间轮 TTL
- 内存淘汰策略（LRU/LFU）

### 迭代 4: 事务支持
- MULTI/EXEC 事务
- WATCH 乐观锁

### 迭代 5: 持久化增强
- AOF 重写
- RDB 快照
- 混合持久化

### 迭代 6: 主从复制
- 主从同步
- PSYNC 增量同步

### 迭代 7: 安全与监控
- AUTH 认证
- SLOWLOG 慢查询
- INFO 监控

### 迭代 8: 性能优化
- 性能调优
- 编码优化（ziplist/intset）

### 迭代 9: 测试与文档
- 完善测试覆盖
- 完善文档

## 测试覆盖率

当前测试覆盖率：

- config: **85.5%**
- logger: **81.4%**

目标：所有模块 ≥ 80%

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 参考资料

- [RESP 协议规范](https://redis.io/docs/reference/protocol-spec/)
- Redis 设计思想
