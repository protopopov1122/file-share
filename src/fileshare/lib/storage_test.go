/*
   SPDX short identifier: MIT

   Copyright 2020 Jevgēnijs Protopopovs

   Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
   to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
   and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
   IN THE SOFTWARE.
*/

package lib

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // sqlite3 database driver
	"github.com/spf13/afero"
)

func constructMocks() (*sql.DB, afero.Fs, *Logging, error) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, nil, nil, err
	}
	fs := afero.NewMemMapFs()
	logging := NewLogging("none")
	return database, fs, logging, nil
}

const FilePath = "/tmp/files"

func makeStorage(t *testing.T) *Storage {
	db, fs, logging, err := constructMocks()
	if err != nil {
		t.Error("Failed to contruct mocks due to", err)
	}
	storage, err := NewStorage(db, fs, FilePath, logging)
	if err != nil {
		t.Error("Failed to construct storage due to", err)
	}
	return storage
}

func expectField(t *testing.T, res *sql.Rows, columnID int, expectedName, expectedType string, expectedNotNull, expectedPrimaryKey int) {
	var columnName, columnType string
	var notNull, primaryKey int
	var ignore interface{}

	if !res.Next() {
		t.Errorf("Expected at least %d columns in SharedFiles table", columnID)
	}
	err := res.Scan(&ignore, &columnName, &columnType, &notNull, &ignore, &primaryKey)
	if err != nil {
		t.Errorf("Failed to retrieve #%d column parameters due to %d", columnID, err)
	}
	if columnName != expectedName {
		t.Errorf("Expected #%d column to be '%s'; got '%s' instead", columnID, expectedName, columnName)
	}
	if columnType != expectedType {
		t.Errorf("Expected #%d column to have type '%s'; got '%s' instead", columnID, expectedType, columnType)
	}
	if notNull != expectedNotNull || primaryKey != expectedPrimaryKey {
		t.Errorf("Expected #%d column to have NOT_NULL=%d and PRIMARY_KEY=%d; got NOT_NULL=%d and PRIMARY_KEY=%d instead", columnID, expectedNotNull, expectedPrimaryKey, notNull, primaryKey)
	}
}

func testDbSchema(t *testing.T, db *sql.DB) {
	res, err := db.Query("pragma table_info('SharedFiles')")
	if err != nil {
		t.Error("Failed to query DB table info due to", err)
	}
	defer res.Close()
	expectField(t, res, 1, "uuid", "CHAR(36)", 0, 1)
	expectField(t, res, 2, "expires", "INTEGER", 1, 0)
	expectField(t, res, 3, "name", "VARCHAR(255)", 0, 0)
	if res.Next() {
		t.Error("Expected 3 columns in SharedFiles table")
	}
}

func TestNewStorage(t *testing.T) {
	storage := makeStorage(t)
	defer storage.Close()
	testDbSchema(t, storage.database)
	res, err := afero.Exists(storage.fs, FilePath)
	if err != nil {
		t.Error("Failed to test", FilePath, "existence due to", err)
	}
	if !res {
		t.Error("Expected", FilePath, "to exist")
	}
}
