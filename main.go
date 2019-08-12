package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/atotto/clipboard"

	"github.com/esafirm/appdiff/zipper"
)

const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"

	NewFile     = "\033[1;33m[!] %s : New File!\033[0m\n"
	RemovedFile = "\033[1;33m[!] %s : Removed File!\033[0m\n"
	Increase    = "\033[1;33m[>] %s : %d => %d\033[0m\n"
	Decrease    = "\033[1;34m[<] %s : %d => %d\033[0m\n"
	Same        = "\033[1;36m[=] %s : %d => %d\033[0m\n"
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

	fmt.Println("Comparing files…")

	// Using goroutine
	logs := make(chan string)
	var wg sync.WaitGroup
	wg.Add(2)
	compareFiles(logs, &wg, secondDir, firstDir)
	findRemovedFiles(logs, &wg, secondDir, firstDir)
	wg.Wait()

	copyToClipboard(logs)
}

func compareFiles(channel chan<- string, wg *sync.WaitGroup, newApkDir string, oldApkDir string) {
	files, err := ioutil.ReadDir(newApkDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, f := range files {
		var secondDirFileName = filepath.Join(oldApkDir, f.Name())
		var secondSize = getSize(secondDirFileName)

		var name = f.Name()
		var firstSize = getSize(filepath.Join(newApkDir, name))

		channel <- fmt.Sprintf("%s, %d,, %s, %d, %d\n", name, firstSize, name, secondSize, firstSize-secondSize)

		if secondSize == 0 {
			fmt.Printf(NewFile, name)
			continue
		}

		if firstSize > secondSize {
			fmt.Printf(Increase, name, firstSize, secondSize)
		} else if firstSize < secondSize {
			fmt.Printf(Decrease, name, firstSize, secondSize)
		} else {
			fmt.Printf(Same, name, firstSize, secondSize)
		}
	}
}

func findRemovedFiles(channel chan<- string, wg *sync.WaitGroup, newApkDir string, oldApkDir string) {
	files, err := ioutil.ReadDir(oldApkDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, f := range files {
		name := f.Name()
		fileInNewApk := filepath.Join(newApkDir, name)
		isExist := isExists(fileInNewApk)

		if !isExist {
			fmt.Printf(RemovedFile, name)
			channel <- fmt.Sprintf("%s, %d,, %s, %d, %d\n", name, 0, name, f.Size(), f.Size())
		}
	}

	wg.Done()
}

func isExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getSize(fileName string) int64 {
	var fileInfo, err = os.Stat(fileName)
	var size = int64(0)
	if err == nil {
		if fileInfo.IsDir() {
			size, err := getDirSize(fileName)
			if err == nil {
				return size
			}
		} else {
			size = fileInfo.Size()
		}
	}
	return size
}

func copyToClipboard(allData <-chan string) {
	var buffer bytes.Buffer

	for data := range allData {
		buffer.WriteString(data)
	}

	clipboard.WriteAll(buffer.String())

	fmt.Println("\n\nAll data has been copied to clipboard!")
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

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
