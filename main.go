package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

type VM struct {
	Name     string
	Settings settings
}

type Credentials struct {
	Username string
	Password string
}

func index(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	viewModel := VM{
		Name:     "testing",
		Settings: set,
	}

	fmt.Fprintln(os.Stdout, req.Method)
	fmt.Fprintln(os.Stdout, req.URL)
	fmt.Fprintln(os.Stdout, req.RequestURI)
	fmt.Fprintln(os.Stdout, req.Body)

	err := tpl.ExecuteTemplate(w, "index.gohtml", viewModel)
	check(err)
}

func about(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	viewModel := VM{
		Settings: set,
	}

	err := tpl.ExecuteTemplate(w, "about.gohtml", viewModel)
	check(err)
}

func someJSON(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	set = getSettings()
	w.Header().Set("Content-Type", "application/json")

	c, err := json.Marshal(VM{
		Name:     "I am JSON",
		Settings: set,
	})
	check(err)

	i, err := w.Write(c)
	check(err)
	fmt.Fprintln(os.Stdout, "i: ", i)
}

func showForm(w http.ResponseWriter, req *http.Request, par httprouter.Params) {
	// viewModel := VM{
	// 	Settings: set,
	// }
	req.ParseForm()
	fmt.Fprintln(os.Stdout, par)
	fmt.Fprintln(os.Stdout, req.Method)
	fmt.Fprintln(os.Stdout, req.Form)
	fmt.Fprintln(os.Stdout, req.FormValue("username"))

	if len(par) > 0 {
		creds := Credentials{
			Username: req.Form.Get("username"),
			Password: req.Form.Get("password"),
		}

		err := tpl.ExecuteTemplate(w, "form.gohtml", creds)
		check(err)
		return
	}

	err := tpl.ExecuteTemplate(w, "form.gohtml", nil)
	check(err)

}

func file(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	var s string

	if req.Method == http.MethodPost {
		f, fh, err := req.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		fmt.Println("\nFile: ", f, "\nHeaders: ", fh, "e\n Error: ", err)

		// read content of file
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// write content of file to ./files directory
		dst, err := os.Create(filepath.Join("./files", fh.Filename))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = dst.Write(bs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		s = string(bs)
	}

	err := tpl.ExecuteTemplate(w, "file.gohtml", s)
	check(err)
}

var tpl *template.Template
var set settings

type settings struct {
	AppName string
}

func init() {
	set = getSettings()
	tpl = template.Must(template.ParseGlob("**/*.gohtml"))
}

func main() {
	router := httprouter.New()
	router.GET("/", index)
	router.GET("/about/", about)
	router.GET("/json/", someJSON)
	router.GET("/form/*test", showForm) // catch all parameter with *
	router.POST("/form/*test", showForm)
	router.GET("/file/", file)
	router.POST("/file/", file)

	router.ServeFiles("/assets/*filepath", http.Dir("assets"))

	log.Fatal(http.ListenAndServe(":8080", router))
}

func getSettings() settings {
	c, err := ioutil.ReadFile("settings.json")
	if err != nil {
		log.Fatalln(err)
	}

	var o settings
	err = json.Unmarshal(c, &o)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintln(os.Stdout, o)
	return o
}

func check(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}
