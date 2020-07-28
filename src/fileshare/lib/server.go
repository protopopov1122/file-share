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
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jasonlvhit/gocron"
)

// Params structure contains global parameters for file share
type Params struct {
	Address                   string   `json:"bindTo"`
	PublicURL                 string   `json:"publicUrl"`
	APIPathPrefix             string   `json:"apiPrefix"`
	Storage                   string   `json:"storage"`
	GarbageCollectionInterval uint64   `json:"gcInterval"`
	LoggingLevel              LogLevel `json:"loglevel"`
}

// Server structure contains all file share related objects
type Server struct {
	handler      *mux.Router
	server       http.Server
	storageIndex *Storage
	Params       *Params
	Log          *Logging
	scheduler    *gocron.Scheduler
}

// NewServer constructs new file share server
func NewServer(params *Params) (*Server, error) {
	logging := NewLogging(params.LoggingLevel)
	logging.Debug.Println("File share server initialization: serving", params.Storage, "on", params.Address)
	storageIndex, err := NewStorage(params.Storage, logging)
	if err != nil {
		logging.Warning.Println("Failed to initialize storage due to", err)
		return nil, err
	}
	handler := mux.NewRouter()
	srv := &Server{
		handler: handler,
		server: http.Server{
			Addr:    params.Address,
			Handler: handler,
		},
		storageIndex: storageIndex,
		Params:       params,
		Log:          logging,
		scheduler:    gocron.NewScheduler(),
	}

	return srv, nil
}

// Setup performs initial server setup and cleanup
func (srv *Server) Setup() {
	srv.handleFunc("PUT", "/upload/{lifetime:[0-9]+}", fileshareUpload)
	srv.handleFunc("GET", "/download/{uuid:[0-9a-z-]+}", fileshareDownload)
}

// Start file share server
func (srv *Server) Start() error {
	srv.Log.Info.Println("File share server starting: serving", srv.Params.Storage, "on", srv.Params.Address)
	err := srv.storageIndex.CollectGarbage()
	if err != nil {
		srv.Log.Warning.Println("File share server startup failed due to", err)
	}
	srv.scheduler.Every(srv.Params.GarbageCollectionInterval).Seconds().Do(srv.storageIndex.CollectGarbage)
	srv.scheduler.Start()
	return srv.server.ListenAndServe()
}

func (srv *Server) fullPath(path string) string {
	return fmt.Sprint(srv.Params.APIPathPrefix, path)
}

func (srv *Server) handleFunc(method string, path string, handle func(*Server, http.ResponseWriter, *http.Request)) {
	srv.handler.HandleFunc(srv.fullPath(path), func(response http.ResponseWriter, request *http.Request) {
		handle(srv, response, request)
	}).Methods(method)
}

// Close and stop file share server
func (srv *Server) Close() {
	srv.scheduler.Clear()
	srv.server.Close()
	srv.storageIndex.Close()
}
