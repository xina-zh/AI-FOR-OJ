package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

func TestAISolveServiceGetRunPreloadsAttempts(t *testing.T) {
	db := openAISolveRunGetRunTestDB(t)
	runRepo := repository.NewAISolveRunRepository(db)
	service := NewAISolveService(nil, runRepo, nil, nil, "default-model")

	run, err := service.GetRun(context.Background(), 17)
	if err != nil {
		t.Fatalf("get run returned error: %v", err)
	}

	if run.AttemptCount != 3 {
		t.Fatalf("expected summary attempt count to be loaded, got %+v", run)
	}
	if run.FailureType != "time_limit" {
		t.Fatalf("expected summary failure type to be loaded, got %+v", run)
	}
	if run.StrategyPath != "analysis,repair" {
		t.Fatalf("expected summary strategy path to be loaded, got %+v", run)
	}
	if len(run.Attempts) != 2 {
		t.Fatalf("expected attempts to be preloaded, got %+v", run.Attempts)
	}
	if run.Attempts[0].Stage != "analysis" || run.Attempts[1].Stage != "repair" {
		t.Fatalf("expected attempts ordered by attempt no, got %+v", run.Attempts)
	}
	if run.Attempts[0].JudgeVerdict != "WA" || run.Attempts[1].JudgeVerdict != "AC" {
		t.Fatalf("expected attempt verdicts to be loaded, got %+v", run.Attempts)
	}
	if run.Attempts[0].FailureType != "wrong_answer" || run.Attempts[1].RepairReason != "tighten edge cases" {
		t.Fatalf("expected attempt metadata to be loaded, got %+v", run.Attempts)
	}
}

func openAISolveRunGetRunTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	store := &aiSolveRunTestStore{
		run: model.AISolveRun{
			BaseModel:    model.BaseModel{ID: 17},
			ProblemID:    9,
			Model:        "mock-cpp17",
			PromptName:   "default",
			AgentName:    "direct_codegen_v1",
			AttemptCount: 3,
			FailureType:  "time_limit",
			StrategyPath: "analysis,repair",
			Status:       model.AISolveRunStatusSuccess,
		},
		attempts: []model.AISolveAttempt{
			{
				BaseModel:       model.BaseModel{ID: 201},
				AISolveRunID:    17,
				AttemptNo:       2,
				Stage:           "repair",
				FailureType:     "time_limit",
				RepairReason:    "tighten edge cases",
				JudgeVerdict:    "AC",
				StrategyPath:    "analysis,repair",
				JudgeTotalCount: 10,
			},
			{
				BaseModel:       model.BaseModel{ID: 200},
				AISolveRunID:    17,
				AttemptNo:       1,
				Stage:           "analysis",
				FailureType:     "wrong_answer",
				RepairReason:    "clarify the invariant",
				JudgeVerdict:    "WA",
				StrategyPath:    "analysis",
				JudgeTotalCount: 10,
			},
		},
	}

	driverName := fmt.Sprintf("ai-solve-run-repo-test-%d", time.Now().UnixNano())
	sql.Register(driverName, &aiSolveRunTestDriver{store: store})

	sqlDB, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm db: %v", err)
	}

	return db
}

type aiSolveRunTestStore struct {
	mu       sync.Mutex
	run      model.AISolveRun
	attempts []model.AISolveAttempt
}

type aiSolveRunTestDriver struct {
	store *aiSolveRunTestStore
}

func (d *aiSolveRunTestDriver) Open(name string) (driver.Conn, error) {
	return &aiSolveRunTestConn{store: d.store}, nil
}

type aiSolveRunTestConn struct {
	store *aiSolveRunTestStore
}

func (c *aiSolveRunTestConn) Prepare(query string) (driver.Stmt, error) {
	return &aiSolveRunTestStmt{conn: c, query: query}, nil
}

func (c *aiSolveRunTestConn) Close() error { return nil }

func (c *aiSolveRunTestConn) Begin() (driver.Tx, error) { return &aiSolveRunTestTx{}, nil }

func (c *aiSolveRunTestConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return &aiSolveRunTestTx{}, nil
}

func (c *aiSolveRunTestConn) Ping(context.Context) error { return nil }

func (c *aiSolveRunTestConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.exec(query, namedValuesToValues(args))
}

func (c *aiSolveRunTestConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.query(query, namedValuesToValues(args))
}

type aiSolveRunTestStmt struct {
	conn  *aiSolveRunTestConn
	query string
}

func (s *aiSolveRunTestStmt) Close() error { return nil }

func (s *aiSolveRunTestStmt) NumInput() int { return -1 }

func (s *aiSolveRunTestStmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.conn.exec(s.query, args)
}

func (s *aiSolveRunTestStmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.conn.query(s.query, args)
}

type aiSolveRunTestTx struct{}

func (t *aiSolveRunTestTx) Commit() error   { return nil }
func (t *aiSolveRunTestTx) Rollback() error { return nil }

func (c *aiSolveRunTestConn) exec(query string, args []driver.Value) (driver.Result, error) {
	normalized := normalizeSQL(query)

	switch {
	case strings.HasPrefix(normalized, "create table "):
		return aiSolveRunTestResult(0), nil
	default:
		return nil, fmt.Errorf("unexpected exec statement: %s", query)
	}
}

func (c *aiSolveRunTestConn) query(query string, args []driver.Value) (driver.Rows, error) {
	normalized := normalizeSQL(query)

	switch {
	case normalized == "select database()":
		return &aiSolveRunTestRows{
			columns: []string{"DATABASE()"},
			values:  [][]driver.Value{{"test"}},
		}, nil
	case strings.Contains(normalized, "from information_schema.schemata"):
		return &aiSolveRunTestRows{
			columns: []string{"SCHEMA_NAME"},
			values:  [][]driver.Value{{"test"}},
		}, nil
	case strings.Contains(normalized, "from information_schema.tables"):
		return &aiSolveRunTestRows{
			columns: []string{"TABLE_NAME"},
		}, nil
	case strings.Contains(normalized, "from `ai_solve_runs`"):
		if len(args) == 0 {
			return nil, fmt.Errorf("expected run id argument")
		}
		runID := asUint(args[0])
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		if c.store.run.ID != runID {
			return &aiSolveRunTestRows{columns: aiSolveRunColumns()}, nil
		}
		return newAISolveRunRows([]model.AISolveRun{c.store.run}), nil
	case strings.Contains(normalized, "from `ai_solve_attempts`"):
		if len(args) == 0 {
			return nil, fmt.Errorf("expected run id argument")
		}
		runID := asUint(args[0])
		c.store.mu.Lock()
		defer c.store.mu.Unlock()

		filtered := make([]model.AISolveAttempt, 0, len(c.store.attempts))
		for _, attempt := range c.store.attempts {
			if attempt.AISolveRunID == runID {
				filtered = append(filtered, attempt)
			}
		}
		sort.Slice(filtered, func(i, j int) bool {
			if filtered[i].AttemptNo == filtered[j].AttemptNo {
				return filtered[i].ID < filtered[j].ID
			}
			return filtered[i].AttemptNo < filtered[j].AttemptNo
		})
		return newAISolveAttemptRows(filtered), nil
	default:
		return nil, fmt.Errorf("unexpected query statement: %s", query)
	}
}

type aiSolveRunTestResult int64

func (r aiSolveRunTestResult) LastInsertId() (int64, error) { return int64(r), nil }

func (r aiSolveRunTestResult) RowsAffected() (int64, error) { return 1, nil }

type aiSolveRunTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func aiSolveRunColumns() []string {
	return []string{
		"id",
		"created_at",
		"updated_at",
		"problem_id",
		"model",
		"prompt_name",
		"agent_name",
		"attempt_count",
		"failure_type",
		"strategy_path",
		"prompt_preview",
		"raw_response",
		"extracted_code",
		"submission_id",
		"verdict",
		"status",
		"error_message",
		"token_input",
		"token_output",
		"llm_latency_ms",
		"total_latency_ms",
	}
}

func newAISolveRunRows(runs []model.AISolveRun) *aiSolveRunTestRows {
	rows := &aiSolveRunTestRows{
		columns: aiSolveRunColumns(),
		values:  make([][]driver.Value, 0, len(runs)),
	}
	for _, run := range runs {
		rows.values = append(rows.values, []driver.Value{
			int64(run.ID),
			run.CreatedAt,
			run.UpdatedAt,
			int64(run.ProblemID),
			run.Model,
			run.PromptName,
			run.AgentName,
			int64(run.AttemptCount),
			run.FailureType,
			run.StrategyPath,
			run.PromptPreview,
			run.RawResponse,
			run.ExtractedCode,
			run.SubmissionID,
			run.Verdict,
			run.Status,
			run.ErrorMessage,
			run.TokenInput,
			run.TokenOutput,
			int64(run.LLMLatencyMS),
			int64(run.TotalLatencyMS),
		})
	}
	return rows
}

func newAISolveAttemptRows(attempts []model.AISolveAttempt) *aiSolveRunTestRows {
	rows := &aiSolveRunTestRows{
		columns: []string{
			"id",
			"created_at",
			"updated_at",
			"ai_solve_run_id",
			"attempt_no",
			"stage",
			"failure_type",
			"repair_reason",
			"strategy_path",
			"prompt_preview",
			"raw_response",
			"extracted_code",
			"judge_verdict",
			"judge_passed_count",
			"judge_total_count",
			"timed_out",
			"compile_stderr",
			"run_stderr",
			"run_stdout",
			"error_message",
			"token_input",
			"token_output",
			"llm_latency_ms",
		},
		values: make([][]driver.Value, 0, len(attempts)),
	}
	for _, attempt := range attempts {
		rows.values = append(rows.values, []driver.Value{
			int64(attempt.ID),
			attempt.CreatedAt,
			attempt.UpdatedAt,
			int64(attempt.AISolveRunID),
			int64(attempt.AttemptNo),
			attempt.Stage,
			attempt.FailureType,
			attempt.RepairReason,
			attempt.StrategyPath,
			attempt.PromptPreview,
			attempt.RawResponse,
			attempt.ExtractedCode,
			attempt.JudgeVerdict,
			int64(attempt.JudgePassedCount),
			int64(attempt.JudgeTotalCount),
			attempt.TimedOut,
			attempt.CompileStderr,
			attempt.RunStderr,
			attempt.RunStdout,
			attempt.ErrorMessage,
			attempt.TokenInput,
			attempt.TokenOutput,
			int64(attempt.LLMLatencyMS),
		})
	}
	return rows
}

func (r *aiSolveRunTestRows) Columns() []string {
	return r.columns
}

func (r *aiSolveRunTestRows) Close() error {
	return nil
}

func (r *aiSolveRunTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func namedValuesToValues(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, 0, len(args))
	for _, arg := range args {
		values = append(values, arg.Value)
	}
	return values
}

func normalizeSQL(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(query)), " ")
}

func asUint(value driver.Value) uint {
	switch v := value.(type) {
	case int64:
		return uint(v)
	case int:
		return uint(v)
	case uint64:
		return uint(v)
	case uint:
		return v
	case []byte:
		var parsed uint
		_, _ = fmt.Sscanf(string(v), "%d", &parsed)
		return parsed
	default:
		var parsed uint
		_, _ = fmt.Sscanf(fmt.Sprint(value), "%d", &parsed)
		return parsed
	}
}
