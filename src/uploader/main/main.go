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
	"fmt"
	"os"
	"strconv"

	uploaderLib "github.com/protopopov1122/file-share/src/uploader/lib"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Misconfigured: provide upload parameters")
		os.Exit(1)
	}
	lifetime, err := strconv.ParseUint(os.Args[2], 10, 32)
	if err != nil {
		fmt.Println("Failed to parse lifetime parameter due to", err)
		os.Exit(1)
	}
	uploader := uploaderLib.Uploader{
		APIURL:   os.Args[1],
		Lifetime: uint(lifetime),
	}
	res, err := uploader.Upload(os.Stdin, "file")
	if err != nil {
		fmt.Println("Failed to upload file due to ", err)
	} else {
		fmt.Println("Uploaded file available at", res, "for", uploader.Lifetime, "second(s)")
	}
}
