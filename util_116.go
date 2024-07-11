//go:build go1.16

package xhttp

import "io/fs"

func MustSubFS(fsys fs.FS, dir string) fs.FS {
	fsys2, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return fsys2
}
