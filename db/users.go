package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtKey = "JWT_KEY"
)

// User represents a user model that maps to the database users table.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type claims struct {
	Username string `json:"username"`
	ID       string `json:"id"`
	jwt.StandardClaims
}

// FindUser takes an email/username as an identifier and returns the associated user
func (user *User) FindUser(db *sql.DB, identifier string) error {
	query := `SELECT * FROM users WHERE username = $1 OR email = $1`
	row := db.QueryRow(query, identifier)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return err
	}
	return nil
}

// Authenticate returns true or false
func (user *User) Authenticate(db *sql.DB, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false
	}
	return true
}

// CreateJwtToken from a user
func (user *User) CreateJwtToken() (string, error) {
	key, ok := os.LookupEnv(jwtKey)
	if !ok {
		key = "secret-key"
	}

	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expirationTime := time.Now().Add(5 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claim := &claims{
		ID:       user.ID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	// Create the JWT string and return it
	return token.SignedString([]byte(key))
}

// CreateUser does exactly what you'd think
func (user *User) CreateUser(db *sql.DB) error {
	q1 := `SELECT username FROM users WHERE username = $1 OR email = $2`
	q2 := `INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)`
	var err error

	// res1, err := db.Query(q1, user.Username, user.Email)
	res1 := db.QueryRow(q1, user.Username, user.Email)
	err = res1.Scan(nil)
	if err != sql.ErrNoRows {
		return sql.ErrNoRows
	}

	userID, err := uuid.NewRandom()
	if err != nil {
		return errors.New("create user: error creating uuid")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		// fmt.Println(err)
		return errors.New("create user: error generating password")
	}

	res2, err := db.Exec(q2, userID, user.Username, user.Email, hash)
	if err != nil {
		fmt.Println(err)
		return err
	}
	rows, err := res2.RowsAffected()
	if rows == 0 || err != nil {
		// log.Fatal("User could not be created")
		return err
	}
	user.ID = userID.String()
	return nil
}

// func GetAllUsers() ([]*User, error) {
// 	users := make([]*User, 0)
// 	query := `SELECT * FROM users`

// 	rows, err := db.Query(query)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for rows.Next() {
// 		var user = new(User)
// 		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
// 		if err != nil {
// 			return nil, err
// 		}
// 		users = append(users, user)
// 	}

// 	return users, nil
// }
