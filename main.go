package main

import (
	"embed"
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
}

type dataconfig struct {
	TDItem    *todoitem
	Templates *template.Template
}

func (cfg *dataconfig) Toggle(w http.ResponseWriter, r *http.Request) {
	cfg.TDItem.Checked = !cfg.TDItem.Checked
	if cfg.TDItem.Checked {
		cfg.Templates.ExecuteTemplate(w, "todo-item-checked.html", cfg.TDItem)
	} else {
		cfg.Templates.ExecuteTemplate(w, "todo-item-unchecked.html", cfg.TDItem)
	}
}

func (cfg *dataconfig) HandleIndex(w http.ResponseWriter, r *http.Request) {
	cfg.Templates.ExecuteTemplate(w, "index.html", cfg.TDItem)
}

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT not found in .env")
	}

	addr := fmt.Sprintf(":" + port)

	r := http.NewServeMux()

	td := todoitem{
		Checked:   false,
		Label:     "Test ToDo Item",
		LabelId:   "label_id",
		InputName: "cb_name",
		InputId:   "input_id",
	}

	pattern := filepath.Join("templates", "*.html")
	templates := template.Must(template.ParseGlob(pattern))

	config := dataconfig{
		TDItem:    &td,
		Templates: templates,
	}

	r.HandleFunc("/", config.HandleIndex)
	r.HandleFunc("/checkbox", config.Toggle)
	r.Handle("/css/output.css", http.FileServer(http.FS(css)))

	s := http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Fatal(s.ListenAndServe())
}
