package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore
var secretPassword string

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(w http.ResponseWriter, r *http.Request) {
	// fmt.Printf("%#v\n", *r)
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if password != secretPassword {
		log.Println("login: Failed login attempt")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	session, err := store.Get(r, "session")
	if err != nil {
		log.Println("login: Failed to get session")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["logged_in"] = true
	session.Values["username"] = username

	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println("logout: Failed to get session")
		http.SetCookie(w,
			&http.Cookie{
				Name: "session",
				Path: "/",
			},
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["logged_in"] = false

	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func home(w http.ResponseWriter, r *http.Request) {
	// Check for session
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println("home: Failed to get session")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	val := session.Values["logged_in"]
	var loggedIn, ok bool
	if loggedIn, ok = val.(bool); !ok || !loggedIn {
		log.Println("home: User isn't logged in")
		http.ServeFile(w, r, "login.html")
		return
	}

	val = session.Values["username"]
	var username string
	if username, ok = val.(string); !ok {
		log.Println("home: Can't find username in session")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	log.Printf("home: %q logged in", username)
	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()

	r := mux.NewRouter()

	hub := newHub()
	go hub.run()

	r.HandleFunc("/", home).Methods("GET")
	r.HandleFunc("/login", login).Methods("POST")
	r.HandleFunc("/logout", logout).Methods("GET")
	r.Handle("/ws", hub)

	var err error

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	cookieSecret := os.Getenv("COOKIE_SECRET")
	if port == "" {
		log.Fatal("$COOKIE_SECRET must be set")
	}

	secretPassword = os.Getenv("PASSWORD")
	if secretPassword == "" {
		log.Fatal("$PASSWORD must be set")
	}

	store = sessions.NewCookieStore([]byte(cookieSecret))

	log.Println("Starting server on port " + port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
