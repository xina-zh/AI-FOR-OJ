# AI For OJ

AI 算法题实验平台的 Go 项目骨架。当前阶段重点是先保证 OJ 基础设施、AI 实验能力和可观测能力可以在同一工程里平稳演进。

## 当前阶段总结

当前阶段的工程总结、已验证能力、系统语义、已知限制和下一步候选方向，统一整理在：

- [docs/dev_progress.md](/home/xina/projects/AI-For-Oj/docs/dev_progress.md)

这份文档用于后续开发直接接续当前真实进度，不是产品宣传材料。

## 当前数据库迁移策略

当前阶段使用 GORM `AutoMigrate`，原因是项目仍在快速迭代，核心表结构大概率还会继续调整。这样可以优先推进 OJ、Judge、LLM、Experiment 主流程，减少早期 migration 维护成本。

后续当表结构趋于稳定、开始积累真实数据、或者进入多人协作频繁改表阶段时，可以平滑升级为“版本化 migration”方案。

## AutoMigrate 触发位置

启动链路：

1. `cmd/server/main.go`
2. `internal/bootstrap.Build(...)`
3. 数据库初始化 `internal/bootstrap.NewDatabase(...)`
4. 自动迁移 `internal/bootstrap.RunMigrations(...)`

也就是说，服务启动时连接数据库成功后，会根据配置决定是否执行 `AutoMigrate`。

关键代码位置：

- 启动时触发迁移：[internal/bootstrap/app.go](/home/xina/projects/AI-For-Oj/internal/bootstrap/app.go)
- 自动迁移逻辑：[internal/bootstrap/migrate.go](/home/xina/projects/AI-For-Oj/internal/bootstrap/migrate.go)
- 模型注册入口：[internal/model/schema.go](/home/xina/projects/AI-For-Oj/internal/model/schema.go)

## AutoMigrate 配置开关

配置文件：

```yaml
database:
  auto_migrate: true
```

环境变量：

```bash
DB_AUTO_MIGRATE=true
```

当该开关为 `true` 时，应用启动后会自动执行建表/补字段等 GORM 自动迁移行为；为 `false` 时则跳过迁移。

## 本地启动

1. 启动 MySQL：

```bash
docker compose up -d mysql
```

2. 启动服务：

```bash
go run ./cmd/server
```

3. 检查健康状态：

```bash
curl http://127.0.0.1:8080/health
```

## 当前真实判题沙箱

当前 Judge 默认使用第一版 `DockerSandbox`，不再默认走 `MockSandbox`。

设计边界：

- `judge` 仍然只依赖 `sandbox` 接口，不感知 Docker 细节
- 默认目标镜像是 `gcc:13`
- 当前仅支持 `cpp17`
- 当前仅支持单文件源码、标准输入输出题

当前版本的安全边界：

- 使用短生命周期容器，执行完成后自动清理
- 使用隔离工作目录保存源码和编译产物
- 运行时使用超时控制、基础 CPU / memory / pids limit、`--network none`
- 还没有做到更严格的 syscall / namespace / seccomp 级别加固

后续可以平滑升级到更严格的 `NsJailSandbox` 或更强隔离方案，而不需要推翻 `judge/service/handler` 主链。

## DockerSandbox 运行前提

需要满足：

1. 本机 Docker daemon 可用
2. 已拉取 `gcc:13`
3. 如果通过 `docker compose` 启动应用，`app` 容器需要能访问 `/var/run/docker.sock`

如果 `gcc:13` 尚未拉取成功，Judge 会返回清晰错误，说明这是运行环境问题，而不是判题逻辑问题。

## 真实判题手动验证步骤

1. 拉取编译镜像：

```bash
docker pull gcc:13
```

2. 启动依赖：

```bash
docker compose up -d mysql
```

3. 启动服务：

```bash
go run ./cmd/server
```

4. 创建题目：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/problems \
  -H 'Content-Type: application/json' \
  -d '{
    "title":"Echo",
    "description":"读取一行并原样输出",
    "input_spec":"输入一行字符串",
    "output_spec":"输出同样的字符串",
    "samples":"[{\"input\":\"hello\",\"output\":\"hello\"}]",
    "time_limit_ms":1000,
    "memory_limit_mb":256,
    "difficulty":"easy",
    "tags":"implementation"
  }'
```

5. 添加测试点：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/problems/1/testcases \
  -H 'Content-Type: application/json' \
  -d '{
    "input":"hello\n",
    "expected_output":"hello\n",
    "is_sample":true
  }'
```

6. 提交正确代码，验证 `AC`：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/submissions/judge \
  -H 'Content-Type: application/json' \
  -d '{
    "problem_id":1,
    "language":"cpp17",
    "source_code":"#include <bits/stdc++.h>\nusing namespace std;\nint main(){ios::sync_with_stdio(false);cin.tie(nullptr);string s;getline(cin,s);cout<<s<<\"\\n\";return 0;}"
  }'
```

7. 提交错误代码，验证 `WA`：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/submissions/judge \
  -H 'Content-Type: application/json' \
  -d '{
    "problem_id":1,
    "language":"cpp17",
    "source_code":"#include <bits/stdc++.h>\nusing namespace std;\nint main(){cout<<\"wrong\\n\";return 0;}"
  }'
```

8. 提交编译失败代码，验证 `CE`：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/submissions/judge \
  -H 'Content-Type: application/json' \
  -d '{
    "problem_id":1,
    "language":"cpp17",
    "source_code":"#include <bits/stdc++.h>\nint main( { return 0; }"
  }'
```

9. 提交运行时错误代码，验证 `RE`：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/submissions/judge \
  -H 'Content-Type: application/json' \
  -d '{
    "problem_id":1,
    "language":"cpp17",
    "source_code":"#include <bits/stdc++.h>\nusing namespace std;\nint main(){int x=0;cout<<(1/x)<<\"\\n\";return 0;}"
  }'
```

10. 提交死循环代码，验证 `TLE`：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/submissions/judge \
  -H 'Content-Type: application/json' \
  -d '{
    "problem_id":1,
    "language":"cpp17",
    "source_code":"#include <bits/stdc++.h>\nusing namespace std;\nint main(){while(true){} return 0;}"
  }'
```
