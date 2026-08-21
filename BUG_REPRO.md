# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

机房切换演练把 PostgreSQL、对象存储和消息代理等十二个就绪检查同时放行后，/readyz 明细偶尔少项，原本固定的展示顺序也会跳，race 构建还会报结果集合冲突；让依赖依次返回时则看不出异常。请修复并行汇总，每个已注册检查都要恰好保留一条结果并按注册顺序输出，成功状态和失败原因不能串到别的依赖。

## 含 Bug 版本

- 仓库：VanceMichael/go-label-19
- 仓库地址：https://github.com/VanceMichael/go-label-19.git
- parent SHA：9d20a633c58bab1988b776c3d1767e0726e87ee6

## 复现步骤

```bash
git clone -- https://github.com/VanceMichael/go-label-19.git bug-repro
cd bug-repro
git checkout --detach 9d20a633c58bab1988b776c3d1767e0726e87ee6
go test ./internal/health -run ^TestParallelReadinessPreservesEveryRegisteredCheck$ -count=1
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/health -run ^TestParallelReadinessPreservesEveryRegisteredCheck$ -count=1
--- FAIL: TestParallelReadinessPreservesEveryRegisteredCheck (0.00s)
    health_test.go:42: status 0 = {Name:dependency-11 OK:true Error: Duration:251.75µs}, want name=dependency-00 ok=false
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/health	0.031s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/health -run ^TestParallelReadinessPreservesEveryRegisteredCheck$ -count=1
--- FAIL: TestParallelReadinessPreservesEveryRegisteredCheck (0.00s)
    health_test.go:42: status 0 = {Name:dependency-11 OK:true Error: Duration:4.042µs}, want name=dependency-00 ok=false
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/health	0.001s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

同步屏障让十二个依赖检查同时返回时，Registry.Run 必须仍按 dependency-00 到 dependency-11 的注册顺序生成十二条结果；序号能被三整除的检查保留各自的 unavailable 原因，其余检查为成功且错误为空，不能遗漏、重复或串写。TestParallelReadinessPreservesEveryRegisteredCheck 需要在 -race 下由红转绿，同时保持真正并行执行；health 相关回归、全仓可执行测试和 go build ./... 应通过，不得改成串行、删改测试或放宽数量与顺序断言。
