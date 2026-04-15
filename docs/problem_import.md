# Problem Import

## Directory Format

```text
problems/
  shortest-path/
    statement.txt
    1.in
    1.out
    2.in
    2.out
  tree-dp/
    statement.txt
    1.in
    1.out
```

## `statement.txt` Format

```text
Problem Description:
给定一个 n 个点 m 条边的有向图，求 1 到 n 的最短路。

Input:
第一行两个整数 n, m。

Output:
输出 1 到 n 的最短路长度。

Sample Input:
3 3
1 2 5
2 3 7
1 3 20

Sample Output:
12

Time Limit:
1000

Memory Limit:
256
```

支持字段：

- `Problem Description`
- `Input`
- `Output`
- `Sample Input`
- `Sample Output`
- `Time Limit`
- `Memory Limit`

默认值：

- `title`: 使用题目目录名
- `time_limit_ms`: `1000`
- `memory_limit_mb`: `256`
- `difficulty`: `unknown`
- `tags`: 空字符串

## Usage

Dry run:

```bash
python3 scripts/import_problems.py --dir ./problems --dry-run
```

Real import:

```bash
python3 scripts/import_problems.py --dir ./problems
```

Custom base URL:

```bash
python3 scripts/import_problems.py --base-url http://127.0.0.1:8080 --dir ./problems
```

脚本会调用现有接口：

- `POST /api/v1/problems`
- `POST /api/v1/problems/:id/testcases`
