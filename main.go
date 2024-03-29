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

type filterdata struct {
	Active int
	Total  int
	Filter string
}

type dataconfig struct {
	TDItems    *[]todoitem
	Templates  *template.Template
	FilterData *filterdata
}

func parseIdFromURL(r *http.Request) int {
	id := path.Base(r.URL.String())
	intId, _ := strconv.Atoi(id)
	return intId
}

func reIdTodos(todos []todoitem) []todoitem {
	for i := range todos {
		todos[i].Id = fmt.Sprintf("%d", i)
		todos[i].LabelId = fmt.Sprintf("label_%d", i)
		todos[i].Editing = false
		todos[i].InputName = fmt.Sprintf("checkbox_%d", i)
	}

	return todos
}

func (cfg *dataconfig) filterItems() []todoitem {
	switch cfg.FilterData.Filter {
	case "Active":
		activeTodos := []todoitem{}
		for _, todo := range *cfg.TDItems {
			if !todo.Checked {
				activeTodos = append(activeTodos, todo)
			}
		}
		return activeTodos
	case "Completed":
		completedTodos := []todoitem{}
		for _, todo := range *cfg.TDItems {
			if todo.Checked {
				completedTodos = append(completedTodos, todo)
			}
		}
		return completedTodos
	}
	return *cfg.TDItems
}

func (cfg *dataconfig) handleFooter(w http.ResponseWriter, r *http.Request) {
	if cfg.FilterData.Filter == "" {
		cfg.FilterData.Filter = "All"
	}
	switch r.Method {
	case http.MethodGet:
		cfg.Templates.ExecuteTemplate(w, "footer.html", *cfg.FilterData)
	case http.MethodPost:
		w.Header().Add("HX-Trigger", "todosUpdated")
		cfg.FilterData.Filter = path.Base(r.URL.String())
		cfg.Templates.ExecuteTemplate(w, "todo-list.html", cfg.filterItems())
	case http.MethodDelete:
		w.Header().Add("HX-Trigger", "todosUpdated")
		temp := cfg.FilterData.Filter
		cfg.FilterData.Filter = "Active"
		filteredItems := reIdTodos(cfg.filterItems())
		cfg.TDItems = &filteredItems
		cfg.FilterData.Filter = temp
		cfg.FilterData.Active = len(filteredItems)
		cfg.FilterData.Total = len(filteredItems)
		cfg.Templates.ExecuteTemplate(w, "todo-list.html", cfg.filterItems())
	}
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
			cfg.FilterData.Active -= 1
		} else {
			itemSlice[intId].Class = ""
			cfg.FilterData.Active += 1
		}

		w.Header().Add("HX-Trigger", "todosUpdated")
		cfg.TDItems = &itemSlice
		cfg.Templates.ExecuteTemplate(w, "todo-list.html", cfg.filterItems())
	}
}

func (cfg *dataconfig) handleTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("HX-Trigger", "todosUpdated")
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
		cfg.FilterData.Active += 1
		cfg.FilterData.Total += 1
		cfg.Templates.ExecuteTemplate(w, "todo-list.html", cfg.filterItems())
	case http.MethodDelete:
		intId := parseIdFromURL(r)
		itemSlice := []todoitem{}
		for i, todo := range *cfg.TDItems {
			if i != intId {
				itemSlice = append(itemSlice, todo)
			} else if !todo.Checked {
				cfg.FilterData.Active -= 1
			}
		}

		cfg.FilterData.Total -= 1
		itemSlice = reIdTodos(itemSlice)
		cfg.TDItems = &itemSlice
		cfg.Templates.ExecuteTemplate(w, "todo-list.html", cfg.filterItems())
	}
}

func (cfg *dataconfig) handleIndex(w http.ResponseWriter, r *http.Request) {
	cfg.TDItems = &[]todoitem{}
	cfg.FilterData = &filterdata{}

	type indexdata struct {
		Items      []todoitem
		FilterData filterdata
	}

	data := indexdata{
		Items:      *cfg.TDItems,
		FilterData: *cfg.FilterData,
	}

	cfg.Templates.ExecuteTemplate(w, "index.html", data)
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
		TDItems:    &[]todoitem{},
		Templates:  templates,
		FilterData: &filterdata{},
	}

	r.HandleFunc("/", config.handleIndex)
	r.Handle("/css/output.css", http.FileServer(http.FS(css)))
	r.HandleFunc("/todos/", config.handleTodo)
	r.HandleFunc("/todos/toggle/", config.handleToggle)
	r.HandleFunc("/todos/edit/", config.handleEdit)
	r.HandleFunc("/footer/", config.handleFooter)

	s := http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Fatal(s.ListenAndServe())
}
