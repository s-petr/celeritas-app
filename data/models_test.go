package data

import (
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	db2 "github.com/upper/db/v4"
)

func TestNew(t *testing.T) {
	mockDB, _, _ := sqlmock.New()
	defer mockDB.Close()

	_ = os.Setenv("DATABASE_TYPE", "postgres")
	m := New(mockDB)
	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Error("wrong type returned", fmt.Sprintf("%T", m))

	}

	_ = os.Setenv("DATABASE_TYPE", "mysql")
	m = New(mockDB)
	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Error("wrong type returned", fmt.Sprintf("%T", m))

	}
}

func TestGetInsertID(t *testing.T) {
	var id db2.ID = int64(1)

	returnedID := getInsertID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Error("wrong type returned", fmt.Sprintf("%T", returnedID))
	}

	id = 1
	returnedID = getInsertID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Error("wrong type returned", fmt.Sprintf("%T", returnedID))
	}
}
