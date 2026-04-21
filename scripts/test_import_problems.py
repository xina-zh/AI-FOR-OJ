import tempfile
import unittest
from pathlib import Path

from scripts.import_problems import parse_problem_dir


class ImportProblemsLimitParsingTest(unittest.TestCase):
    def test_parses_time_and_memory_units_case_insensitively(self):
        parsed = self.parse_with_limits("1S", "65536K")

        self.assertEqual(parsed.time_limit_ms, 1000)
        self.assertEqual(parsed.memory_limit_mb, 64)

    def test_parses_units_with_spaces(self):
        parsed = self.parse_with_limits("1000 MS", "128 mb")

        self.assertEqual(parsed.time_limit_ms, 1000)
        self.assertEqual(parsed.memory_limit_mb, 128)

    def test_keeps_existing_default_units_when_unit_is_missing(self):
        parsed = self.parse_with_limits("1500", "256")

        self.assertEqual(parsed.time_limit_ms, 1500)
        self.assertEqual(parsed.memory_limit_mb, 256)

    def parse_with_limits(self, time_limit: str, memory_limit: str):
        with tempfile.TemporaryDirectory() as temp_dir:
            problem_dir = Path(temp_dir) / "A Plus B"
            problem_dir.mkdir()
            (problem_dir / "statement.txt").write_text(
                "\n".join(
                    [
                        "Problem Description",
                        "Add two numbers.",
                        "Input",
                        "Two integers.",
                        "Output",
                        "Their sum.",
                        "Sample Input",
                        "1 2",
                        "Sample Output",
                        "3",
                        "Time Limit",
                        time_limit,
                        "Memory Limit",
                        memory_limit,
                    ]
                ),
                encoding="utf-8",
            )
            (problem_dir / "1.in").write_text("1 2", encoding="utf-8")
            (problem_dir / "1.out").write_text("3", encoding="utf-8")

            return parse_problem_dir(problem_dir)


if __name__ == "__main__":
    unittest.main()
