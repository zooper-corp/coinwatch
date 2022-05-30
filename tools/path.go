package tools

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func ExpandPath(path string) string {
	usr, err := user.Current()
	dir := "/"
	if err == nil {
		dir = usr.HomeDir
	}
	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}
	return path
}

func PathExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
