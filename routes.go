package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"github.com/amaansaeed/go-user-service/db"
)

type credentials struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (a *app) getUser(w http.ResponseWriter, r *http.Request) {
	identifier := "calvin33"
	var user = db.User{}

	err := user.FindUser(a.DB, identifier)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(user.Email))
}

func (a *app) getAllUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Not yet implemented"))
}

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	var err error
	var user = db.User{}

	err = json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Issue with incoming JSON"))
		return
	}
	if creds.Identifier == "" || creds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incomplete credentials"))
		return
	}

	err = user.FindUser(a.DB, creds.Identifier)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("login attempt failed"))
		return
	}

	ok := user.Authenticate(a.DB, creds.Password)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("login attempt failed"))
		return
	}

	token, err := user.CreateJwtToken()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Token", token)
	w.WriteHeader(http.StatusOK)
}

func (a *app) createUser(w http.ResponseWriter, r *http.Request) {
	var user = db.User{}
	var err error
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	if user.Username == "" || user.Email == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("incomplete user details"))
		return
	}

	err = user.CreateUser(a.DB)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("username/email already exists"))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("could not create user"))
		return
	}

	token, err := user.CreateJwtToken()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Token", token)
	w.WriteHeader(http.StatusOK)
}
