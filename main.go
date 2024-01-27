package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

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
	Editing   bool
	Label     string
	LabelId   string
	Id        string
	InputName string
	Class     string
}

type dataconfig struct {
	TDItems   *[]todoitem
	Templates *template.Template
}

func parseIdFromURL(r *http.Request) int {
	id := path.Base(r.URL.String())
	intId, _ := strconv.Atoi(id)
	return intId
}

func (cfg *dataconfig) handleEdit(w http.ResponseWriter, r *http.Request) {
	intId := parseIdFromURL(r)
	itemSlice := *cfg.TDItems
	switch r.Method {
	case http.MethodGet:
		itemSlice[intId].Editing = true

		cfg.Templates.ExecuteTemplate(w, "todo-item.html", itemSlice[intId])
	case http.MethodPut:
		type todo struct {
			Value string `json:"todo"`
		}

		newTodo := todo{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newTodo)
		if err != nil {
			fmt.Printf("Err decoding json: %v\n", err)
		}

		itemSlice[intId].Label = newTodo.Value
		itemSlice[intId].Editing = false

		cfg.Templates.ExecuteTemplate(w, "todo-item.html", itemSlice[intId])
	}
	cfg.TDItems = &itemSlice
}

func (cfg *dataconfig) handleToggle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		intId := parseIdFromURL(r)
		itemSlice := *cfg.TDItems
		itemSlice[intId].Checked = !itemSlice[intId].Checked
		if itemSlice[intId].Checked {
			itemSlice[intId].Class = "completed"
		} else {
			itemSlice[intId].Class = ""
		}

		cfg.TDItems = &itemSlice
		cfg.Templates.ExecuteTemplate(w, "todo-item.html", itemSlice[intId])
	}
}

func (cfg *dataconfig) handleTodo(w http.ResponseWriter, r *http.Request) {
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
			Editing:   false,
			Label:     newTodo.Value,
			LabelId:   fmt.Sprintf("label_%d", numTodos),
			Id:        fmt.Sprintf("%d", numTodos),
			InputName: fmt.Sprintf("checkbox_%d", numTodos),
			Class:     "",
		})

		cfg.TDItems = &newList

		cfg.Templates.ExecuteTemplate(w, "todo-list.html", cfg.TDItems)
	case http.MethodDelete:
		intId := parseIdFromURL(r)
		itemSlice := []todoitem{}
		for i, todo := range *cfg.TDItems {
			if i != intId {
				newTodo := todoitem{
					Id:        fmt.Sprintf("%d", i),
					Label:     todo.Label,
					LabelId:   fmt.Sprintf("label_%d", i),
					Checked:   todo.Checked,
					Editing:   false,
					Class:     todo.Class,
					InputName: fmt.Sprintf("checkbox_%d", i),
				}
				itemSlice = append(itemSlice, newTodo)
			}
		}

		cfg.Templates.ExecuteTemplate(w, "todo-list.html", itemSlice)
		cfg.TDItems = &itemSlice
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
	r.HandleFunc("/todos/", config.handleTodo)
	r.HandleFunc("/todos/toggle/", config.handleToggle)
	r.HandleFunc("/todos/edit/", config.handleEdit)

	s := http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Fatal(s.ListenAndServe())
}
