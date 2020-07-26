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

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/protopopov1122/fileshare/src/fileshare"
)

func loadConfig(params *fileshare.Params, config string) error {
	jsonFile, err := os.Open(config)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	jsonContent, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonContent, params)
}

func main() {
	params := &fileshare.Params{
		Address:                   ":8080",
		PublicURL:                 "http://localhost:8080",
		APIPathPrefix:             "/file-share-v1",
		Storage:                   "./storage",
		LoggingLevel:              "debug",
		GarbageCollectionInterval: 60,
	}
	if len(os.Args) > 1 {
		err := loadConfig(params, os.Args[1])
		if err != nil {
			log.Fatal("Failed to load configuration due to ", err)
		}
		log.Println("Loaded configuration from", os.Args[1])
	}
	srv, err := fileshare.NewServer(params)
	if err != nil {
		log.Fatal("Failed to initialize file share server due to ", err)
		return
	}
	defer srv.Close()
	srv.Setup()
	srv.Log.Warning.Fatal("Failed to start file share server due to ", srv.Start())
}
