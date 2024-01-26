package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var (
	//go:embed all:templates/*
	templateFS embed.FS

	//go:embed css/output.css
	css embed.FS

	//parsed templates
	html *template.Template
)

type todoitem struct {
	Checked   bool
	Label     string
	LabelId   string
	InputId   string
	InputName string
	Class     string
}

type dataconfig struct {
	TDItems   *[]todoitem
	Templates *template.Template
}

func (cfg *dataconfig) handleNewTodo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		type todo struct {
			Value string `json:"todo"`
		}

		newTodo := todo{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newTodo)
		if err != nil {
			fmt.Printf("Err decoding json: %v\n", err)
		}

		numTodos := len(*cfg.TDItems)

		newList := append(*cfg.TDItems, todoitem{
			Checked:   false,
			Label:     newTodo.Value,
			LabelId:   fmt.Sprintf("label_%d", numTodos),
			InputId:   fmt.Sprintf("%d", numTodos),
			InputName: fmt.Sprintf("checkbox_%d", numTodos),
			Class:     "",
		})

		cfg.TDItems = &newList

		cfg.Templates.ExecuteTemplate(w, "todo-list.html", cfg.TDItems)
	}
}

func (cfg *dataconfig) handleIndex(w http.ResponseWriter, r *http.Request) {
	cfg.TDItems = &[]todoitem{}
	cfg.Templates.ExecuteTemplate(w, "index.html", cfg.TDItems)
}

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT not found in .env")
	}

	addr := fmt.Sprintf(":" + port)

	r := http.NewServeMux()

	pattern := filepath.Join("templates", "*.html")
	templates := template.Must(template.ParseGlob(pattern))

	config := dataconfig{
		TDItems:   &[]todoitem{},
		Templates: templates,
	}

	r.HandleFunc("/", config.handleIndex)
	r.Handle("/css/output.css", http.FileServer(http.FS(css)))
	r.HandleFunc("/todos", config.handleNewTodo)

	s := http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Fatal(s.ListenAndServe())
}
