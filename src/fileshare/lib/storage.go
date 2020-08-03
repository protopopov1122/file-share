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
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"

	"github.com/google/uuid"
)

// Storage manages file share storage index and contents
type Storage struct {
	database    *sql.DB
	fs          afero.Fs
	storagePath string
	log         *Logging
}

// FileDescriptor contains information regarding single file
type FileDescriptor struct {
	uuid    string
	path    string
	expires int64
	name    string
}

// NewStorage constructs new file share storage object
func NewStorage(database *sql.DB, fs afero.Fs, storagePath string, log *Logging) (*Storage, error) {
	err := fs.MkdirAll(storagePath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS SharedFiles (uuid CHAR(36) PRIMARY KEY, expires INTEGER NOT NULL, name VARCHAR(255))")
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	_, err = statement.Exec()
	if err != nil {
		return nil, err
	}
	return &Storage{
		database:    database,
		fs:          fs,
		storagePath: storagePath,
		log:         log,
	}, nil
}

// Close destroys database connection
func (index *Storage) Close() error {
	err := index.database.Close()
	if err == nil {
		index.database = nil
	}
	return err
}

func (index *Storage) uploadNew(lifetime int64, source io.Reader, name string) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC().Unix()
	expiry := now + lifetime
	statement, err := index.database.Prepare("INSERT INTO SharedFiles VALUES (?, ?, ?)")
	if err != nil {
		return "", err
	}
	defer statement.Close()
	_, err = statement.Exec(id.String(), expiry, name)
	if err != nil {
		return "", err
	}
	destination, err := index.fs.OpenFile(filepath.Join(index.storagePath, id.String()), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}
	defer destination.Close()
	io.Copy(destination, source)
	return id.String(), nil
}

// New creates an index database entry and saves file
func (index *Storage) New(lifetime int64, source io.Reader, name string) (string, error) {
	id, err := index.uploadNew(lifetime, source, name)
	if err != nil {
		index.log.Warning.Println("Uploading file", name, "failed due to", err)
	} else {
		index.log.Info.Println("Saved file", name, "as", id, "with lifetime", lifetime, "second(s)")
	}
	return id, err
}

// Get searches file by UUID and returns a file descriptor
func (index *Storage) Get(id string) (*FileDescriptor, error) {
	res, err := index.database.Query("SELECT expires, name FROM SharedFiles WHERE uuid = ?", id)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if res.Next() {
		result := &FileDescriptor{
			uuid: id,
			path: filepath.Join(index.storagePath, id),
		}
		res.Scan(&result.expires, &result.name)
		return result, nil
	}
	return nil, os.ErrNotExist
}

// Count returns a number of database records
func (index *Storage) Count() (int, error) {
	res, err := index.database.Query("SELECT COUNT(*) FROM SharedFiles")
	if err != nil {
		index.log.Warning.Println("Failed to retrieve database record count due to", err)
		return -1, err
	}
	defer res.Close()
	res.Next()
	var count int = -1
	err = res.Scan(&count)
	if err != nil {
		index.log.Warning.Println("Failed to retrieve database record count due to", err)
		return -1, err
	}
	return count, nil
}

func (index *Storage) dropOldEntries() (int, int, error) {
	initialCount, _ := index.Count()
	now := time.Now().UTC().Unix()
	tx, err := index.database.Begin()
	if err != nil {
		return -1, -1, err
	}
	res, err := tx.Prepare("DELETE FROM SharedFiles WHERE expires < ?")
	if err != nil {
		tx.Rollback()
		return -1, -1, err
	}
	defer res.Close()
	_, err = res.Exec(now)
	if err != nil {
		tx.Rollback()
		return -1, -1, err
	}
	err = tx.Commit()
	if err != nil {
		return -1, -1, err
	}
	finalCount, _ := index.Count()
	return initialCount, finalCount, err
}

func (index *Storage) dropObsoleteFiles() (uint, error) {
	var removalCounter uint = 0
	err := afero.Walk(index.fs, index.storagePath, func(path string, info os.FileInfo, err error) error {
		if path == index.storagePath {
			return nil
		}
		res, err := index.database.Query("SELECT uuid FROM SharedFiles WHERE uuid = ?", info.Name())
		if err != nil {
			return err
		}
		defer res.Close()
		if !res.Next() {
			err = index.fs.Remove(path)
			if err == nil {
				removalCounter++
			}
		}
		return err
	})
	return removalCounter, err
}

// CollectGarbage removes obsolete database records and cleans up storage
func (index *Storage) CollectGarbage() error {
	index.log.Debug.Println("Starting database cleanup")
	initialCount, finalCount, err := index.dropOldEntries()
	if err != nil {
		index.log.Warning.Println("Database cleanup failed due to", err)
	}
	if initialCount > 0 && finalCount >= 0 && finalCount < initialCount {
		index.log.Info.Println("Database cleanup finished;", initialCount-finalCount, "record(s) removed")
	} else {
		index.log.Info.Println("Database cleanup finished")
	}
	index.log.Debug.Println("Starting obsolete file removal")
	removalCounter, err := index.dropObsoleteFiles()
	if err != nil {
		index.log.Warning.Println("Obsolete file removal failed due to", err)
	}
	index.log.Info.Println("Removed", removalCounter, "obsolete file(s)")
	return err
}
