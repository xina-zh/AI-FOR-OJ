package repository

import (
	"context"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

func TestAISolveAttemptRepositoryListByRunIDOrdersAttempts(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry run db: %v", err)
	}

	repo := NewAISolveAttemptRepository(db)
	query := repo.listByRunIDQuery(context.Background(), 7).Find(&[]model.AISolveAttempt{})
	sql := query.Statement.SQL.String()

	if !strings.Contains(sql, "WHERE ai_solve_run_id = ?") {
		t.Fatalf("expected run id filter in SQL, got %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY attempt_no ASC, id ASC") {
		t.Fatalf("expected stable attempt ordering in SQL, got %s", sql)
	}
}
