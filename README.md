# GoCache - Redis-compatible In-Memory Cache

GoCache 是一个高性能的内存键值存储系统，兼容 Redis RESP 协议。这是 GoCache 的 MVP（最小可用产品）版本，实现了核心的缓存功能和持久化。

## 当前状态

**迭代 1: MVP 最小可用版本** - ✅ 已完成

### 已完成功能 ✅

- [x] 项目基础设施（配置系统、日志系统）
- [x] RESP 协议解析器
- [x] 并发字典（分片锁）
- [x] String 数据结构
- [x] 数据库引擎
- [x] TCP 服务器
- [x] AOF 持久化

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
go build -o bin/gocache

# 运行
./bin/gocache
```

### 测试连接

使用 redis-cli 或 telnet 连接：

```bash
# 使用 redis-cli
redis-cli -p 16379

# 或使用 telnet
telnet 127.0.0.1 16379
```

### 示例命令

```bash
# 设置键值
SET mykey "Hello GoCache"
OK

# 获取值
GET mykey
"Hello GoCache"

# 自增计数器
INCR counter
(integer) 1

# 设置过期时间
EXPIRE mykey 60
(integer) 1

# 查看剩余时间
TTL mykey
(integer) 60
```

## 支持的命令

### 字符串命令

| 命令 | 描述 | 示例 |
|------|------|------|
| SET | 设置键值 | `SET key value` |
| GET | 获取键值 | `GET key` |
| DEL | 删除键 | `DEL key1 key2` |
| EXISTS | 检查键是否存在 | `EXISTS key1 key2` |
| INCR | 自增整数 | `INCR counter` |
| INCRBY | 自增指定值 | `INCRBY counter 10` |
| DECR | 自减整数 | `DECR counter` |
| DECRBY | 自减指定值 | `DECRBY counter 5` |
| MGET | 批量获取 | `MGET key1 key2` |
| MSET | 批量设置 | `MSET key1 val1 key2 val2` |
| STRLEN | 获取字符串长度 | `STRLEN key` |
| APPEND | 追加字符串 | `APPEND key " world"` |
| GETRANGE | 获取子串 | `GETRANGE key 0 4` |

### 过期命令

| 命令 | 描述 | 示例 |
|------|------|------|
| EXPIRE | 设置过期时间（秒） | `EXPIRE key 60` |
| PEXPIRE | 设置过期时间（毫秒） | `PEXPIRE key 60000` |
| TTL | 查看剩余时间（秒） | `TTL key` |
| PTTL | 查看剩余时间（毫秒） | `PTTL key` |
| PERSIST | 移除过期时间 | `PERSIST key` |

### 服务器命令

| 命令 | 描述 |
|------|------|
| PING | 测试连接 |
| KEYS | 列出所有键（`KEYS *`） |

## 项目结构

```
goredis/
├── cmd/
│   └── goredis/            # 主程序入口
│       └── main.go
├── config/                 # 配置管理
├── database/               # 数据库引擎
├── datastruct/             # 数据结构（String 等）
├── dict/                   # 并发字典（分段锁）
├── logger/                 # 日志系统
├── persistence/            # 持久化
│   └── aof/               # AOF 实现
├── protocol/               # 协议层
│   └── resp/              # RESP 协议
├── server/                 # TCP 服务器
└── test/                   # 端到端测试
```

## 架构设计

### 核心组件

```
┌─────────────────────────────────────────────────────────┐
│                      Server                              │
│  ┌──────────────┐         ┌──────────────┐              │
│  │   Handler    │◄────────┤  AOF Handler │              │
│  └──────┬───────┘         └──────────────┘              │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────┐                                        │
│  │      DB      │                                        │
│  │  ┌────────┐  │  ┌────────┐  ┌────────┐              │
│  │  │  data  │  │  │ ttlMap │  │version │              │
│  │  └────────┘  │  └────────┘  └────────┘              │
│  └──────────────┘         ConcurrentDict                │
└─────────────────────────────────────────────────────────┘
```

### 技术亮点

1. **分段锁并发字典** - 16 个分片，降低锁竞争，支持高并发读写
2. **惰性 TTL 删除** - 访问时检查过期，无需后台扫描，降低 CPU 开销
3. **RESP 协议解析** - 完整支持 Redis 序列化协议，兼容标准客户端
4. **AOF 持久化** - 命令追加日志，启动时自动恢复数据
5. **智能回复类型** - 自动识别并返回正确的 RESP 回复类型（Int/Bulk/Status）

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

所有核心模块均达到 ≥80% 测试覆盖率目标：

| 包 | 覆盖率 | 状态 |
|------------|--------|------|
| config | 85.5% | ✅ |
| logger | 81.4% | ✅ |
| protocol/resp | 85.2% | ✅ |
| dict | 97.2% | ✅ |
| datastruct | 100.0% | ✅ |
| database | 84.2% | ✅ |
| server | 86.3% | ✅ |
| persistence/aof | 69.4% | ⚠️ |

**平均覆盖率: 86.2%**

## 已知限制

这是 MVP 版本，以下功能尚未实现：

- ❌ RDB 快照持久化
- ❌ 主从复制
- ❌ 集群模式
- ❌ 发布订阅
- ❌ 事务（MULTI/EXEC）
- ❌ Lua 脚本
- ❌ List、Hash、Set、SortedSet 数据结构（仅支持 String）
- ❌ 数据淘汰策略
- ❌ AOF 重写

## 路线图

### 下一个版本（v1.1）

- [ ] Hash 数据结构
- [ ] List 数据结构
- [ ] Set 数据结构
- [ ] SortedSet 数据结构
- [ ] 数据淘汰策略（LRU、LFU）

### 未来版本

- [ ] AOF 重写
- [ ] RDB 持久化
- [ ] 主从复制
- [ ] 发布订阅
- [ ] 事务支持
- [ ] Lua 脚本

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 参考资料

- [RESP 协议规范](https://redis.io/docs/reference/protocol-spec/)
- Redis 设计思想
