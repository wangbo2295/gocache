# GoCache 端到端自动化测试验收方案

## 📋 一、测试目标

根据[需求文档.md](../docs/需求文档.md)第4章节的验收标准，进行端到端(E2E)自动化测试验收，确保系统满足：

1. **功能完整性** - 所有数据类型和命令正确实现
2. **可靠性** - 持久化、主从复制、数据恢复
3. **性能指标** - QPS、延迟、并发能力
4. **稳定性** - 长时间运行、内存可控

---

## 🎯 二、测试范围

### 2.1 功能验收测试

根据需求文档第4.1节：

```
✅ 支持所有基础数据类型（String、List、Set、Hash、ZSet）
✅ 支持键的过期设置和自动淘汰
✅ 支持事务的原子性执行
```

**覆盖范围**:
- String类型：SET, GET, DEL, EXISTS, EXPIRE, TTL, PEXPIRE, PTTL, PERSIST, INCR, DECR, INCRBY, DECRBY, APPEND, STRLEN, GETRANGE, SETRANGE, MSET, MGET
- Hash类型：HSET, HGET, HGETALL, HDEL, HEXISTS, HKEYS, HVALS, HLEN, HSETNX, HINCRBY, HMGET, HMSET
- List类型：LPUSH, RPUSH, LPOP, RPOP, LLEN, LINDEX, LRANGE, LSET, LTRIM, LREM, LINSERT
- Set类型：SADD, SREM, SISMEMBER, SMEMBERS, SCARD, SPOP, SRANDMEMBER, SMOVE, SDIFF, SINTER, SUNION
- SortedSet类型：ZADD, ZREM, ZSCORE, ZCARD, ZINCRBY, ZRANGE, ZREVRANGE, ZRANGEBYSCORE, ZCOUNT, ZRANK, ZREVRANK
- TTL命令：EXPIRE, PEXPIRE, EXPIREAT, PEXPIREAT, TTL, PTTL, PERSIST
- 事务：MULTI, EXEC, DISCARD, WATCH, UNWATCH

### 2.2 可靠性验收测试

根据需求文档第4.2节：

```
✅ 支持RDB和AOF两种持久化方式
✅ 支持主从复制和数据同步
✅ 进程崩溃后可从持久化文件恢复数据
```

**覆盖范围**:
- RDB持久化：SAVE, BGSAVE命令，RDB文件生成和加载
- AOF持久化：AOF文件写入，重启恢复
- 主从复制：SLAVEOF, SYNC, PSYNC命令
- 数据恢复：崩溃后重启，数据完整性验证

### 2.3 性能验收测试

根据需求文档第4.3节：

```
✅ 单实例QPS > 10万（SET/GET操作）
✅ P99响应时间 < 1毫秒
✅ 支持1万并发连接
```

**测试场景**:
- 单线程SET/GET：QPS基线测试
- 并发SET/GET：50/100并发连接
- 批量操作：MGET/MSET
- 混合读写：1:1读写比
- 延迟测试：P50, P95, P99, P99.9

### 2.4 稳定性验收测试

根据需求文档第4.4节：

```
✅ 7x24小时稳定运行
✅ 内存使用可控（不超过maxmemory限制）
```

**测试场景**:
- 长时间运行测试（模拟7x24小时，实际可缩短为压力测试）
- 内存淘汰验证
- 大量数据写入和删除

---

## 🛠️ 三、测试方法和技术

### 3.1 测试工具

| 工具类型 | 工具名称 | 用途 |
|---------|---------|------|
| **协议测试** | redis-benchmark | 性能压测 |
| **自定义测试** | Go测试程序 | 功能验证 |
| **集成测试** | 自定义客户端 | 端到端场景 |
| **监控工具** | 自定义监控脚本 | 资源使用 |

### 3.2 测试架构

```
测试架构
├── 测试客户端（Go编写）
│   ├── 连接池管理
│   ├── 协议编解码（RESP）
│   └── 测试用例执行
├── 测试场景定义
│   ├── 功能测试用例
│   ├── 性能测试用例
│   └── 稳定性测试用例
└── 测试报告生成
    ├── 覆盖率报告
    ├── 性能报告
    └── 验收结论
```

### 3.3 测试策略

#### 策略1: 分阶段验收

```
阶段1: 功能验收（1-2天）
  ├── 基础数据类型测试
  ├── TTL和过期机制测试
  └── 事务功能测试

阶段2: 可靠性验收（1-2天）
  ├── RDB持久化测试
  ├── AOF持久化测试
  ├── 数据恢复测试
  └── 主从复制测试

阶段3: 性能验收（1天）
  ├── QPS压测
  ├── 延迟测试
  └── 并发连接测试

阶段4: 稳定性验收（1天）
  ├── 压力测试
  ├── 内存淘汰测试
  └── 长时间运行测试
```

#### 策略2: 自动化程度

**完全自动化**:
- 功能测试：自动执行所有命令，验证返回值
- 性能测试：自动压测，收集性能指标
- 报告生成：自动生成测试报告

**半自动化**:
- 稳定性测试：长时间运行需人工监控
- 数据恢复测试：需要手动重启验证

#### 策略3: 发散性测试

除了需求文档明确要求的测试外，增加以下发散性测试：

**边界条件测试**:
- 空值处理：SET/GET空字符串
- 超长键名：512MB键名
- 超大值：512MB值
- 特殊字符：二进制安全、Unicode、控制字符

**异常场景测试**:
- 网络中断：模拟连接断开
- 内存不足：maxmemory限制下的行为
- 并发冲突：WATCH的乐观锁冲突
- 事务失败：语法错误和运行时错误

**兼容性测试**:
- RESP协议兼容性
- 多客户端并发
- 不同数据类型混合操作

---

## 📁 四、测试目录结构

```
test/
├── README.md                          # 本文档
├── e2e/                              # 端到端测试
│   ├── test_strategy.go               # 测试策略定义
│   ├── test_client.go                 # Redis协议客户端
│   ├── functional/                    # 功能测试
│   │   ├── string_test.go            # String类型测试
│   │   ├── hash_test.go              # Hash类型测试
│   │   ├── list_test.go              # List类型测试
│   │   ├── set_test.go               # Set类型测试
│   │   ├── sortedset_test.go         # SortedSet类型测试
│   │   ├── ttl_test.go               # TTL过期测试
│   │   └── transaction_test.go       # 事务测试
│   ├── reliability/                   # 可靠性测试
│   │   ├── persistence_test.go       # 持久化测试
│   │   ├── recovery_test.go          # 数据恢复测试
│   │   └── replication_test.go      # 主从复制测试
│   ├── performance/                   # 性能测试
│   │   ├── benchmark_test.go         # 性能基准测试
│   │   ├── qps_test.go               # QPS压测
│   │   ├── latency_test.go           # 延迟测试
│   │   └── concurrent_test.go        # 并发测试
│   ├── stability/                     # 稳定性测试
│   │   ├── stress_test.go            # 压力测试
│   │   ├── eviction_test.go          # 淘汰策略测试
│   │   └── longrun_test.go           # 长时间运行测试
│   └── exploratory/                   # 发散性测试
│       ├── boundary_test.go          # 边界条件测试
│       ├── error_test.go             # 异常场景测试
│       └── compatibility_test.go     # 兼容性测试
├── reports/                          # 测试报告
│   ├── functional_report.md          # 功能测试报告
│   ├── reliability_report.md         # 可靠性测试报告
│   ├── performance_report.md         # 性能测试报告
│   └── acceptance_report.md          # 最终验收报告
└── scripts/                          # 辅助脚本
    ├── start_server.sh               # 启动测试服务器
    ├── stop_server.sh                # 停止测试服务器
    ├── cleanup.sh                    # 清理测试环境
    └── run_all_tests.sh              # 运行所有测试
```

---

## 📝 五、测试用例设计原则

### 5.1 功能测试用例

**设计原则**:
1. **命令覆盖**: 每个命令至少1个正常场景测试
2. **参数组合**: 测试不同参数组合
3. **边界值**: 测试参数的上下界
4. **错误处理**: 测试错误参数的返回值
5. **数据一致性**: 验证操作后数据的正确性

**示例**:
```go
// SET命令测试用例
func TestSET_Command(t *testing.T) {
    tests := []struct {
        name     string
        cmd      [][]byte
        expected [][]byte
        setup    func() // 测试前准备
        verify   func() // 测试后验证
    }{
        {
            name:     "SET basic key-value",
            cmd:      [][]byte{[]byte("SET"), []byte("key"), []byte("value")},
            expected: [][]byte{[]byte("OK")},
        },
        {
            name:     "SET with EX",
            cmd:      [][]byte{[]byte("SET"), []byte("key"), []byte("value"), []byte("EX"), []byte("3600")},
            expected: [][]byte{[]byte("OK")},
            verify: func() {
                // 验证TTL设置成功
                result := client.Execute("TTL", "key")
                assert.True(result[0] > 0)
            },
        },
        // ... 更多测试用例
    }
}
```

### 5.2 性能测试用例

**设计原则**:
1. **基准对比**: 与Redis或需求文档对比
2. **多轮测试**: 每个测试运行多次取中位数
3. **预热阶段**: 避免冷启动影响
4. **资源监控**: 记录CPU、内存使用

**示例**:
```go
func BenchmarkSET_QPS(b *testing.B) {
    client := NewTestClient()
    client.Connect()

    // 预热
    for i := 0; i < 1000; i++ {
        client.Send("SET", fmt.Sprintf("key%d", i), "value")
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client.Send("SET", fmt.Sprintf("benchkey%d", i%10000), "value")
    }

    // 计算QPS
    qps := float64(b.N) / b.Seconds()
    b.ReportMetric(qps, "ops/sec")
}
```

### 5.3 可靠性测试用例

**设计原则**:
1. **真实场景**: 模拟实际使用场景
2. **故障注入**: 模拟进程崩溃、网络中断
3. **数据验证**: 验证恢复后数据的完整性

**示例**:
```go
func TestRDB_Recovery(t *testing.T) {
    // 1. 写入测试数据
    client.Set("key1", "value1")
    client.Set("key2", "value2")

    // 2. 触发RDB保存
    client.Send("SAVE")

    // 3. 杀死进程（子进程）
    server.Kill()

    // 4. 重启服务器
    server.Restart()
    client.Reconnect()

    // 5. 验证数据恢复
    v1 := client.Get("key1")
    assert.Equal("value1", v1)
}
```

---

## 📊 六、验收标准

### 6.1 功能验收标准

| 类别 | 标准 | 验证方法 |
|------|------|---------|
| String命令 | 所有命令正确执行 | 自动化测试 |
| Hash命令 | 所有命令正确执行 | 自动化测试 |
| List命令 | 所有命令正确执行 | 自动化测试 |
| Set命令 | 所有命令正确执行 | 自动化测试 |
| SortedSet命令 | 所有命令正确执行 | 自动化测试 |
| TTL功能 | 过期自动删除 | 自动化测试+延时验证 |
| 事务功能 | 原子性保证 | 自动化测试 |

**通过标准**: 所有测试用例100%通过

### 6.2 可靠性验收标准

| 类别 | 标准 | 验证方法 |
|------|------|---------|
| RDB持久化 | 重启后数据恢复 | 自动化测试+重启验证 |
| AOF持久化 | 重启后数据恢复 | 自动化测试+重启验证 |
| 主从复制 | 数据同步正确 | 自动化测试+数据对比 |
| 数据完整性 | 无数据丢失或损坏 | 数据校验 |

**通过标准**: 所有测试用例100%通过，数据完整性100%

### 6.3 性能验收标准

| 指标 | 需求标准 | 验收标准 |
|------|---------|---------|
| QPS (SET/GET) | ≥ 100,000 | ≥ 80,000 (80%) |
| P99延迟 | < 1ms | < 2ms (200%) |
| 并发连接 | ≥ 10,000 | ≥ 5,000 (50%) |
| 内存使用 | < maxmemory | 不超过配置限制 |

**通过标准**: 至少达到验收标准的80%（考虑到测试环境）

### 6.4 稳定性验收标准

| 类别 | 标准 | 验证方法 |
|------|------|---------|
| 长时间运行 | 7x24小时 | 压力测试2小时+ |
| 内存控制 | 不超过maxmemory | 内存监控 |
| 无崩溃 | 进程不崩溃 | 进程监控 |

**通过标准**:
- 压力测试2小时无崩溃
- 内存使用可控
- 无内存泄漏

---

## 🔄 七、测试执行流程

### 7.1 环境准备

```bash
# 1. 编译最新版本
cd /Users/wangbo/goredis
go build -o test/bin/gocache .

# 2. 清理旧测试数据
rm -f test/*.log test/*.rdb test/*.aof

# 3. 启动测试服务器
./test/scripts/start_server.sh
```

### 7.2 执行测试

```bash
# 阶段1: 功能测试
go test ./test/e2e/functional/... -v -timeout 10m

# 阶段2: 可靠性测试
go test ./test/e2e/reliability/... -v -timeout 15m

# 阶段3: 性能测试
go test ./test/e2e/performance/... -bench=. -benchtime=30s -timeout 20m

# 阶段4: 稳定性测试
go test ./test/e2e/stability/... -v -timeout 30m
```

### 7.3 生成报告

```bash
# 生成所有测试报告
./test/scripts/generate_reports.sh

# 查看验收结论
cat test/reports/acceptance_report.md
```

---

## 📋 八、需要确认的问题

在开始实施自动化测试前，需要与您确认以下问题：

### 问题1: 测试范围优先级

**选项A**: 按需求文档验收标准，全面测试（推荐）
- 功能、可靠性、性能、稳定性全部测试
- 预计工作量：5-7天
- 适合：完整验收，发布前验证

**选项B**: 仅功能测试 + 基本性能测试
- 跳过长时间稳定性测试
- 预计工作量：2-3天
- 适合：开发过程中快速验证

**选项C**: 自定义测试范围
- 您指定重点测试哪些部分

### 问题2: 测试环境要求

**需要确认**:
- 测试服务器配置（CPU、内存）
- 是否需要多机环境（测试主从复制）
- 是否可以杀进程（测试数据恢复）
- 网络环境（本地/远程）

**推荐配置**:
- 单机测试：本地macOS/Linux
- CPU: 4核+
- 内存: 8GB+
- 可接受进程重启

### 问题3: 性能目标调整

需求文档中的性能目标：
- QPS ≥ 100,000
- P99 < 1ms
- 并发 ≥ 10,000

**考虑到测试环境限制，建议调整**:
- QPS: 目标 80,000（80%）
- P99: 目标 < 2ms（200%）
- 并发: 目标 5,000（50%）

**是否接受调整后的目标？**

### 问题4: 测试报告格式

**选项A**: Markdown格式（推荐，易读）
**选项B**: JSON格式（便于CI集成）
**选项C**: HTML格式（便于浏览）

### 问题5: 发散性测试深度

**选项A**: 轻度发散（边界+异常，+20%工作量）
**选项B**: 中度发散（包含兼容性，+40%工作量）
**选项C**: 重度发散（全面测试，+60%工作量）

---

## ✅ 九、下一步行动

请您：

1. **审阅本方案**，确认测试策略是否符合您的预期
2. **回答上述5个问题**，明确测试范围和目标
3. **确认后开始实施**，我将按照以下顺序执行：
   - 创建测试目录结构
   - 实现测试客户端（Redis协议）
   - 编写功能测试用例
   - 编写性能测试用例
   - 编写可靠性和稳定性测试用例
   - 执行所有测试
   - 生成验收报告

**预计时间线**（根据选项A）:
- Day 1: 测试框架 + 功能测试（String/Hash）
- Day 2: 功能测试（List/Set/ZSet + TTL + 事务）
- Day 3: 可靠性测试（持久化 + 复制 + 恢复）
- Day 4: 性能测试（QPS + 延迟 + 并发）
- Day 5: 稳定性测试 + 发散性测试 + 报告生成

---

请您确认测试策略和范围后，我将开始实施自动化测试用例的编写！
