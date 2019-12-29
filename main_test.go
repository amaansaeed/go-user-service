package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var router = mux.NewRouter()

var a app

func TestMain(m *testing.M) {
	godotenv.Load()
	a = app{}
	a.Initialize(os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		fmt.Sprintf("%s_test", os.Getenv("DB_NAME")))
	createTable()
	seedTestData()
	code := m.Run()
	destroyTable()
	os.Exit(code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func prepPayload(p string) io.Reader {
	return bytes.NewBuffer([]byte(p))
}

func createTable() {
	query := `CREATE TABLE IF NOT EXISTS public.users 
	(
		id uuid NOT NULL,
		username character varying(20) NOT NULL,
		email character varying(20) NOT NULL,
		password character varying(80) NOT NULL
	)`
	if _, err := a.DB.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func seedTestData() {
	query := `INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)`

	userID, _ := uuid.NewRandom()
	hash, _ := bcrypt.GenerateFromPassword([]byte("birchtree"), 8)

	if _, err := a.DB.Exec(query, userID, "hobbes", "hobbes@email.com", hash); err != nil {
		log.Fatal(err)
	}
}

func destroyTable() {
	a.DB.Exec("DROP TABLE IF EXISTS public.users")
}

func TestHealthCheck(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "server healthy" {
		t.Errorf("Expected 'server healthy'. Got %s", body)
	}
}

func TestCreateUser(t *testing.T) {
	pPass := `{ "username": "bob", "email": "bob@email.com", "password": "password"}`
	pFail1 := `{ "email": "bob@email.com", "password": "password"}`
	pFail2 := `{ "username": "hobbes", "email": "hobbes@email.com", "password": "birchtree"}`

	req1, _ := http.NewRequest("POST", "/api/v1/create", prepPayload(pPass))
	res1 := executeRequest(req1)

	checkResponseCode(t, http.StatusOK, res1.Code)

	if header := res1.Header().Get("Access-Token"); len(header) < 5 {
		t.Errorf("Expected 'access token'. Got nada")
	}

	req2, _ := http.NewRequest("POST", "/api/v1/create", prepPayload(pFail1))
	res2 := executeRequest(req2)

	checkResponseCode(t, http.StatusBadRequest, res2.Code)

	if body := res2.Body.String(); body != "incomplete user details" {
		t.Errorf("Expected 'incomplete user details'. Got %s", body)
	}

	req3, _ := http.NewRequest("POST", "/api/v1/create", prepPayload(pFail2))
	res3 := executeRequest(req3)

	checkResponseCode(t, http.StatusBadRequest, res3.Code)

	if body := res3.Body.String(); body != "username/email already exists" {
		t.Errorf("Expected 'username/email already exists'. Got %s", body)
	}
}

func TestLogin(t *testing.T) {
	pPass := `{ "identifier": "hobbes", "password": "birchtree"}`
	pFail1 := `{ "identifier": "hobbes", "password": "birchree"}`
	pFail2 := `{ "identifier": "hobes", "password": "birchtree"}`

	req1, _ := http.NewRequest("POST", "/api/v1/login", prepPayload(pPass))
	res1 := executeRequest(req1)

	checkResponseCode(t, http.StatusOK, res1.Code)

	if header := res1.Header().Get("Access-Token"); len(header) < 5 {
		t.Errorf("Expected 'access token'. Got nada")
	}

	req2, _ := http.NewRequest("POST", "/api/v1/login", prepPayload(pFail1))
	res2 := executeRequest(req2)

	checkResponseCode(t, http.StatusBadRequest, res2.Code)

	if body := res2.Body.String(); body != "login attempt failed" {
		t.Errorf("Expected 'login attempt failed'. %s", body)
	}

	req3, _ := http.NewRequest("POST", "/api/v1/login", prepPayload(pFail2))
	res3 := executeRequest(req3)

	checkResponseCode(t, http.StatusBadRequest, res3.Code)

	if body := res2.Body.String(); body != "login attempt failed" {
		t.Errorf("Expected 'login attempt failed'. %s", body)
	}
}
