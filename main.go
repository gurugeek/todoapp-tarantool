package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	"github.com/viciious/go-tarantool"
)

var rnd *renderer.Render
var db *tarantool.Connector

const (
	hostName       = "localhost:3301"
	login          = "todo"
	password       = "test"
	collectionName = "todo"
	port           = ":9000"
)

func init() {
	rnd = renderer.New()

	db = tarantool.New(hostName, &tarantool.Options{
		User:     login,
		Password: password,
	})
	_, err := db.Connect()
	checkErr(err)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
	checkErr(err)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t todoModel

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusBadRequest, err)
		return
	}

	// simple validation
	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is requried",
		})
		return
	}

	// if input is okay, create a todo
	tm := todoModel{
		ID:        uint64(rand.Uint32()),
		Title:     t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	con, err := db.Connect()
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to establish connect to db",
			"error":   err,
		})
	}

	q := &tarantool.Insert{
		Space: collectionName,
		Tuple: tm.Pack(),
	}

	if _, err := con.Execute(q); err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to save todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "Todo created successfully",
		"id": tm.ID,
	})
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	idString := strings.TrimSpace(chi.URLParam(r, "id"))

	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	var t todoModel

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusBadRequest, err)
		return
	}

	// simple validation
	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is requried",
		})
		return
	}

	con, err := db.Connect()
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to establish connect to db",
			"error":   err,
		})
	}

	q := &tarantool.Update{
		Space: collectionName,
		Key:   id,
		Set: []tarantool.Operator{
			&tarantool.OpAssign{Field: 1, Argument: t.Title},
			&tarantool.OpAssign{Field: 2, Argument: t.Completed},
		},
	}

	if _, err := con.Execute(q); err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to update todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
	})
}

func fetchTodos(w http.ResponseWriter, r *http.Request) {
	con, err := db.Connect()
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "Failed to establish connect to db",
			"error":   err,
		})
	}

	q := &tarantool.Select{
		Space:    collectionName,
		Index:    "created",
		Key:      uint64(math.MaxUint64),
		Iterator: tarantool.IterLe,
	}

	resp, err := con.Execute(q)
	if err != nil {
		log.Println(err)
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to fetch todo",
			"error":   err,
		})
		return
	}

	todoList := make([]todoModel, 0, len(resp))
	for _, t := range resp {
		m := todoModel{}
		err := m.Unpack(t)

		if err != nil {
			todoList = append(todoList, m)
		}
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	idString := strings.TrimSpace(chi.URLParam(r, "id"))

	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	con, err := db.Connect()
	if err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to establish connect to db",
			"error":   err,
		})
	}

	q := &tarantool.Delete{
		Space: collectionName,
		Key:   id,
	}

	if _, err := con.Execute(q); err != nil {
		rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to delete todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo deleted successfully",
	})
}

func main() {
	rand.Seed(time.Now().Unix())

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)

	r.Mount("/todo", todoHandlers())

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Listening on port ", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	log.Println("Server gracefully stopped!")
}

func todoHandlers() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err) //respond with error page or message
	}
}
