package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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
const extraPathForIpa = "/Payload/bl_ios.app"

var whitelistFolder = []string{"Payload", "bl_ios.app", "Frameworks", "PlugIns"}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: appdiff <new_app> <old_app>")
		return
	}

	newFile := os.Args[1]
	oldFile := os.Args[2]

	newDir, _ := ioutil.TempDir("", "app")
	oldDir, _ := ioutil.TempDir("", "app")

	unzip(newFile, newDir)
	unzip(oldFile, oldDir)

	var isIpa = isIpaPackage(newDir)
	if isIpa {
		newDir = newDir + extraPathForIpa
		oldDir = oldDir + extraPathForIpa
	}

	// Using goroutine
	var wg sync.WaitGroup
	wg.Add(2)

	var firstLog []string
	var secondLog []string

	go func() {
		defer wg.Done()
		firstLog = diffFilesToRecords(newDir, oldDir)
	}()
	go func() {
		defer wg.Done()
		secondLog = findRemovedFiles(newDir, oldDir)
	}()
	wg.Wait()

	// Copy result to clipboard
	copyToClipboard(append(firstLog, secondLog...))
}

func diffFilesToRecords(newDir string, oldDir string) []string {

	files := readDir(newDir)
	records := make([]string, len(files))

	for _, f := range files {

		var name = f.Name()
		var oldDirFileName = filepath.Join(oldDir, name)
		var oldSize = getSize(oldDirFileName)
		var newSize = getSize(filepath.Join(newDir, name))

		records = append(records, fmt.Sprintf("%s, %d, , %s, %d, %d\n", name, newSize, name, oldSize, newSize-oldSize))

		if oldSize == 0 {
			fmt.Printf(NewFile, name)
			continue
		}

		if newSize > oldSize {
			fmt.Printf(Increase, name, oldSize, newSize)
		} else if newSize < oldSize {
			fmt.Printf(Decrease, name, oldSize, newSize)
		} else {
			fmt.Printf(Same, name, newSize, oldSize)
		}

		// Handling IPA files
		if f.IsDir() && contains(whitelistFolder, name) {
			subPath := filepath.Join(newDir, name)
			subSecondPath := filepath.Join(oldDir, name)
			subRecords := diffFilesToRecords(subPath, subSecondPath)
			if len(subRecords) > 0 {
				records = append(records, subRecords...)
			}
		}
	}
	return records
}

func findRemovedFiles(newDir string, oldDir string) []string {
	files, err := ioutil.ReadDir(oldDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	removedFiles := make([]string, 0)

	for _, f := range files {
		size := f.Size()
		name := f.Name()
		fileInNewApk := filepath.Join(newDir, name)
		isExist := isExists(fileInNewApk)

		if !isExist {
			fmt.Printf(RemovedFile, name)
			removedFiles = append(removedFiles, fmt.Sprintf("%s, %d,, %s, %d, %d\n", name, 0, name, size, size))
		}
	}

	return removedFiles
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

func copyToClipboard(allData []string) {
	// header columns
	allData = append([]string{"Right version, Size, , Left version, , Size, Diff\n"}, allData...)

	var buffer bytes.Buffer

	for _, data := range allData {
		buffer.WriteString(data)
	}

	clipboard.WriteAll(buffer.String())

	fmt.Println("\n\nAll data has been copied to clipboard!")
}

func unzip(path string, destDir string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println(err)
		os.Exit(0)
	}

	_, err = zipper.Unzip(absPath, destDir)

	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println(err)
		os.Exit(0)
	}
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

func isIpaPackage(filename string) bool {
	return strings.Contains(filename, ".ipa")
}

func readDir(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println(err)
		os.Exit(0)
	}
	return files
}

func dirToFileInfo(dir string) os.FileInfo {
	fileinfo, err := os.Stat(dir)
	if err != nil {
		log.Println(err)
		return nil
	}
	return fileinfo
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.Contains(a, e) {
			return true
		}
	}
	return false
}
