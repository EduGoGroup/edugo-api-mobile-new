package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

var (
	testDB    *sql.DB
	setupOnce sync.Once
	setupErr  error
)

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		log.Println("Skipping integration tests (set INTEGRATION_TESTS=true to run)")
		os.Exit(0)
	}

	code := m.Run()

	// Cleanup
	if testDB != nil {
		testDB.Close()
	}

	os.Exit(code)
}

func getTestDB(t *testing.T) *sql.DB {
	t.Helper()

	setupOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		dsn := os.Getenv("TEST_POSTGRES_DSN")
		if dsn == "" {
			setupErr = fmt.Errorf("TEST_POSTGRES_DSN environment variable is required for integration tests")
			return
		}

		db, err := sql.Open("postgres", dsn)
		if err != nil {
			setupErr = fmt.Errorf("opening postgres: %w", err)
			return
		}

		if err := db.PingContext(ctx); err != nil {
			setupErr = fmt.Errorf("pinging postgres: %w", err)
			return
		}

		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(2)

		testDB = db

		// Run migrations / ensure schema exists
		if err := runMigrations(ctx, db); err != nil {
			setupErr = fmt.Errorf("running migrations: %w", err)
			return
		}
	})

	if setupErr != nil {
		t.Fatalf("test setup failed: %v", setupErr)
	}

	return testDB
}

// runMigrations ensures the tables required by integration tests exist.
// In a real project this would run the migration files. Here we create minimal tables.
func runMigrations(_ context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS schools (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email TEXT NOT NULL,
			school_id UUID REFERENCES schools(id),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS materials (
			id UUID PRIMARY KEY,
			school_id UUID NOT NULL REFERENCES schools(id),
			uploaded_by_teacher_id UUID NOT NULL REFERENCES users(id),
			academic_unit_id UUID,
			title TEXT NOT NULL,
			description TEXT,
			subject TEXT,
			grade TEXT,
			file_url TEXT DEFAULT '',
			file_type TEXT DEFAULT '',
			file_size_bytes BIGINT DEFAULT 0,
			status TEXT DEFAULT 'draft',
			processing_started_at TIMESTAMPTZ,
			processing_completed_at TIMESTAMPTZ,
			is_public BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS assessment (
			id UUID PRIMARY KEY,
			material_id UUID NOT NULL REFERENCES materials(id),
			mongo_document_id TEXT DEFAULT '',
			questions_count INT DEFAULT 0,
			total_questions INT,
			title TEXT,
			pass_threshold INT,
			max_attempts INT,
			time_limit_minutes INT,
			status TEXT DEFAULT 'draft',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS assessment_attempt (
			id UUID PRIMARY KEY,
			assessment_id UUID NOT NULL REFERENCES assessment(id),
			student_id UUID NOT NULL,
			started_at TIMESTAMPTZ NOT NULL,
			completed_at TIMESTAMPTZ,
			score DECIMAL(5,2),
			max_score DECIMAL(5,2),
			percentage DECIMAL(5,2),
			time_spent_seconds INT,
			idempotency_key TEXT,
			status TEXT DEFAULT 'in_progress',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS assessment_attempt_answer (
			id UUID PRIMARY KEY,
			attempt_id UUID NOT NULL REFERENCES assessment_attempt(id),
			question_index INT NOT NULL,
			student_answer TEXT,
			is_correct BOOLEAN,
			points_earned DECIMAL(5,2),
			max_points DECIMAL(5,2),
			time_spent_seconds INT,
			answered_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS progress (
			material_id UUID NOT NULL,
			user_id UUID NOT NULL,
			percentage INT DEFAULT 0,
			last_page INT DEFAULT 0,
			status TEXT DEFAULT 'not_started',
			last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (material_id, user_id)
		)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("executing migration: %w\n  SQL: %s", err, stmt)
		}
	}
	return nil
}

// setupTestSchool inserts a test school record.
func setupTestSchool(t *testing.T, db *sql.DB, schoolID uuid.UUID) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO schools (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, schoolID, "Test School")
	if err != nil {
		t.Fatalf("setup test school: %v", err)
	}
}

// setupTestUser inserts a test user record.
func setupTestUser(t *testing.T, db *sql.DB, userID, schoolID uuid.UUID) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO users (id, email, school_id) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`,
		userID, fmt.Sprintf("user-%s@test.com", userID.String()[:8]), schoolID,
	)
	if err != nil {
		t.Fatalf("setup test user: %v", err)
	}
}

// cleanupMaterial deletes test data in reverse FK order.
func cleanupMaterial(t *testing.T, db *sql.DB, materialID uuid.UUID) {
	t.Helper()
	db.Exec("DELETE FROM assessment_attempt_answer WHERE attempt_id IN (SELECT id FROM assessment_attempt WHERE assessment_id IN (SELECT id FROM assessment WHERE material_id = $1))", materialID)
	db.Exec("DELETE FROM assessment_attempt WHERE assessment_id IN (SELECT id FROM assessment WHERE material_id = $1)", materialID)
	db.Exec("DELETE FROM assessment WHERE material_id = $1", materialID)
	db.Exec("DELETE FROM progress WHERE material_id = $1", materialID)
	db.Exec("DELETE FROM materials WHERE id = $1", materialID)
}
