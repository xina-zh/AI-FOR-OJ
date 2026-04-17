package repository

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
)

func TestAISolveAttemptRepositoryCreateAndListByRunID(t *testing.T) {
	db := openAttemptRepositoryTestDB(t)
	repo := NewAISolveAttemptRepository(db)

	ctx := context.Background()
	first := &model.AISolveAttempt{AISolveRunID: 7, AttemptNo: 2, Stage: "repair"}
	second := &model.AISolveAttempt{AISolveRunID: 7, AttemptNo: 1, Stage: "analysis"}
	otherRun := &model.AISolveAttempt{AISolveRunID: 9, AttemptNo: 1, Stage: "other"}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first attempt: %v", err)
	}
	if err := repo.Create(ctx, second); err != nil {
		t.Fatalf("create second attempt: %v", err)
	}
	if err := repo.Create(ctx, otherRun); err != nil {
		t.Fatalf("create other run attempt: %v", err)
	}

	attempts, err := repo.ListByRunID(ctx, 7)
	if err != nil {
		t.Fatalf("list attempts by run id: %v", err)
	}

	if len(attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d: %+v", len(attempts), attempts)
	}
	if attempts[0].AttemptNo != 1 || attempts[1].AttemptNo != 2 {
		t.Fatalf("expected attempts sorted by attempt no, got %+v", attempts)
	}
	if attempts[0].AISolveRunID != 7 || attempts[1].AISolveRunID != 7 {
		t.Fatalf("expected only run 7 attempts, got %+v", attempts)
	}
	if attempts[0].Stage != "analysis" || attempts[1].Stage != "repair" {
		t.Fatalf("unexpected attempt payloads: %+v", attempts)
	}
}

func openAttemptRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	store := &attemptStore{}
	driverName := fmt.Sprintf("attempt-repo-test-%d", time.Now().UnixNano())
	sql.Register(driverName, &attemptDriver{store: store})

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

	if err := db.AutoMigrate(&model.AISolveAttempt{}); err != nil {
		t.Fatalf("auto migrate attempt schema: %v", err)
	}

	return db
}

type attemptStore struct {
	mu       sync.Mutex
	nextID   int64
	attempts []model.AISolveAttempt
}

type attemptDriver struct {
	store *attemptStore
}

func (d *attemptDriver) Open(name string) (driver.Conn, error) {
	return &attemptConn{store: d.store}, nil
}

type attemptConn struct {
	store *attemptStore
}

func (c *attemptConn) Prepare(query string) (driver.Stmt, error) {
	return &attemptStmt{conn: c, query: query}, nil
}

func (c *attemptConn) Close() error { return nil }

func (c *attemptConn) Begin() (driver.Tx, error) {
	return &attemptTx{}, nil
}

func (c *attemptConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return &attemptTx{}, nil
}

func (c *attemptConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.exec(query, namedValuesToValues(args))
}

func (c *attemptConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.query(query, namedValuesToValues(args))
}

type attemptStmt struct {
	conn  *attemptConn
	query string
}

type attemptTx struct{}

func (t *attemptTx) Commit() error   { return nil }
func (t *attemptTx) Rollback() error { return nil }

func (s *attemptStmt) Close() error { return nil }

func (s *attemptStmt) NumInput() int { return -1 }

func (s *attemptStmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.conn.exec(s.query, args)
}

func (s *attemptStmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.conn.query(s.query, args)
}

func (c *attemptConn) exec(query string, args []driver.Value) (driver.Result, error) {
	if !strings.Contains(strings.ToLower(query), "insert into `ai_solve_attempts`") {
		return attemptResult(0), nil
	}
	if len(args) < 5 {
		return nil, fmt.Errorf("expected insert args, got %d", len(args))
	}

	c.store.mu.Lock()
	defer c.store.mu.Unlock()

	c.store.nextID++
	attempt := model.AISolveAttempt{
		BaseModel: model.BaseModel{
			ID:        uint(c.store.nextID),
			CreatedAt: asTime(args[0]),
			UpdatedAt: asTime(args[1]),
		},
		AISolveRunID: asUint(args[2]),
		AttemptNo:    asInt(args[3]),
		Stage:        asString(args[4]),
	}
	c.store.attempts = append(c.store.attempts, attempt)

	return attemptResult(c.store.nextID), nil
}

func (c *attemptConn) query(query string, args []driver.Value) (driver.Rows, error) {
	if !strings.Contains(strings.ToLower(query), "from `ai_solve_attempts`") {
		return &attemptRows{}, nil
	}
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

	return newAttemptRows(filtered), nil
}

type attemptResult int64

func (r attemptResult) LastInsertId() (int64, error) { return int64(r), nil }

func (r attemptResult) RowsAffected() (int64, error) { return 1, nil }

type attemptRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func newAttemptRows(attempts []model.AISolveAttempt) *attemptRows {
	rows := &attemptRows{
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

func (r *attemptRows) Columns() []string {
	return r.columns
}

func (r *attemptRows) Close() error {
	return nil
}

func (r *attemptRows) Next(dest []driver.Value) error {
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

func asString(value driver.Value) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprint(value)
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

func asInt(value driver.Value) int {
	switch v := value.(type) {
	case int64:
		return int(v)
	case int:
		return v
	case uint64:
		return int(v)
	case uint:
		return int(v)
	case []byte:
		var parsed int
		_, _ = fmt.Sscanf(string(v), "%d", &parsed)
		return parsed
	default:
		var parsed int
		_, _ = fmt.Sscanf(fmt.Sprint(value), "%d", &parsed)
		return parsed
	}
}

func asTime(value driver.Value) time.Time {
	switch v := value.(type) {
	case time.Time:
		return v
	case nil:
		return time.Time{}
	default:
		return time.Time{}
	}
}
