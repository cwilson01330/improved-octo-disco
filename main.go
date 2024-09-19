package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	// These are not included in the standard library
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Define a struct to represent a todo item
type Todo struct {
	gorm.Model
	ID    uint
	Title string
	Done  bool
}

// Define a struct to represent the data that will be passed to the template
type TodoPageData struct {
	PageTitle string
	Todos     []Todo
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Todo{})

	tmpl := template.Must(template.ParseFiles("template/index.html"))
	tmpl2 := template.Must(template.ParseFiles("template/deletedItems.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			}
			todo := r.FormValue("todo")
			db.Create(&Todo{Title: todo, Done: false})
		}
		//Request not POST
		var todos []Todo
		db.Find(&todos)
		data := TodoPageData{
			PageTitle: "My TODO list",
			Todos:     todos,
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/done/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/done/")
		var todo Todo
		db.First(&todo, id)
		todo.Done = true
		db.Save(&todo)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/delete/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/delete/")
		db.Delete(&Todo{}, id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Unscoped provides access to soft deleted items that are not visible by default
	http.HandleFunc("/deletedItems/", func(w http.ResponseWriter, r *http.Request) {
		var todos []Todo
		db.Unscoped().Where("deleted_at IS NOT NULL").Find(&todos)
		data := TodoPageData{
			PageTitle: "Deleted Items",
			Todos:     todos,
		}
		tmpl2.Execute(w, data)
	})

	http.ListenAndServe(":8080", nil)
}
