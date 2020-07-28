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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type uploadResponse struct {
	URL     string `json:"url"`
	UUID    string `json:"uuid"`
	Success bool   `json:"success"`
}

func fileshareUpload(srv *Server, response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	lifetime, err := strconv.ParseInt(vars["lifetime"], 10, 64)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	request.ParseMultipartForm(32 << 20)
	file, handle, err := request.FormFile("file")
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	uuid, err := srv.storageIndex.New(lifetime, file, handle.Filename)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(response).Encode(&uploadResponse{
		URL:     fmt.Sprint(srv.Params.PublicURL, srv.Params.APIPathPrefix, "/download/", uuid),
		UUID:    uuid,
		Success: true,
	})
}

func fileshareDownload(srv *Server, response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	uuid := vars["uuid"]
	file, err := srv.storageIndex.Get(uuid)
	if err != nil {
		http.Error(response, err.Error(), http.StatusNotFound)
		return
	}
	response.Header().Set("Content-Disposition", "attachment; filename="+file.name)
	http.ServeFile(response, request, file.path)
}
