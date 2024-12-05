package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type AddTaskRequest struct {
	Title string `json:"title"`
}

type ToggleTaskRequest struct {
	ID int `json:"id"`
}

type DeleteTaskRequest struct {
	ID int `json:"id"`
}

type TaskStore struct {
	db *sql.DB
}

func NewTaskStore(dbPath string) (*TaskStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tasks table: %v", err)
	}

	return &TaskStore{db: db}, nil
}

func (ts *TaskStore) Close() error {
	return ts.db.Close()
}

func (ts *TaskStore) GetTasks() ([]Task, error) {
	rows, err := ts.db.Query("SELECT id, title, completed FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

func (ts *TaskStore) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := ts.GetTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (ts *TaskStore) ToggleTask(id int) error {
	_, err := ts.db.Exec("UPDATE tasks SET completed = NOT completed WHERE id = ?", id)
	return err
}

func (ts *TaskStore) ToggleTaskHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}
	var toggleTaskReq ToggleTaskRequest

	if err := json.Unmarshal(body, &toggleTaskReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := ts.ToggleTask(toggleTaskReq.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task toggled successfully"})
}

func (ts *TaskStore) DeleteTask(id int) error {
	_, err := ts.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func (ts *TaskStore) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}
	var deleteTaskReq DeleteTaskRequest

	if err := json.Unmarshal(body, &deleteTaskReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := ts.DeleteTask(deleteTaskReq.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted successfully"})
}

func (ts *TaskStore) AddTask(title string) error {
	_, err := ts.db.Exec("INSERT INTO tasks (title, completed) VALUES (?, ?)", title, false)
	return err
}

func (ts *TaskStore) AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}
	var addTaskReq AddTaskRequest

	if err := json.Unmarshal(body, &addTaskReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if addTaskReq.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if err := ts.AddTask(addTaskReq.Title); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task added successfully"})
}

func main() {
	taskStore, err := NewTaskStore("./data.db")
	if err != nil {
		log.Fatalf("Failed to initialize task store: %v", err)
	}
	defer taskStore.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Todo API"))
	})

	r.Post("/tasks", taskStore.AddTaskHandler)
	r.Post("/tasks/toggle", taskStore.ToggleTaskHandler)
	r.Post("/tasks/delete", taskStore.DeleteTaskHandler)
	r.Get("/tasks", taskStore.GetTasksHandler)

	handler := cors.Default().Handler(r)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
