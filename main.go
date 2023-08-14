package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/davecgh/go-spew/spew"
)

// EmailRx Very simple email address regex just to eliminate obvious mistakes
var EmailRx = regexp.MustCompile(`^\S+@\S+$`)

type ContactForm struct {
	Name           string
	Email          string
	Message        string
	FavoriteNumber uint8
	HearBack       bool
	Success        bool
	Errors         map[string][]string
}

func (f *ContactForm) Parse(r *http.Request) {
	f.Name = r.PostFormValue("email")
	f.Email = r.PostFormValue("email")
	f.Message = r.PostFormValue("message")

	if r.PostFormValue("hear_back") == "on" {
		f.HearBack = true
	}

	favNumStr := r.PostFormValue("fav_number")
	if len(favNumStr) > 0 {
		// TODO: better handle this error in production
		favNumInt, err := strconv.Atoi(favNumStr)
		if err != nil {
			log.Fatal(err)
		}
		f.FavoriteNumber = uint8(favNumInt)
	}
}

func (f *ContactForm) Validate() {
	f.Errors = make(map[string][]string)

	if !EmailRx.MatchString(f.Email) {
		f.Errors["email"] = append(f.Errors["email"], "Invalid format for email")
	}
	if len(f.Email) < 3 || len(f.Email) > 320 {
		f.Errors["email"] = append(f.Errors["email"], "Email should be between 3 and 320 characters long")
	}
	if len(f.Message) < 5 || len(f.Message) > 1000 {
		f.Errors["message"] = append(f.Errors["message"], "Message should be between 5 and 1000 characters long")
	}
	if f.FavoriteNumber < 1 || f.FavoriteNumber > 10 {
		f.Errors["fav_number"] = append(f.Errors["fav_number"], "Favorite number should be between 1 and 10")
	}

	f.Success = len(f.Errors) == 0
}

func parse(file string, layout string) (*template.Template, error) {
	return template.New(layout).
		Funcs(map[string]any{
			"spew": spew.Sdump,
		}).
		ParseFS(getFS(), "*")
}

func getFS() fs.FS {
	return os.DirFS("./html")
}

func renderError(err error, w http.ResponseWriter) {
	fmt.Fprintf(w, err.Error())
}

func contactFormHandler(w http.ResponseWriter, r *http.Request) {
	var form ContactForm
	if r.Method == "POST" {
		form.Parse(r)
		form.Validate()
		spew.Dump(form)
	}

	contactTemplate, err := parse("contact.html", "full")
	if err != nil {
		renderError(err, w)
		return
	}

	data := map[string]any{"contactForm": form}
	err = contactTemplate.ExecuteTemplate(w, "full", data)
	if err != nil {
		renderError(err, w)
		return
	}
}

func main() {
	http.HandleFunc("/", contactFormHandler)

	fmt.Println("Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
