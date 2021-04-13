package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	Storage
	*mux.Router
	Aero    *Aerospike
	CacheOn bool
}

func NewApp() *App {
	var a = &App{
		Storage: NewStorage("postgres", "host=localhost port=5432 user=postgres password=password dbname=shorturl sslmode=disable"),
		Aero:    NewAerospike("localhost", 3000, "test"),
		CacheOn: false,
	}
	a.Router = NewRouter(a)
	return a
}

func main() {
	var app = NewApp()

	app.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./index.html")
	}).Methods("GET")

	fmt.Println("Listening on http://localhost:4000")

	if app.CacheOn {
		fmt.Println("Cache enabled")
	} else {
		fmt.Println("Cache disabled")
	}

	if err := http.ListenAndServe(":4000", app); err != nil {
		log.Fatal(err)
	}
}
