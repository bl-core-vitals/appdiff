package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	zipper "github.com/bukalapak/apkdiff/zipper"
)

const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"

	NewFile  = "\033[1;33m[!] %s : New File!\033[0m\n"
	Increase = "\033[1;33m[>] %s : %d => %d\033[0m\n"
	Decrease = "\033[1;34m[<] %s : %d => %d\033[0m\n"
	Same     = "\033[1;36m[=] %s : %d => %d\033[0m\n"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: apkdiff <new_apk> <old_apk>")
		return
	}

	firstApk := os.Args[1]
	secondApk := os.Args[2]

	firstDir, _ := ioutil.TempDir("", "apk")
	secondDir, _ := ioutil.TempDir("", "apk")

	err := unzip(firstApk, firstDir)
	err = unzip(secondApk, secondDir)

	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := ioutil.ReadDir(firstDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, f := range files {
		var secondDirFileName = filepath.Join(secondDir, f.Name())
		var secondFileInfo, err = os.Stat(secondDirFileName)

		if err != nil {
			fmt.Printf(NewFile, f.Name())
			continue
		}

		if f.Size() > secondFileInfo.Size() {
			fmt.Printf(Increase, f.Name(), f.Size(), secondFileInfo.Size())
		} else if f.Size() < secondFileInfo.Size() {
			fmt.Printf(Decrease, f.Name(), f.Size(), secondFileInfo.Size())
		} else {
			fmt.Printf(Same, f.Name(), f.Size(), secondFileInfo.Size())
		}
	}
}

func unzip(path string, destDir string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	_, err = zipper.Unzip(absPath, destDir)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
