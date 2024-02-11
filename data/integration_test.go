//go:build integration

// run tests with this command: go test . --tags integration --count=1
package data

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	port     = "5435"
	user     = "postgres"
	password = "password"
	dbName   = "celeritas_test"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var dummyUser = User{
	FirstName: "John",
	LastName:  "Smith",
	Email:     "john.smith@test.com",
	Active:    1,
	Password:  "password",
}

var models Models
var testDB *sql.DB
var resource *dockertest.Resource
var pool *dockertest.Pool

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_TYPE", "postgres")
	os.Setenv("UPPER_DB_LOG", "ERROR")

	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	pool = p

	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to docker: %s", err)
	}

	if err = createTables(testDB); err != nil {
		log.Fatalf("error creating tables: %s", err)
	}

	models = New(testDB)

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables(db *sql.DB) error {
	file, err := os.ReadFile("create-test-tables.sql")
	if err != nil {
		return err
	}

	stmt := string(file)

	_, err = db.Exec(stmt)
	return err
}

func TestUser_Table(t *testing.T) {
	s := models.Users.Table()
	if s != "users" {
		t.Error("wrong table name returned", s)
	}
}

func TestUser_Insert(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("failed to insert new user record:", err)
	}

	if id == 0 {
		t.Error("0 returned as id after insert")
	}
}

func TestUser_Get(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	if u.ID == 0 {
		t.Error("id of retured user is 0:", err)
	}
}

func TestUser_GetAll(t *testing.T) {
	u, err := models.Users.GetAll()
	if err != nil {
		t.Error("failed to get users:", err)
	}

	if len(u) == 0 {
		t.Error("zero user records returned", err)
	}
}

func TestUser_GetByEmail(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	if u.ID == 0 {
		t.Error("id of retured user is 0:", err)
	}
}

func TestUser_Update(t *testing.T) {
	newLastName := "Test"

	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	u.LastName = newLastName
	err = u.Update(*u)
	if err != nil {
		t.Error("failed to update user:", err)
	}

	u, err = models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	if u.LastName != newLastName {
		t.Error("last name failed to update")
	}
}

func TestUser_PasswordMatches(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	matches, err := u.PasswordMatches(dummyUser.Password)
	if err != nil {
		t.Error("failed to validate password:", err)
	}

	if !matches {
		t.Error("password does not match, expected a match")
	}

	matches, err = u.PasswordMatches("wrongpassword")
	if matches {
		t.Error("used incorrect password, expected no match, got a match")
	}

	matches, err = u.PasswordMatches("")
	if matches {
		t.Error("used empty password, expected no match, got a match")
	}
}
func TestUser_ResetPassword(t *testing.T) {
	newPassword := "newpassword"
	if err := models.Users.ResetPassword(1, newPassword); err != nil {
		t.Error("failed to reset password:", err)
	}

	if err := models.Users.ResetPassword(2, newPassword); err == nil {
		t.Error("resetting password for non-existent user, expected an error, received none")
	}
}

func TestUser_Delete(t *testing.T) {
	if err := models.Users.Delete(1); err != nil {
		t.Error("failed to delete user:", err)
	}

	_, err := models.Users.Get(1)
	if err == nil {
		t.Error("trying to retrieve record of deleted user, expected error, received none")
	}
}

func TestToken_Table(t *testing.T) {
	s := models.Tokens.Table()
	if s != "tokens" {
		t.Error("wrong table name returned for token")
	}
}

func TestToken_GenerateToken(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("failed to insert new user record:", err)
	}

	_, err = models.Tokens.GenerateToken(id, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token:", err)
	}
}

func TestToken_Insert(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	token, err := models.Tokens.GenerateToken(u.ID, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token: ", err)
	}

	err = models.Tokens.Insert(*token, *u)
	if err != nil {
		t.Error("error insering token: ", err)
	}
}

func TestToken_GetUserForToken(t *testing.T) {
	token := "abc"

	if _, err := models.Tokens.GetUserForToken(token); err == nil {
		t.Error("search with invalid token, error expected, none received", err)
	}

	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	if _, err := models.Tokens.GetUserForToken(u.Token.PlainText); err != nil {
		t.Error("using a valid token to search for a user, error received:", err)
	}

}

func TestToken_GetTokensForUser(t *testing.T) {

	tokens, err := models.Tokens.GetTokensForUser(1)
	if err != nil {
		t.Error("failed to get tokens for user:", err)
	}

	if len(tokens) > 0 {
		t.Error("tokens returned for non-existent user")
	}
}

func TestToken_Get(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	if _, err := models.Tokens.Get(u.Token.ID); err != nil {
		t.Error("failed to get token by id:", err)
	}

	if _, err := models.Tokens.Get(0); err == nil {
		t.Error("used invalid id, expected an error, received none")
	}
}

func TestToken_GetByToken(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	if _, err := models.Tokens.GetByToken(u.Token.PlainText); err != nil {
		t.Error("failed to get token data by token:", err)
	}

	if _, err := models.Tokens.GetByToken("invalidtoken"); err == nil {
		t.Error("attempted to get token data using invalid token, expected error, received none")
	}
}

var authData = []struct {
	name        string
	token       string
	email       string
	errExpected bool
	message     string
}{
	{"invalid_token", "abcdefghijklmnopqrstuvwxyz", "not.exist@test.com", true, "invalid token accepted as valid"},
	{"invalid_length", "abcdefghijklmnopqrstuvwxy", "not.exist@test.com", true, "invalid token with invalid length accepted as valid"},
	{"no_user", "abcdefghijklmnopqrstuvwxyz", "not.exist@test.com", true, "invalid token with invalid length accepted as valid"},
	{"valid", "", dummyUser.Email, false, "valid token reported as invalid"},
}

func TestToken_AuthenticateToken(t *testing.T) {
	for _, tt := range authData {
		token := ""
		if tt.email == dummyUser.Email {
			user, err := models.Users.GetByEmail(tt.email)
			if err != nil {
				t.Error("failed to get user:", err)
			}
			token = user.Token.PlainText
		} else {
			token = tt.token
		}

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer "+token)

		_, err := models.Tokens.AuthenticateToken(req)
		if tt.errExpected && err == nil {
			t.Errorf("%s: %s", tt.name, tt.message)
		} else {
			t.Logf("passed %s", tt.name)
		}
	}
}

func TestToken_Delete(t *testing.T) {
	user, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	if err := models.Tokens.DeleteByToken(user.Token.PlainText); err != nil {
		t.Error("error deleting token:", err)
	}
}

func TestToken_ExpiredToken(t *testing.T) {
	user, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	token, err := models.Tokens.GenerateToken(user.ID, -time.Hour*24)
	if err != nil {
		t.Error("error generating token: ", err)
	}

	err = models.Tokens.Insert(*token, *user)
	if err != nil {
		t.Error("error insering token: ", err)
	}

	ok, err := models.Tokens.ValidToken(token.PlainText)
	if ok || err == nil {
		t.Error("using expired token, passed validation, expected to fail")
	}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer "+token.PlainText)

	_, err = models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("using expired token, expected error, received none")
	}
}

func TestToken_BadHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	_, err := models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("sending request without auth header, expected error, received none")
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "invalid")

	_, err = models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("using invalid auth header, expected error, received none")
	}

	newUser := User{
		FirstName: "temp",
		LastName:  "temp",
		Email:     "temp@test.com",
		Active:    1,
		Password:  "temp",
	}

	id, err := models.Users.Insert(newUser)
	if err != nil {
		t.Error("error inserting new user:", err)
	}

	token, err := models.Tokens.GenerateToken(id, 1*time.Hour)
	if err != nil {
		t.Error("error generating token: ", err)
	}

	err = models.Tokens.Insert(*token, newUser)
	if err != nil {
		t.Error("error inserting token:", err)
	}

	err = models.Users.Delete(id)

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer "+token.PlainText)

	_, err = models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("using token from deleted user, expected error, received none")
	}
}

func TestToken_DeleteNonExistentToken(t *testing.T) {
	if err := models.Tokens.DeleteByToken("invalidtoken"); err != nil {
		t.Error("error deleting token:", err)
	}
}

func TestToken_ValidToken(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	newToken, err := models.Tokens.GenerateToken(u.ID, 1*time.Hour)
	if err != nil {
		t.Error("error generating token: ", err)
	}

	err = models.Tokens.Insert(*newToken, *u)
	if err != nil {
		t.Error("error inserting token:", err)
	}

	ok, err := models.Tokens.ValidToken(newToken.PlainText)
	if err != nil {
		t.Error("error validating token:", err)
	}

	if !ok {
		t.Error("using valid token, failed validation, expected to pass")
	}

	ok, err = models.Tokens.ValidToken("invalidtoken")
	if ok || err == nil {
		t.Error("using invalid token, passed validation, expected to fail")
	}

	u, err = models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	if err = models.Tokens.Delete(u.Token.ID); err != nil {
		t.Error("failed to delete token:", err)
	}

	ok, err = models.Tokens.ValidToken(u.Token.PlainText)
	if ok || err == nil {
		t.Error("using deleted token, passed validation, expected to fail")
	}
}
