//go:build go1.16

package xhttp

import (
	"io"
	"os"
)

var _Discard = io.Discard

var _NopCloser = io.NopCloser

var _ReadAll = io.ReadAll

var _ReadFile = os.ReadFile

var _WriteFile = os.WriteFile

var _ReadDir = os.ReadDir

var _CreateTemp = os.CreateTemp

var _TempFile = os.CreateTemp

var _MkdirTemp = os.MkdirTemp

var _TempDir = os.MkdirTemp
