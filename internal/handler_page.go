package lakehouse

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Page struct {
	store  *SQLStore
	logger *zap.Logger
}

func NewHandlerPage(store *SQLStore) *Page {
	return &Page{
		store: store,
	}
}

func (page *Page) RenderIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/index.html",
	)
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(KeyIsAuthenticated),
		"Username":        r.Context().Value(KeyUsername),
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/dashboard.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile dashboard template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(KeyIsAuthenticated),
		"Username":        r.Context().Value(KeyUsername),
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/login.html",
	)
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
	tokenHash := c.Value

	// delete session
	err = page.store.DeleteSession(r.Context(), tokenHash)
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

	user, err := page.store.GetOneUserByUsername(r.Context(), data.Username)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(data.Password),
	)
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
	session := &Session{
		UserID:    user.ID,
		TokenHash: tokenString,
	}
	_, err = page.store.InsertSession(r.Context(), session)
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

func (page *Page) RenderNewUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/signup.html",
	)
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

func (page *Page) SaveNewUser(w http.ResponseWriter, r *http.Request) {
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
	_, err = page.store.InsertUserPage(r.Context(), username, email, passwordHash)
	if err != nil {
		panic(err)
	}

	// respond
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (page *Page) RenderOneDocument(w http.ResponseWriter, r *http.Request) {
	// parse url doc id
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get document from database
	doc, err := page.store.GetOneDocument(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page.logger.With(
			zap.Error(err),
		).Error("failed to get document")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// compile markdown to html
	unsafeHTML := blackfriday.Run([]byte(doc.Body))
	bodyHTML := bluemonday.UGCPolicy().SanitizeBytes(unsafeHTML)

	// respond
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/document.html",
	)
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(KeyIsAuthenticated),
		"Username":        r.Context().Value(KeyUsername),
		"Document":        doc,
		"BodyHTML":        template.HTML(bodyHTML),
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderAllDocument(w http.ResponseWriter, r *http.Request) {
	docs, err := page.store.GetAllDocument(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page.logger.With(
			zap.Error(err),
		).Error("failed to get all documents")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/document_list.html",
	)
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(KeyIsAuthenticated),
		"Username":        r.Context().Value(KeyUsername),
		"DocumentList":    docs,
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderNewDocument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/document_new.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile doc new template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(KeyIsAuthenticated),
		"Username":        r.Context().Value(KeyUsername),
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) SaveNewDocument(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")

	type ReqBody struct {
		Title string
		Body  string
	}
	rb := &ReqBody{
		Title: title,
		Body:  body,
	}
	fmt.Printf("%+v", rb)

	if rb.Title == "" || rb.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now()
	d := &Document{
		Title:     rb.Title,
		Body:      rb.Body,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := page.store.InsertDocument(r.Context(), d)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/docs", http.StatusFound)
}

func (page *Page) RenderEditDocument(w http.ResponseWriter, r *http.Request) {
	// parse url doc id
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get doc based on url id
	doc, err := page.store.GetOneDocument(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page.logger.With(
			zap.Error(err),
		).Error("failed to get document")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// render
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/document_edit.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile doc edit template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(KeyIsAuthenticated),
		"Username":        r.Context().Value(KeyUsername),
		"Document":        doc,
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) SaveEditDocument(w http.ResponseWriter, r *http.Request) {
	// parse doc id from url
	idAsString := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idAsString, 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// gather post form data
	var data struct {
		Title string
		Body  string
	}
	data.Title = r.FormValue("title")
	data.Body = r.FormValue("body")

	// validate data
	if data.Title == "" || data.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// write updated doc on database
	err = page.store.UpdateDocument(r.Context(), id, "title", data.Title)
	if err != nil {
		panic(err)
	}
	err = page.store.UpdateDocument(r.Context(), id, "body", data.Body)
	if err != nil {
		panic(err)
	}

	// respond
	http.Redirect(w, r, "/docs/"+idAsString, http.StatusFound)
}

func (page *Page) RenderEditor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/editor.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile dashboard template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, map[string]interface{}{
		"IsAuthenticated": r.Context().Value(KeyIsAuthenticated),
		"Username":        r.Context().Value(KeyUsername),
	})
	if err != nil {
		panic(err)
	}
}
