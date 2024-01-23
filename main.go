package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"html/template"
	"log"
	"net/http"
	"os"
)

type checkbox struct {
	Checked bool
}

func (cb *checkbox) Toggle(w http.ResponseWriter, r *http.Request) {
	cb.Checked = !cb.Checked

	fmt.Printf("%v\n", r.Body)
	fmt.Printf("%v\n", r.Header)
	tmpl, _ := template.New("checkboxLabel").Parse(`<label id='cb_label' for='test-checkbox'>{{.}}</label>`)
	label := "Unchecked"
	if cb.Checked {
		label = "Checked"
	}

	tmpl.Execute(w, label)
}

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT not found in .env")
	}

	addr := fmt.Sprintf(":" + port)

	r := http.NewServeMux()
	r.Handle("/", http.FileServer(http.Dir("./templates/")))

	cb := checkbox{
		Checked: false,
	}

	r.HandleFunc("/checkbox", cb.Toggle)

	s := http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Fatal(s.ListenAndServe())
}
