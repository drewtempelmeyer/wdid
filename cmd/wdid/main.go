package main

import (
	"bytes"
	"fmt"
	"log"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/drewtempelmeyer/wdid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

func main() {
	db, err := sqlx.Open("sqlite3", dbPath())
	defer db.Close()

	if err != nil {
		log.Fatalf("error connecting to database: %s", err)
	}

	repo := wdid.NewSQLRepository(db)
	rootCmd := &cobra.Command{
		Use: "wdid",
	}

	rootCmd.AddCommand(
		didCmd(repo),
		doCmd(repo),
		completeCmd(repo),
		deleteCmd(repo),
		standupCmd(repo),
	)

	rootCmd.Execute()
}

func didCmd(r wdid.TaskRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "did",
		Short: "Logs a task you did",
		Long:  "Logs a task you performed to keep track of your tasks. Helpful for stand-ups, reviews, and retro.",
		Run: func(cmd *cobra.Command, args []string) {
			task := wdid.NewTask(strings.Join(args, " "))
			task.CompletedAt.Time = task.CreatedAt
			task.CompletedAt.Valid = true
			err := r.AddTask(task)

			if err != nil {
				log.Fatalf("Error adding task: %s\n", err)
			}
		},
	}
}

func doCmd(r wdid.TaskRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "do",
		Short: "Logs a task you're planning on doing",
		Long:  "Logs a task you're planning on doing to keep track of your tasks. Helpful for stand-ups, reviews, and retro.",
		Run: func(cmd *cobra.Command, args []string) {
			task := wdid.NewTask(strings.Join(args, " "))
			err := r.AddTask(task)

			if err != nil {
				log.Fatalf("Error adding task: %s\n", err)
			}
		},
	}
}

func completeCmd(r wdid.TaskRepository) *cobra.Command {
	return &cobra.Command{
		Use:     "complete",
		Short:   "Marks the task(s) as completed",
		Example: "wdid complete 23 24",
		Run: func(cmd *cobra.Command, args []string) {
			ids := stringSliceToInt64(args)
			err := r.CompleteTasks(ids...)

			if err != nil {
				log.Fatalf("Error completing task: %s\n", err)
			}
		},
	}
}

func deleteCmd(r wdid.TaskRepository) *cobra.Command {
	return &cobra.Command{
		Use:     "complete",
		Short:   "Marks the task(s) as completed",
		Example: "wdid complete 23 24",
		Run: func(cmd *cobra.Command, args []string) {
			ids := stringSliceToInt64(args)
			err := r.DeleteTasks(ids...)

			if err != nil {
				log.Fatalf("Error deleting task: %s\n", err)
			}
		},
	}
}

func standupCmd(r wdid.TaskRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "standup",
		Short: "Generates a stand-up report in Markdown format.",
		Run: func(cmd *cobra.Command, args []string) {
			report, err := genStandup(r)

			if err != nil {
				log.Fatalf("Error generating report: %s\n", err)
			}

			fmt.Print(report)
		},
	}
}

func genStandup(r wdid.TaskRepository) (string, error) {
	today := time.Now()
	y := today.Add(time.Hour * -24)

	y_tasks, err := r.TasksForDate(y)

	if err != nil {
		return "", err
	}

	t_tasks, err := r.TasksForDate(today)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	// Yesterday's tasks
	if len(y_tasks) > 0 {
		fmt.Fprintf(&buf, "## Yesterday (%s)\n", y.Format("Mon Jan 2, 2006"))
	}

	for _, task := range y_tasks {
		// Task is not completed, but was created on
		if !task.CompletedAt.Valid {
			fmt.Fprintf(&buf, "- [STARTED] %s (%d)\n", task.Description, task.ID)
			continue
		}

		complete := task.CompletedAt.Time
		started := task.CreatedAt

		if complete.Year() == started.Year() && complete.YearDay() == started.YearDay() {
			fmt.Fprintf(&buf, "- %s (%d)\n", task.Description, task.ID)
		} else {
			fmt.Fprintf(&buf, "- [COMPLETED] %s (%d)\n", task.Description, task.ID)
		}
	}

	// Add today's tasks
	if len(t_tasks) > 0 {
		// Yesterday was printed so we need to add new lines
		if len(y_tasks) > 0 {
			fmt.Fprint(&buf, "\n\n")
		}

		fmt.Fprint(&buf, "## Today\n")
	}

	for _, task := range t_tasks {
		fmt.Fprintf(&buf, "- %s (%d)\n", task.Description, task.ID)
	}

	// In progress
	p_tasks, err := r.PendingTasks()

	if err != nil {
		return "nil", err
	}

	if len(p_tasks) > 0 {
		// Yesterday or Today was printed so we need to add new lines
		if len(y_tasks) > 0 || len(t_tasks) > 0 {
			fmt.Fprint(&buf, "\n\n")
		}

		fmt.Fprint(&buf, "## TODO\n")

		for _, task := range p_tasks {
			fmt_date := task.CreatedAt.Format("2006-01-02")
			fmt.Fprintf(&buf, "- %s (%d; started on %s)\n", task.Description, task.ID, fmt_date)
		}
	}

	return buf.String(), nil
}

func stringSliceToInt64(strs []string) []int64 {
	ids := []int64{}
	for _, s := range strs {
		id, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}

	return ids
}

// dbPath returns the sqlite3 database path
func dbPath() string {
	user, err := user.Current()

	if err != nil {
		log.Fatalf(err.Error())
	}

	home := user.HomeDir
	return filepath.Join(home, ".wdid-db")
}
