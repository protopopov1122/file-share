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
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // sqlite3 database driver
	"github.com/spf13/afero"
)

func constructMocks() (*sql.DB, afero.Fs, *Logging, error) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, nil, nil, err
	}
	fs := afero.NewMemMapFs()
	logging := NewLogging("none", ioutil.Discard)
	return database, fs, logging, nil
}

const FilePath = "/tmp/files"

type fixedTime struct {
	Value int64
}

func (ftime fixedTime) UTCNow() int64 {
	return ftime.Value
}

func makeStorage(t *testing.T, time Time) *Storage {
	db, fs, logging, err := constructMocks()
	if err != nil {
		t.Error("Failed to contruct mocks due to", err)
	}
	storage, err := NewStorage(db, fs, FilePath, time, logging)
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
	time := fixedTime{
		Value: 0,
	}
	storage := makeStorage(t, &time)
	defer storage.Close()
	testDbSchema(t, storage.database)
	res, err := afero.Exists(storage.fs, FilePath)
	if err != nil {
		t.Error("Failed to test", FilePath, "existence due to", err)
	}
	if !res {
		t.Error("Expected", FilePath, "to exist")
	}
	count, err := storage.Count()
	if err != nil {
		t.Error("Failed to retrieve storage size due to", err)
	}
	if count > 0 {
		t.Errorf("Expected storage to be empty; got %d entires", count)
	}
}

func TestNonExistentFile(t *testing.T) {
	time := fixedTime{
		Value: 0,
	}
	storage := makeStorage(t, &time)
	defer storage.Close()
	res, err := storage.Get("SomeUUID")
	if err != os.ErrNotExist {
		t.Errorf("Expected err=%s, res=nil; got err=%s, res=%p", os.ErrNotExist, err, res)
	}
}

func TestNewFileUpload(t *testing.T) {
	time := fixedTime{
		Value: 10,
	}
	storage := makeStorage(t, &time)
	defer storage.Close()
	content := []byte{
		0, 1, 2, 3, 5, 8,
		13, 21, 34, 56,
	}
	uuidVal, err := storage.New(120, bytes.NewReader(content), "file1")
	if err != nil {
		t.Error("Failed to upload file due to", err)
	}
	_, err = uuid.Parse(uuidVal)
	if err != nil {
		t.Errorf("Failed to parse returned uuid=%s", uuidVal)
	}
	count, err := storage.Count()
	if err != nil {
		t.Error("Failed to retrieve entry count due to", err)
	}
	if count != 1 {
		t.Errorf("Expected count=1; got count=%d", count)
	}
	file, err := storage.Get(uuidVal)
	if err != nil {
		t.Error("Failed to retrieve file due to", err)
	}
	if file.name != "file1" {
		t.Errorf("Expected name='file1'; got='%s'", file.name)
	}
	if file.uuid != uuidVal {
		t.Errorf("Expected uuid='%s'; got uuid='%s'", uuidVal, file.uuid)
	}
	expectedPath := filepath.Join(FilePath, uuidVal)
	if file.path != expectedPath {
		t.Errorf("Expected path='%s'; got path='%s'", expectedPath, file.path)
	}
	exists, err := afero.Exists(storage.fs, expectedPath)
	if err != nil {
		t.Errorf("Failed to check whether '%s' exists due to %s", expectedPath, err)
	}
	if !exists {
		t.Errorf("Expected '%s' to exist", expectedPath)
	}
	if file.expires != 130 {
		t.Errorf("Expected expiry time=%d to be equal to %d", file.expires, 130)
	}
	fileReader, err := storage.fs.OpenFile(file.path, os.O_RDONLY, 0)
	if err != nil {
		t.Error("Failed to open file due to", err)
	}
	defer fileReader.Close()
	fileContent, err := ioutil.ReadAll(fileReader)
	if err != nil {
		t.Error("Failed to read file content due to", err)
	}
	if !reflect.DeepEqual(content, fileContent) {
		t.Error("File content does not match to expected")
	}
}

func TestMultipleFileUpload(t *testing.T) {
	time := fixedTime{
		Value: 0,
	}
	storage := makeStorage(t, &time)
	defer storage.Close()
	contents := [][]byte{
		{1, 2, 3, 4},
		{},
		{200, 100, 64, 65, 88, 0, 1, 5},
		{5},
	}
	uuids := []string{}
	for idx, content := range contents {
		time.Value = int64(idx)
		uuid, err := storage.New(60, bytes.NewReader(content), fmt.Sprintf("file%d", idx))
		if err != nil {
			t.Errorf("Failed to upload file #%d due to %s", idx, err)
		}
		uuids = append(uuids, uuid)
	}
	count, err := storage.Count()
	if err != nil {
		t.Error("Failed to retrieve file count due to", err)
	}
	if count != len(contents) {
		t.Errorf("Expected count=%d; got count=%d", len(contents), count)
	}
	for idx, uuid := range uuids {
		file, err := storage.Get(uuid)
		if err != nil {
			t.Errorf("Failed to get %s due to %s", uuid, err)
		}
		if file.expires != int64(idx)+60 {
			t.Errorf("Expected expiry time=%d to be equal to %d", file.expires, idx+60)
		}
		fileReader, err := storage.fs.OpenFile(file.path, os.O_RDONLY, 0)
		if err != nil {
			t.Errorf("Failed to open file %s due to %s", file.path, err)
		}
		defer fileReader.Close()
		fileContent, err := ioutil.ReadAll(fileReader)
		if err != nil {
			t.Errorf("Failed to read file %s content due to %s", file.path, err)
		}
		if !reflect.DeepEqual(contents[idx], fileContent) {
			t.Errorf("File %s content does not match to expected", file.path)
		}
	}
}

func TestGarbageCollection(t *testing.T) {
	time := fixedTime{
		Value: 0,
	}
	storage := makeStorage(t, &time)
	defer storage.Close()
	contents := [][]byte{
		{1, 2, 3, 4},
		{},
		{200, 100, 64, 65, 88, 0, 1, 5},
		{5},
	}
	uuids := []string{}
	for idx, content := range contents {
		time.Value = int64(idx)
		uuid, err := storage.New(int64(idx*10)+10, bytes.NewReader(content), fmt.Sprintf("file%d", idx))
		if err != nil {
			t.Errorf("Failed to upload file #%d due to %s", idx, err)
		}
		uuids = append(uuids, uuid)
	}
	count, err := storage.Count()
	if err != nil {
		t.Error("Failed to retrieve file count due to", err)
	}
	if count != len(contents) {
		t.Errorf("Expected count=%d; got count=%d", len(contents), count)
	}
	err = storage.CollectGarbage()
	if err != nil {
		t.Error("Failed to collect garbage due to", err)
	}
	for idx, uuid := range uuids {
		_, err := storage.Get(uuid)
		if err != nil {
			t.Errorf("Failed to locate %s due to %s", uuid, err)
		}
		time.Value = int64(idx*10 + 15)
		err = storage.CollectGarbage()
		if err != nil {
			t.Error("Failed to collect garbage due to", err)
		}
		_, err = storage.Get(uuid)
		if err != os.ErrNotExist {
			t.Errorf("Expected %s to not exist; got %s instead", uuid, err)
		}
		count, err := storage.Count()
		if err != nil {
			t.Errorf("Failed to get entry count due to %s", err)
		}
		expectedCount := len(uuids) - idx - 1
		if count != expectedCount {
			t.Errorf("Expected count=%d; got count=%d instead", expectedCount, count)
		}
	}
}
