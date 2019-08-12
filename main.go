package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/esafirm/appdiff/zipper"
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
const extraPathForIpa = "/Payload/bl_ios.app"

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: apkdiff <new_apk> <old_apk>")
		return
	}

	firstApk := os.Args[1]
	secondApk := os.Args[2]

	firstDir, _ := ioutil.TempDir("", "apk")
	secondDir, _ := ioutil.TempDir("", "apk")

	unzip(firstApk, firstDir)
	unzip(secondApk, secondDir)

	files := readDir(firstDir)

	var isIpa = isIpaPackage(firstApk)
	if isIpa {
		tmpFirstApp := filepath.Join(firstDir, extraPathForIpa)
		files = readDir(tmpFirstApp)

		secondDir = secondDir + extraPathForIpa
	}

	allData := make([]string, len(files))

	fmt.Println("Comparing files…")

	for index, f := range files {
		var secondDirFileName = filepath.Join(secondDir, f.Name())
		var secondSize = getSize(secondDirFileName)
		var secondFileInfo = dirToFileInfo(secondDirFileName)

		secondSize := int64(0)
		if secondFileInfo == nil {
			// file is new, leave it appear on record.
		} else {
			secondSize = secondFileInfo.Size()
		}

		var name = f.Name()
		var firstSize = getSize(filepath.Join(firstDir, name))

		allData[index] = fmt.Sprintf("%s, %d,, %s, %d, %d\n", name, firstSize, name, secondSize, firstSize-secondSize)

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

	copyToClipboard(allData)
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
