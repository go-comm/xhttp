//go:build !go1.16

package xhttp

import (
	"io/ioutil"
)

var _Discard = ioutil.Discard

var _NopCloser = ioutil.NopCloser

var _ReadAll = ioutil.ReadAll

var _ReadFile = ioutil.ReadFile

var _WriteFile = ioutil.WriteFile

var _ReadDir = ioutil.ReadDir

var _CreateTemp = ioutil.TempFile

var _TempFile = ioutil.TempFile

var _MkdirTemp = ioutil.TempDir

var _TempDir = ioutil.TempDir
