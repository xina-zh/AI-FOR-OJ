# AI-For-OJ Frontend Port Startup Implementation Plan

> **For Codex:** Use `executing-plans` to implement this plan task-by-task. Use `subagent-driven-development` only when the user explicitly authorizes subagents, delegation, or parallel agents.

**Goal:** Make AI-For-OJ start its own frontend on port `5188` whenever the AI-For-OJ Docker dev stack is started, without conflicting with XCPC-Training-Agent on port `5173`.

**Architecture:** Keep the AI-For-OJ backend on `8080` because the backend has no conflict with XCPC-Training-Agent and existing API docs/tests assume `8080`. Add a real Vite React frontend under `web/`, configure it to listen on `5188`, and add a `frontend` service to `docker-compose.yml` so `docker compose up -d` starts MySQL, backend, and frontend together. Do not modify the XCPC-Training-Agent repository; if its old containers should stop auto-starting, handle that as a Docker runtime operation outside this repo.

**Tech Stack:** Docker Compose, Go/Gin backend, MySQL 8.4, Vite, React, TypeScript, Vitest.

---

## Current State

- `AI-For-OJ/docker-compose.yml` currently defines only `app` and `mysql`.
- `AI-For-OJ/web/` currently has no runnable frontend files.
- `AI-For-OJ` backend currently listens on host port `8080`.
- `XCPC-Training-Agent` currently uses host port `5173` for its frontend.
- The requested non-conflicting AI-For-OJ frontend port is `5188`.

## Design Decisions

- Use `5188` for the AI-For-OJ frontend dev server.
- Keep the AI-For-OJ backend on `8080`.
- Configure Vite with `strictPort: true` so port conflicts fail loudly instead of silently moving to another port.
- Configure the frontend dev proxy so browser requests to `/api` and `/health` go through the frontend origin.
- Add `frontend` to the default Docker Compose stack. Running `docker compose up -d` from the AI-For-OJ repo should start all three services.
- Do not set `restart: always` for the frontend. Use `restart: unless-stopped` to match the backend and avoid surprising auto-restarts after the user explicitly stops the stack.
- Do not edit `/home/xina/projects/XCPC-Training-Agent`. Port separation is enough for no conflict. Disabling XCPC auto-start should be an explicit runtime command.

## Task 1: Baseline Verification

**Files:**
- Read: `docker-compose.yml`
- Read: `web/`
- Read: `/home/xina/projects/XCPC-Training-Agent/docker-compose.yml`

**Step 1: Verify current AI-For-OJ services**

Run:

```bash
docker compose -p ai-for-oj ps
```

Expected:

- `ai-for-oj-app` is present.
- `ai-for-oj-mysql` is present.
- No `frontend` service is present.

**Step 2: Verify current web directory state**

Run:

```bash
find web -maxdepth 4 -type f -print | sort
```

Expected before implementation:

- No frontend project files such as `web/package.json`, `web/vite.config.ts`, or `web/index.html`.

**Step 3: Verify XCPC-Training-Agent uses port 5173**

Run:

```bash
docker compose -p xcpc-training-agent -f /home/xina/projects/XCPC-Training-Agent/docker-compose.yml ps
```

Expected:

- If running, `xcpc-training-agent-frontend-1` maps `0.0.0.0:5173->5173/tcp`.
- This confirms AI-For-OJ should not use `5173`.

**Step 4: Commit**

No commit for this task. It is read-only.

---

## Task 2: Scaffold the AI-For-OJ Frontend on Port 5188

**Files:**
- Create: `web/package.json`
- Create: `web/index.html`
- Create: `web/tsconfig.json`
- Create: `web/tsconfig.node.json`
- Create: `web/vite.config.ts`
- Create: `web/src/main.tsx`
- Create: `web/src/App.tsx`
- Create: `web/src/App.test.tsx`
- Create: `web/src/setupTests.ts`
- Create: `web/src/styles/base.css`

**Step 1: Create `web/package.json`**

Create:

```json
{
  "name": "ai-for-oj-web",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite --host 0.0.0.0 --port 5188",
    "build": "tsc -b && vite build",
    "test": "vitest run"
  },
  "dependencies": {
    "@vitejs/plugin-react": "^5.0.0",
    "vite": "^7.0.0",
    "typescript": "^5.8.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  },
  "devDependencies": {
    "@testing-library/jest-dom": "^6.6.0",
    "@testing-library/react": "^16.0.0",
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "vitest": "^3.0.0"
  }
}
```

**Step 2: Create `web/vite.config.ts`**

Create:

```ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiProxyTarget = process.env.VITE_API_PROXY_TARGET ?? "http://127.0.0.1:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    port: 5188,
    strictPort: true,
    proxy: {
      "/api": {
        target: apiProxyTarget,
        changeOrigin: true
      },
      "/health": {
        target: apiProxyTarget,
        changeOrigin: true
      }
    }
  },
  test: {
    environment: "jsdom",
    setupFiles: "./src/setupTests.ts"
  }
});
```

**Step 3: Create TypeScript configs**

Create `web/tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "useDefineForClassFields": true,
    "lib": ["DOM", "DOM.Iterable", "ES2022"],
    "allowJs": false,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx"
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

Create `web/tsconfig.node.json`:

```json
{
  "compilerOptions": {
    "composite": true,
    "skipLibCheck": true,
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "allowSyntheticDefaultImports": true
  },
  "include": ["vite.config.ts"]
}
```

**Step 4: Create the minimal app shell**

Create `web/index.html`:

```html
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>AI-For-OJ</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

Create `web/src/main.tsx`:

```tsx
import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";
import "./styles/base.css";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

Create `web/src/App.tsx`:

```tsx
export function App() {
  return (
    <main className="app-shell">
      <section className="hero">
        <p className="eyebrow">AI-For-OJ</p>
        <h1>Experiment Console</h1>
        <p className="summary">
          Backend proxy is configured. Use this shell as the base for problem,
          solve, experiment, compare, repeat, and submission workflows.
        </p>
      </section>
    </main>
  );
}
```

Create `web/src/styles/base.css`:

```css
:root {
  color: #172026;
  background: #f7f9fb;
  font-family:
    Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  min-width: 320px;
}

.app-shell {
  min-height: 100vh;
  padding: 48px;
}

.hero {
  max-width: 760px;
}

.eyebrow {
  margin: 0 0 12px;
  color: #0f766e;
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0;
  text-transform: uppercase;
}

h1 {
  margin: 0;
  font-size: 44px;
  line-height: 1.1;
  letter-spacing: 0;
}

.summary {
  margin: 20px 0 0;
  color: #46545f;
  font-size: 18px;
  line-height: 1.6;
}

@media (max-width: 640px) {
  .app-shell {
    padding: 28px;
  }

  h1 {
    font-size: 34px;
  }
}
```

**Step 5: Create the first frontend test**

Create `web/src/setupTests.ts`:

```ts
import "@testing-library/jest-dom/vitest";
```

Create `web/src/App.test.tsx`:

```tsx
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { App } from "./App";

describe("App", () => {
  it("renders the AI-For-OJ shell", () => {
    render(<App />);

    expect(screen.getByText("AI-For-OJ")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Experiment Console" })).toBeInTheDocument();
  });
});
```

**Step 6: Install dependencies**

Run:

```bash
cd web && npm install
```

Expected:

- `web/package-lock.json` is created.
- Command exits `0`.

**Step 7: Run frontend tests**

Run:

```bash
cd web && npm run test
```

Expected:

- Vitest exits `0`.
- The `App` test passes.

**Step 8: Run frontend build**

Run:

```bash
cd web && npm run build
```

Expected:

- TypeScript and Vite build exit `0`.
- `web/dist/` is generated.

**Step 9: Commit**

Run:

```bash
git add web
git commit -m "feat: scaffold ai-for-oj frontend"
```

Expected:

- Commit succeeds.

---

## Task 3: Add the Frontend Service to AI-For-OJ Docker Compose

**Files:**
- Modify: `docker-compose.yml`

**Step 1: Add a `frontend` service**

Modify `docker-compose.yml` to include:

```yaml
  frontend:
    image: node:22-alpine
    container_name: ai-for-oj-frontend
    working_dir: /app/web
    environment:
      VITE_API_PROXY_TARGET: http://app:8080
    command: sh -c "npm install && npm run dev"
    volumes:
      - ./web:/app/web
      - frontend_node_modules:/app/web/node_modules
    ports:
      - "5188:5188"
    depends_on:
      - app
    restart: unless-stopped
```

Add the new named volume under `volumes`:

```yaml
volumes:
  mysql_data:
  frontend_node_modules:
```

The final compose file should still keep:

- `app` on `8080:8080`
- `mysql` on `3306:3306`
- `frontend` on `5188:5188`

**Step 2: Validate the compose file**

Run:

```bash
docker compose config
```

Expected:

- Command exits `0`.
- Output contains `ai-for-oj-frontend`.
- Output contains `5188:5188`.

**Step 3: Start the full AI-For-OJ stack**

Run:

```bash
docker compose up -d --build
```

Expected:

- `ai-for-oj-mysql` starts and becomes healthy.
- `ai-for-oj-app` starts.
- `ai-for-oj-frontend` starts.

**Step 4: Verify services**

Run:

```bash
docker compose ps
```

Expected:

- `ai-for-oj-app` is `Up`.
- `ai-for-oj-mysql` is `Up` and `healthy`.
- `ai-for-oj-frontend` is `Up`.
- `ai-for-oj-frontend` maps `0.0.0.0:5188->5188/tcp`.

**Step 5: Verify backend health**

Run:

```bash
curl --noproxy '*' -sS http://127.0.0.1:8080/health
```

Expected:

```json
{
  "name": "ai-for-oj",
  "env": "docker",
  "status": "ok",
  "database": "up"
}
```

The exact timestamp can differ.

**Step 6: Verify frontend HTTP response**

Run:

```bash
curl --noproxy '*' -sS -I http://127.0.0.1:5188
```

Expected:

- HTTP status is `200 OK`.
- Response comes from Vite.

**Step 7: Verify frontend proxy to backend**

Run:

```bash
curl --noproxy '*' -sS http://127.0.0.1:5188/api/v1/meta/experiment-options
```

Expected:

- JSON contains `default_model`.
- JSON contains `prompts`.
- JSON contains `agents`.

**Step 8: Commit**

Run:

```bash
git add docker-compose.yml
git commit -m "chore: start frontend with ai-for-oj compose stack"
```

Expected:

- Commit succeeds.

---

## Task 4: Add a Canonical Dev Startup Script

**Files:**
- Create: `scripts/dev_up.sh`

**Why:** `docker compose up -d` starts all default services, but old muscle memory like `docker compose up -d mysql app` will not start `frontend`. A script gives one command that always starts the intended AI-For-OJ development stack.

**Step 1: Create `scripts/dev_up.sh`**

Create:

```sh
#!/usr/bin/env sh
set -eu

docker compose up -d --build mysql app frontend

printf '%s\n' 'AI-For-OJ is starting:'
printf '%s\n' '  Backend:  http://127.0.0.1:8080'
printf '%s\n' '  Frontend: http://127.0.0.1:5188'
```

Do not add browser auto-open logic in this script yet. Opening browsers from automation is brittle across Linux desktop, WSL, SSH, CI, and containerized environments. If browser auto-open is still desired after the stack behavior is stable, add it as a separate explicit task.

**Step 2: Make the script executable**

Run:

```bash
chmod +x scripts/dev_up.sh
```

Expected:

- `scripts/dev_up.sh` is executable.

**Step 3: Run the script**

Run:

```bash
./scripts/dev_up.sh
```

Expected:

- Docker Compose starts `mysql`, `app`, and `frontend`.
- Output prints backend and frontend URLs.

**Step 4: Verify**

Run:

```bash
docker compose ps
curl --noproxy '*' -sS http://127.0.0.1:8080/health
curl --noproxy '*' -sS -I http://127.0.0.1:5188
```

Expected:

- Compose shows all three AI-For-OJ services.
- Backend health returns `status: ok`.
- Frontend returns `200 OK`.

**Step 5: Commit**

Run:

```bash
git add scripts/dev_up.sh
git commit -m "chore: add ai-for-oj dev startup script"
```

Expected:

- Commit succeeds.

---

## Task 5: Document the New Startup Flow

**Files:**
- Modify: `README.md`

**Step 1: Update local startup docs**

Replace the old local Docker startup guidance with:

```markdown
## 本地启动

### Docker Compose 开发栈

启动 AI-For-OJ 后端、MySQL 和前端：

```bash
./scripts/dev_up.sh
```

也可以直接运行：

```bash
docker compose up -d --build
```

启动后访问：

- 后端健康检查：`http://127.0.0.1:8080/health`
- 前端控制台：`http://127.0.0.1:5188`

AI-For-OJ 前端固定使用 `5188`，避免和 XCPC-Training-Agent 的 `5173` 冲突。
```

Keep the existing `go run ./cmd/server` instructions only as an advanced backend-only path. State clearly that backend-only startup will not start the frontend.

**Step 2: Add a note about XCPC-Training-Agent**

Add:

```markdown
### 与 XCPC-Training-Agent 的端口关系

XCPC-Training-Agent 的前端通常使用 `5173`。AI-For-OJ 的前端固定使用 `5188`，两个项目可以同时运行。

如果不希望旧的 XCPC-Training-Agent 容器自动恢复，需要在 Docker 运行时关闭它们的 restart policy；这不是 AI-For-OJ 仓库配置的一部分。
```

**Step 3: Run markdown sanity check**

Run:

```bash
rg -n "5188|5173|dev_up|docker compose up" README.md
```

Expected:

- Output includes the new frontend port `5188`.
- Output includes the conflict note for `5173`.
- Output includes `./scripts/dev_up.sh`.

**Step 4: Commit**

Run:

```bash
git add README.md
git commit -m "docs: document ai-for-oj frontend startup"
```

Expected:

- Commit succeeds.

---

## Task 6: Optional Runtime Cleanup for XCPC-Training-Agent Auto-Restart

**Files:**
- No repository files.
- Runtime-only Docker operation.

**Why:** This stops old XCPC-Training-Agent containers from restarting automatically. It does not modify either repository.

**Step 1: Inspect current XCPC restart policy**

Run:

```bash
docker inspect xcpc-training-agent-frontend-1 xcpc-training-agent-app-1 \
  --format '{{.Name}} restart={{.HostConfig.RestartPolicy.Name}} status={{.State.Status}}'
```

Expected before cleanup:

- `restart=always` for the XCPC containers if they still have the old policy.

**Step 2: Disable XCPC auto-restart**

Run only if the user explicitly wants the runtime cleanup:

```bash
docker update --restart=no xcpc-training-agent-frontend-1 xcpc-training-agent-app-1
```

Expected:

- Docker exits `0`.

**Step 3: Stop XCPC containers**

Run only if the user explicitly wants them stopped:

```bash
docker compose -p xcpc-training-agent -f /home/xina/projects/XCPC-Training-Agent/docker-compose.yml stop frontend app
```

Expected:

- XCPC frontend and app stop.
- Port `5173` is no longer occupied by XCPC.

**Step 4: Verify**

Run:

```bash
docker compose -p xcpc-training-agent -f /home/xina/projects/XCPC-Training-Agent/docker-compose.yml ps
docker inspect xcpc-training-agent-frontend-1 xcpc-training-agent-app-1 \
  --format '{{.Name}} restart={{.HostConfig.RestartPolicy.Name}} status={{.State.Status}}'
```

Expected:

- `restart=no`.
- `status=exited` for stopped XCPC containers.

**Step 5: Commit**

No commit. This task changes Docker runtime state only.

---

## Task 7: Final Verification

**Files:**
- Read: `docker-compose.yml`
- Read: `web/package.json`
- Read: `README.md`

**Step 1: Run frontend checks**

Run:

```bash
cd web && npm run test
cd web && npm run build
```

Expected:

- Both commands exit `0`.

**Step 2: Run backend checks**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./...
```

Expected:

- All Go tests pass.

**Step 3: Run compose checks**

Run:

```bash
docker compose config
docker compose up -d --build
docker compose ps
```

Expected:

- Compose config is valid.
- AI-For-OJ stack has `mysql`, `app`, and `frontend`.
- Frontend maps `5188`.
- Backend maps `8080`.

**Step 4: Run HTTP smoke tests**

Run:

```bash
curl --noproxy '*' -sS http://127.0.0.1:8080/health
curl --noproxy '*' -sS -I http://127.0.0.1:5188
curl --noproxy '*' -sS http://127.0.0.1:5188/api/v1/meta/experiment-options
```

Expected:

- Backend health returns `status: ok`.
- Frontend returns `200 OK`.
- Frontend proxy returns experiment options JSON.

**Step 5: Verify no AI-For-OJ service uses 5173**

Run:

```bash
docker compose ps
```

Expected:

- No AI-For-OJ service maps `5173`.
- AI-For-OJ frontend maps `5188`.

**Step 6: Commit any final doc or test fixes**

If any verification-driven doc or test fix is needed:

```bash
git add README.md web docker-compose.yml scripts/dev_up.sh
git commit -m "test: verify ai-for-oj frontend startup"
```

Expected:

- Commit succeeds if there were changes.
- If there were no changes, skip this commit.

---

## Acceptance Criteria

- `docker compose up -d --build` in `/home/xina/projects/AI-For-Oj` starts:
  - `ai-for-oj-mysql`
  - `ai-for-oj-app`
  - `ai-for-oj-frontend`
- `http://127.0.0.1:8080/health` returns backend health JSON.
- `http://127.0.0.1:5188` returns the AI-For-OJ frontend.
- `http://127.0.0.1:5188/api/v1/meta/experiment-options` proxies to the backend and returns JSON.
- AI-For-OJ does not use port `5173`.
- The plan does not require any modification to `/home/xina/projects/XCPC-Training-Agent`.
- If the optional XCPC runtime cleanup is executed, XCPC containers no longer have `restart=always`.

## Rollback

If the frontend service causes problems:

```bash
docker compose stop frontend
```

If the compose change itself must be reverted:

```bash
git revert <commit-that-added-frontend-service>
```

If the scaffolded frontend must be removed:

```bash
git revert <commit-that-scaffolded-frontend>
```

## Execution Notes

- Use @executing-plans for implementation.
- Use @verification-before-completion before claiming the stack works.
- Use @systematic-debugging if port binding, Docker networking, or Vite proxy behavior fails.
- Use @senior-frontend when expanding the shell into the full experiment console.

