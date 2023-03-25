package server

import (
	"documentize/users"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"mime"
	"net/http"
	"net/url"
	"time"
)

// UserTemplates holds all users related templates.
type UserTemplates struct {
	List   *template.Template
	Create *template.Template
	Update *template.Template
}

// Users is a mvc controller that handles all admins related views.
type Users struct {
	users *users.Service

	templates UserTemplates
}

// NewUsers is a constructor for users controller.
func NewUsers(users *users.Service, templates UserTemplates) *Users {
	usersController := &Users{
		users:     users,
		templates: templates,
	}

	return usersController
}

// Create is an endpoint that will create a new user.
func (controller *Users) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		if err := controller.templates.Create.Execute(w, nil); err != nil {
			http.Error(w, "could not execute create users template", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "could not get users form", http.StatusBadRequest)
			return
		}
		email := r.FormValue("email")
		if email == "" {
			http.Error(w, "email is empty", http.StatusBadRequest)
			return
		}
		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "name is empty", http.StatusBadRequest)
			return
		}

		if err := controller.users.Create(ctx, name, email); err != nil {
			log.Println("create user err", err)
			http.Error(w, "could not create user", http.StatusInternalServerError)
			return
		}
		Redirect(w, r, "/users", http.MethodGet)
	}
}

// List is an endpoint that will provide a web page with all users.
func (controller *Users) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := controller.users.List(ctx)
	if err != nil {
		log.Println("list err", err)
		http.Error(w, "could not get users list", http.StatusInternalServerError)
		return
	}

	err = controller.templates.List.Execute(w, users)
	if err != nil {
		http.Error(w, "can not execute list users template", http.StatusInternalServerError)
		return
	}
}

// Export handles exporting of users data.
func (controller *Users) Export(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fileName, path, reader, err := controller.users.Export(ctx)
	if err != nil {
		log.Println("could not export users data", err)
		http.Error(w, "could not export users data", http.StatusInternalServerError)
		return
	}

	mediaType := mime.FormatMediaType("attachment", map[string]string{"filename": fileName})
	w.Header().Set("Content-Disposition", mediaType)
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, path, time.Time{}, reader)

	err = controller.users.DeleteGeneratedFile(fileName)
	if err != nil {
		log.Println("could not delete exported users file", err)
		http.Error(w, "could not delete exported users file", http.StatusInternalServerError)
		return
	}
}

// Generate handles users doc generating.
func (controller *Users) Generate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)

	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "could not parse user id", http.StatusBadRequest)
		return
	}

	fileName, path, reader, err := controller.users.GenerateDoc(ctx, id)
	if err != nil {
		if errors.Is(err, users.ErrAlreadyGenerated) {
			http.Error(w, "file already generated", http.StatusBadRequest)
			return
		}

		log.Println("could not generate user doc", err)
		http.Error(w, "could not generate user doc", http.StatusInternalServerError)
		return
	}

	mediaType := mime.FormatMediaType("attachment", map[string]string{"filename": fileName})
	w.Header().Set("Content-Disposition", mediaType)
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, path, time.Time{}, reader)

	err = controller.users.DeleteGeneratedFile(fileName)
	if err != nil {
		log.Println("could not delete exported users file", err)
		http.Error(w, "could not delete exported users file", http.StatusInternalServerError)
		return
	}
}

// Redirect redirects to specific url.
func Redirect(w http.ResponseWriter, r *http.Request, urlString, method string) {
	newRequest := new(http.Request)
	*newRequest = *r
	newRequest.URL = new(url.URL)
	*newRequest.URL = *r.URL
	newRequest.Method = method

	http.Redirect(w, newRequest, urlString, http.StatusFound)
}
