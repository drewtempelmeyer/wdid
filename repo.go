package wdid

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Task struct {
	ID          uint64       `db:"id"`
	Description string       `db:"description"`
	CreatedAt   time.Time    `db:"created_at"`
	CompletedAt sql.NullTime `db:"completed_at"`
}

type TaskRepository interface {
	AddTask(*Task) error
	TasksForDate(time.Time) ([]Task, error)
	PendingTasks() ([]Task, error)
	CompleteTasks([]int64) error
	DeleteTasks([]int64) error
}

// NewTask generates a new Task struct with the description and the current time
func NewTask(description string) *Task {
	return &Task{
		Description: description,
		CreatedAt:   time.Now(),
	}
}

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	initDB(db)
	return &SQLRepository{db: db}
}

func initDB(db *sqlx.DB) {
	q := `CREATE TABLE IF NOT EXISTS tasks(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        description TEXT NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME
    );

	CREATE INDEX IF NOT EXISTS ix_tasks_created_at ON tasks(created_at);
	CREATE INDEX IF NOT EXISTS ix_tasks_completed_at ON tasks(completed_at);
	`

	db.MustExec(q)
}

func (r *SQLRepository) AddTask(t *Task) error {
	q := `INSERT INTO tasks (description, created_at, completed_at) VALUES (?, ?, ?)`

	_, err := r.db.Exec(q, t.Description, t.CreatedAt, t.CompletedAt)

	if err != nil {
		return fmt.Errorf("could not create task: %s", err)
	}

	return nil
}

func (r *SQLRepository) TasksForDate(t time.Time) ([]Task, error) {
	tasks := []Task{}
	bindings := map[string]interface{}{
		"start": beginningOfDay(t),
		"end":   endOfDay(t),
	}

	q := `
	SELECT *
	FROM tasks
	WHERE (created_at >= :start AND created_at <= :end) OR
		  (completed_at >= :start AND completed_at <= :end)
	ORDER BY completed_at NULLS LAST, created_at;`
	nstmt, err := r.db.PrepareNamed(q)

	if err != nil {
		return nil, fmt.Errorf("error fetching tasks: %s", err.Error())
	}

	err = nstmt.Select(&tasks, bindings)

	if err != nil {
		return nil, fmt.Errorf("error fetching tasks: %s", err.Error())
	}

	return tasks, nil
}

// PendingTasks returns the tasks that have not been completed
func (r *SQLRepository) PendingTasks() ([]Task, error) {
	tasks := []Task{}
	q := `SELECT * FROM tasks WHERE created_at IS NOT NULL AND completed_at IS NULL;`
	err := r.db.Select(&tasks, q)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *SQLRepository) CompleteTasks(ids []int64) error {
	arg := map[string]interface{}{
		"time": time.Now(),
		"ids":  ids,
	}

	query, args, err := sqlx.Named("UPDATE tasks SET completed_at = :time WHERE id IN (:ids) AND completed_at IS NULL", arg)
	query, args, err = sqlx.In(query, args...)

	if err != nil {
		return err
	}

	query = r.db.Rebind(query)
	r.db.MustExec(query, args...)

	return nil
}

func (r *SQLRepository) DeleteTasks(ids []int64) error {
	q := "DELETE FROM tasks WHERE id IN (?)"
	query, args, err := sqlx.In(q, ids)

	if err != nil {
		return err
	}

	query = r.db.Rebind(query)
	r.db.MustExec(query, args...)

	return nil
}

// beginningOfDay beginning of day
func beginningOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}
