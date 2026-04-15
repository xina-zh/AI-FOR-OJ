#!/usr/bin/env python3
import argparse
import json
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Tuple


SECTION_NAMES = [
    "Problem Description",
    "Input",
    "Output",
    "Sample Input",
    "Sample Output",
    "Time Limit",
    "Memory Limit",
]

DEFAULT_BASE_URL = "http://127.0.0.1:8080"
DEFAULT_TIME_LIMIT_MS = 1000
DEFAULT_MEMORY_LIMIT_MB = 256
DEFAULT_DIFFICULTY = "unknown"
DEFAULT_TAGS = ""


class ImportErrorDetail(Exception):
    pass


@dataclass
class ParsedProblem:
    title: str
    description: str
    input_spec: str
    output_spec: str
    samples: str
    time_limit_ms: int
    memory_limit_mb: int
    difficulty: str
    tags: str
    sample_input: str
    sample_output: str


@dataclass
class ParsedTestCase:
    number: int
    input_text: str
    output_text: str
    is_sample: bool


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Import local problems via AI-For-Oj HTTP API")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL, help="API base URL, default: %(default)s")
    parser.add_argument("--dir", required=True, help="Root directory containing problem subdirectories")
    parser.add_argument("--dry-run", action="store_true", help="Validate files only, do not call HTTP API")
    parser.add_argument("--stop-on-error", action="store_true", help="Stop after the first problem failure")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    configure_proxy_behavior(args.base_url)
    root_dir = Path(args.dir).expanduser().resolve()
    if not root_dir.is_dir():
        print(f"[fatal] problem root directory not found: {root_dir}", file=sys.stderr)
        return 1

    problem_dirs = sorted(path for path in root_dir.iterdir() if path.is_dir())
    if not problem_dirs:
        print(f"[fatal] no problem subdirectories found under: {root_dir}", file=sys.stderr)
        return 1

    success_count = 0
    failure_count = 0
    created_problem_ids: List[Tuple[str, int]] = []
    failures: List[Tuple[str, str]] = []

    print(f"[info] scanning {len(problem_dirs)} problem directories under {root_dir}")
    if args.dry_run:
        print("[info] running in dry-run mode, HTTP API will not be called")

    for problem_dir in problem_dirs:
        try:
            parsed_problem = parse_problem_dir(problem_dir)
            if args.dry_run:
                print(
                    f"[dry-run] ok title={parsed_problem.title} "
                    f"time_limit_ms={parsed_problem.time_limit_ms} "
                    f"memory_limit_mb={parsed_problem.memory_limit_mb}"
                )
            else:
                problem_id = create_problem(args.base_url, parsed_problem)
                testcases = load_testcases(problem_dir, parsed_problem.sample_input, parsed_problem.sample_output)
                for testcase in testcases:
                    create_testcase(args.base_url, problem_id, testcase)
                created_problem_ids.append((problem_dir.name, problem_id))
                print(
                    f"[ok] imported {problem_dir.name} -> problem_id={problem_id} "
                    f"testcases={len(testcases)}"
                )
            success_count += 1
        except ImportErrorDetail as exc:
            failure_count += 1
            failures.append((problem_dir.name, str(exc)))
            print(f"[error] {problem_dir.name}: {exc}", file=sys.stderr)
            if args.stop_on_error:
                break
        except Exception as exc:  # pragma: no cover - defensive fallback
            failure_count += 1
            failures.append((problem_dir.name, f"unexpected error: {exc}"))
            print(f"[error] {problem_dir.name}: unexpected error: {exc}", file=sys.stderr)
            if args.stop_on_error:
                break

    print("")
    print("Import Summary")
    print(f"- success_count: {success_count}")
    print(f"- failure_count: {failure_count}")
    if created_problem_ids:
        print("- created_problem_ids:")
        for name, problem_id in created_problem_ids:
            print(f"  - {name}: {problem_id}")
    if failures:
        print("- failures:")
        for name, reason in failures:
            print(f"  - {name}: {reason}")

    return 0 if failure_count == 0 else 1


def configure_proxy_behavior(base_url: str) -> None:
    parsed = urllib.parse.urlparse(base_url)
    host = (parsed.hostname or "").strip().lower()
    if host in {"127.0.0.1", "localhost", "::1"}:
        urllib.request.install_opener(urllib.request.build_opener(urllib.request.ProxyHandler({})))
        print(f"[info] bypassing system proxy for local base-url: {base_url}")


def parse_problem_dir(problem_dir: Path) -> ParsedProblem:
    statement_path = problem_dir / "statement.txt"
    if not statement_path.is_file():
        raise ImportErrorDetail("missing statement.txt")

    statement_text = read_text_file(statement_path)
    sections = parse_statement_sections(statement_text)

    missing_sections = [name for name in SECTION_NAMES[:5] if not sections.get(name, "").strip()]
    if missing_sections:
        raise ImportErrorDetail(f"missing required statement sections: {', '.join(missing_sections)}")

    time_limit_ms = parse_int_or_default(
        sections.get("Time Limit", ""),
        DEFAULT_TIME_LIMIT_MS,
        "Time Limit",
    )
    memory_limit_mb = parse_int_or_default(
        sections.get("Memory Limit", ""),
        DEFAULT_MEMORY_LIMIT_MB,
        "Memory Limit",
    )

    sample_input = normalize_text_block(sections["Sample Input"])
    sample_output = normalize_text_block(sections["Sample Output"])
    samples = json.dumps([{"input": sample_input, "output": sample_output}], ensure_ascii=False)

    testcases = load_testcases(problem_dir, sample_input, sample_output)
    if not testcases:
        raise ImportErrorDetail("no testcase pairs found (need N.in and N.out)")

    return ParsedProblem(
        title=problem_dir.name,
        description=normalize_text_block(sections["Problem Description"]),
        input_spec=normalize_text_block(sections["Input"]),
        output_spec=normalize_text_block(sections["Output"]),
        samples=samples,
        time_limit_ms=time_limit_ms,
        memory_limit_mb=memory_limit_mb,
        difficulty=DEFAULT_DIFFICULTY,
        tags=DEFAULT_TAGS,
        sample_input=sample_input,
        sample_output=sample_output,
    )


def parse_statement_sections(statement_text: str) -> Dict[str, str]:
    normalized = statement_text.replace("\r\n", "\n").replace("\r", "\n")
    lines = normalized.split("\n")
    sections: Dict[str, List[str]] = {}
    current_name = None

    section_headers = {}
    for name in SECTION_NAMES:
        section_headers[name] = name
        section_headers[f"{name}:"] = name

    for line in lines:
        stripped = line.strip()
        if stripped in section_headers:
            current_name = section_headers[stripped]
            sections[current_name] = []
            continue
        if current_name is not None:
            sections[current_name].append(line)

    return {name: "\n".join(content).strip("\n") for name, content in sections.items()}


def parse_int_or_default(raw: str, default_value: int, field_name: str) -> int:
    value = raw.strip()
    if not value:
        return default_value

    match = re.search(r"-?\d+", value)
    if not match:
        raise ImportErrorDetail(f"failed to parse {field_name}: {value!r}")

    try:
        return int(match.group(0))
    except ValueError as exc:  # pragma: no cover - defensive fallback
        raise ImportErrorDetail(f"failed to parse {field_name}: {value!r}") from exc


def load_testcases(problem_dir: Path, sample_input: str, sample_output: str) -> List[ParsedTestCase]:
    numbered_inputs: Dict[int, Path] = {}
    numbered_outputs: Dict[int, Path] = {}

    for path in problem_dir.iterdir():
        if not path.is_file():
            continue
        if path.name == "statement.txt":
            continue
        input_match = re.fullmatch(r"(\d+)\.in", path.name)
        if input_match:
            numbered_inputs[int(input_match.group(1))] = path
            continue
        output_match = re.fullmatch(r"(\d+)\.out", path.name)
        if output_match:
            numbered_outputs[int(output_match.group(1))] = path

    input_numbers = set(numbered_inputs)
    output_numbers = set(numbered_outputs)
    if input_numbers != output_numbers:
        missing_inputs = sorted(output_numbers - input_numbers)
        missing_outputs = sorted(input_numbers - output_numbers)
        parts = []
        if missing_inputs:
            parts.append("missing .in for: " + ", ".join(str(value) for value in missing_inputs))
        if missing_outputs:
            parts.append("missing .out for: " + ", ".join(str(value) for value in missing_outputs))
        raise ImportErrorDetail("; ".join(parts))

    testcases: List[ParsedTestCase] = []
    for number in sorted(input_numbers):
        input_text = read_text_file(numbered_inputs[number])
        output_text = read_text_file(numbered_outputs[number])
        testcases.append(
            ParsedTestCase(
                number=number,
                input_text=input_text,
                output_text=output_text,
                is_sample=(input_text == sample_input and output_text == sample_output),
            )
        )
    return testcases


def create_problem(base_url: str, problem: ParsedProblem) -> int:
    payload = {
        "title": problem.title,
        "description": problem.description,
        "input_spec": problem.input_spec,
        "output_spec": problem.output_spec,
        "samples": problem.samples,
        "time_limit_ms": problem.time_limit_ms,
        "memory_limit_mb": problem.memory_limit_mb,
        "difficulty": problem.difficulty,
        "tags": problem.tags,
    }
    response = post_json(f"{base_url.rstrip('/')}/api/v1/problems", payload)
    if "id" not in response:
        raise ImportErrorDetail(f"create problem succeeded but response has no id: {response}")
    return int(response["id"])


def create_testcase(base_url: str, problem_id: int, testcase: ParsedTestCase) -> None:
    payload = {
        "input": testcase.input_text,
        "expected_output": testcase.output_text,
        "is_sample": testcase.is_sample,
    }
    post_json(f"{base_url.rstrip('/')}/api/v1/problems/{problem_id}/testcases", payload)


def post_json(url: str, payload: Dict) -> Dict:
    request = urllib.request.Request(
        url=url,
        data=json.dumps(payload, ensure_ascii=False).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(request) as response:
            body = response.read().decode("utf-8")
    except urllib.error.HTTPError as exc:
        error_body = exc.read().decode("utf-8", errors="replace")
        raise ImportErrorDetail(f"HTTP {exc.code} {url}: {error_body}") from exc
    except urllib.error.URLError as exc:
        raise ImportErrorDetail(f"request failed {url}: {exc}") from exc

    try:
        return json.loads(body)
    except json.JSONDecodeError as exc:
        raise ImportErrorDetail(f"invalid JSON response from {url}: {body}") from exc


def read_text_file(path: Path) -> str:
    try:
        return path.read_text(encoding="utf-8")
    except FileNotFoundError as exc:
        raise ImportErrorDetail(f"missing file: {path.name}") from exc
    except UnicodeDecodeError as exc:
        raise ImportErrorDetail(f"file is not valid utf-8: {path.name}") from exc


def normalize_text_block(text: str) -> str:
    normalized = text.replace("\r\n", "\n").replace("\r", "\n").strip("\n")
    return normalized


if __name__ == "__main__":
    sys.exit(main())
