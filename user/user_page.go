package user

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
)

type Page struct {
	store  Store
	logger *zap.Logger
}

func NewPage(store Store) *Page {
	return &Page{
		store: store,
	}
}

func (page *Page) RenderLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles("templates/layout.html", "templates/login.html")
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile login template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func (page *Page) CreateSession(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	type ReqBody struct {
		Username string
		Password string
	}
	rb := &ReqBody{
		Username: username,
		Password: password,
	}
	fmt.Printf("%+v", rb)

	//row := us.DB.QueryRow(`
	//	SELECT id, password_hash
	//	FROM users WHERE username=$1`, username)
	//err := row.Scan(&user.ID, &user.PasswordHash)
	//if err != nil {
	//	return nil, fmt.Errorf("authenticate error: %w", err)
	//}

	//err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	//if err != nil {
	//	return nil, fmt.Errorf("authenticate error: %w", err)
	//}
}

func (page *Page) RenderNew(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles("templates/layout.html", "templates/signup.html")
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile signup template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func (page *Page) SaveNew(w http.ResponseWriter, r *http.Request) {
	// get values
	username := r.FormValue("username")
	email := r.FormValue("email")
	email = strings.ToLower(email)

	// generate password
	password := r.FormValue("password")
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		panic(err)
	}
	passwordHash := string(hashedBytes)

	// build req body
	type ReqBody struct {
		Username string
		Email string
		PasswordHash string
	}
	rb := &ReqBody{
		Username: username,
		Email: email,
		PasswordHash: passwordHash,
	}
	fmt.Printf("%+v", rb)

	// sql create
	_, err = page.store.InsertPage(r.Context(), username, email, passwordHash)
	if err != nil {
		panic(err)
	}

	// respond
	http.Redirect(w, r, "/login", http.StatusFound)
}
