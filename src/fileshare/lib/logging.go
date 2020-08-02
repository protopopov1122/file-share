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
	"io/ioutil"
	"log"
	"os"
)

// Logging contains loggers for different verbosity levels
type Logging struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
}

// LogLevel defines possible logging levels
type LogLevel string

const (
	// Debug includes everything
	Debug = "debug"
	// Info includes everything up to informational messages
	Info = "info"
	// Warning includes only warnings and errors
	Warning = "warning"
	// None suppresses all logging
	None = "none"
)

// NewLogging constructs loggers according to specified level
func NewLogging(level LogLevel) *Logging {
	flags := log.Ltime | log.Lshortfile | log.Ldate
	logging := &Logging{
		Debug:   log.New(ioutil.Discard, " [ debug ] ", flags),
		Info:    log.New(ioutil.Discard, " [ info  ] ", flags),
		Warning: log.New(ioutil.Discard, " [warning] ", flags),
	}
	switch level {
	case Debug:
		logging.Debug.SetOutput(os.Stdout)
		fallthrough
	case Info:
		logging.Info.SetOutput(os.Stdout)
		fallthrough
	case Warning:
		logging.Warning.SetOutput(os.Stdout)
	}
	return logging
}
