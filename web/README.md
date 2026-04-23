# AI-For-Oj Web

React 前端实验控制台，覆盖本地实验执行、历史回看、token 成本和 trace 回放。

## 技术栈

- React
- Vite
- TypeScript
- TanStack Query
- React Router
- Vitest
- Playwright

## 本地开发

先启动后端：

```bash
docker compose up -d mysql
go run ./cmd/server
```

再启动前端：

```bash
cd web
npm install
npm run dev
```

默认地址：

- 前端：`http://127.0.0.1:5173`
- 后端：`http://127.0.0.1:8080`

Vite 开发服务器会把 `/api` 和 `/health` 代理到后端。

## 常用命令

```bash
npm run build
npm test -- --run
npm run e2e
```

`npm run e2e` 使用 Vite preview，运行前需要先有可用的 `dist`。通常先执行：

```bash
npm run build
npm run e2e
```

## 页面

- `/` Dashboard：后端健康检查、题目数、submission 数、最近运行
- `/problems` 题目列表、详情、创建和 testcase 管理
- `/solve` 单题 AI solve
- `/ai-runs/:id` AI solve run 详情
- `/experiments` 批量实验运行与历史
- `/experiments/:id` experiment 详情
- `/compare` baseline / candidate 对比实验
- `/compare/:id` compare 详情
- `/repeat` repeat 稳定性实验
- `/repeat/:id` repeat 详情
- `/tokens` token 成本分析
- `/trace/experiment-runs/:id` experiment run trace 回放
- `/submissions` submission 列表
- `/submissions/:id` submission 详情
