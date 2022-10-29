package user

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strings"

	"git.sr.ht/~sirodoht/lakehousewiki/session"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Page struct {
	store        Store
	sessionStore session.Store
	logger       *zap.Logger
}

func NewPage(store Store, sessionStore session.Store) *Page {
	return &Page{
		store:        store,
		sessionStore: sessionStore,
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

func (page *Page) DeleteSession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session")
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusFound)
	}
	token_hash := c.Value

	// delete session
	err = page.sessionStore.Delete(r.Context(), token_hash)
	if err != nil {
		fmt.Println(err)
	}

	// delete cookie by setting a new one with same name and max age < 0
	cookie := http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	// redirect to index
	http.Redirect(w, r, "/", http.StatusFound)
}

func (page *Page) CreateSession(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Username string
		Password string
	}
	data.Username = r.FormValue("username")
	data.Password = r.FormValue("password")
	fmt.Printf("%+v", data)

	user, err := page.store.GetOneByUsername(r.Context(), data.Username)
	fmt.Printf("%+v", user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Checking password=%+v and hash=%+v", data.Password, user.PasswordHash)
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.Password))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	// create session token
	tokenBytes := make([]byte, 32)
	nRead, err := rand.Read(tokenBytes)
	if err != nil {
		panic(fmt.Errorf("session: %w", err))
	}
	if nRead < 32 {
		panic(fmt.Errorf("session: not enough random bytes"))
	}
	tokenHash := sha256.Sum256(tokenBytes)
	tokenString := base64.URLEncoding.EncodeToString(tokenHash[:])
	session := &session.Session{
		UserID:    user.ID,
		TokenHash: tokenString,
	}
	_, err = page.sessionStore.Insert(r.Context(), session)
	if err != nil {
		panic(err)
	}

	// set cookie with session token
	cookie := http.Cookie{
		Name:     "session",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	// respond
	http.Redirect(w, r, "/", http.StatusFound)
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
		Username     string
		Email        string
		PasswordHash string
	}
	rb := &ReqBody{
		Username:     username,
		Email:        email,
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
