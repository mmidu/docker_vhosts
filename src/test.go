package main

import (
	"path/filepath"
	"os"
)

func main() {

	filepath.Walk("/", func(path string, info os.FileInfo, err error) error {
    		if err == nil && info.Name() == "mod_proxy.so" {
        		println(info.Name())
    		}
    		return nil
	})
}
