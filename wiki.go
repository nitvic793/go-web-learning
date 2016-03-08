package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
    "regexp"
    "errors"
)

var templates = template.Must(template.ParseFiles("edit.html","view.html"))

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s", r.URL.Path)
}

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func getTitle(w http.ResponseWriter, r *http.Request)(string, error){
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil{
        http.NotFound(w,r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
    if err != nil {
        http.Redirect(w,r, "/edit/"+title, http.StatusFound)
        return;
    }
    renderTemplate(w,"view",p)
	//fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
    err := templates.ExecuteTemplate(w, tmpl+".html",p)
    if err!=nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w,"edit",p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string){
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err!= nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w,r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
    return func(w http.ResponseWriter, r *http.Request){
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w,r)
            return
        }
        fn(w,r,m[2])
    }
}

func defaultHandler(w http.ResponseWriter, r *http.Request, title string){
    
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}
