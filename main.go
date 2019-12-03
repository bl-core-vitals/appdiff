package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

	Increase = "\033[1;33m[>] %s : %d => %d\033[0m\n"
	Decrease = "\033[1;34m[<] %s : %d => %d\033[0m\n"
	Same     = "\033[1;36m[=] %s : %d => %d\033[0m\n"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: appdiff <new_app> <old_app> <level_folder>")
		return
	}

	newFile := os.Args[1]
	oldFile := os.Args[2]

	var level = int(0)
	if len(os.Args) > 3 {
		rootLevel := os.Args[3]
		i, _ := strconv.Atoi(rootLevel)
		level = i
	}

	newDir, _ := ioutil.TempDir("", "app")
	oldDir, _ := ioutil.TempDir("", "app")

	unzip(newFile, newDir)
	unzip(oldFile, oldDir)

	// Using goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	var logdiff []string

	go func() {
		defer wg.Done()
		logdiff = diffFilesToRecords(newDir, oldDir, "", level)
	}()
	wg.Wait()

	// Copy result to clipboard
	copyToClipboard(append(logdiff))
}

func diffFilesToRecords(newDir string, oldDir string, folderName string, level int) []string {

	newFiles := readDir(newDir)
	oldFiles := readDir(oldDir)

	records := make([]string, len(oldFiles))

	var merged []os.FileInfo
	merged = oldFiles
	merged = merging(merged, newFiles)

	for _, f := range merged {

		name := f.Name()
		filename := trimLog(name, folderName)

		var oldDirFileName = filepath.Join(oldDir, name)
		var oldSize = getSize(oldDirFileName)
		var newSize = getSize(filepath.Join(newDir, name))

		records = append(records, fmt.Sprintf("%s, %d, , %s, %d, %d\n", filename, newSize, filename, oldSize, newSize-oldSize))

		if newSize > oldSize {
			fmt.Printf(Increase, filename, newSize, oldSize)
		} else if newSize < oldSize {
			fmt.Printf(Decrease, filename, newSize, oldSize)
		} else {
			fmt.Printf(Same, filename, newSize, oldSize)
		}

		// check level folder
		if level > 0 {
			splits := strings.Split(filename, "/")
			if len(splits) >= level {
				continue
			}
		}

		// is app folder (ios)
		var isAppFolder = filepath.Ext(name) == ".app"

		// Handling package folder
		if f.IsDir() || isAppFolder {
			// remove folder size
			records = records[:len(records)-1]

			subPath := filepath.Join(newDir, name)
			subSecondPath := filepath.Join(oldDir, name)

			subRecords := diffFilesToRecords(subPath, subSecondPath, filename, level)

			if len(subRecords) > 0 {
				records = append(records, subRecords...)
			}
		}
	}
	return records
}

func trimLog(name string, folderName string) string {
	filename := name
	if len(folderName) > 0 {
		filename = folderName + "/" + filename
	}

	// remove path 'Payload' (iOS)
	if strings.Contains(filename, "Payload/") {
		filename = strings.Trim(filename, "Payload/")
	}

	// remove path 'x.app' (iOS)
	if strings.Contains(filename, ".app/") {
		path := ""
		splits := strings.Split(filename, "/")
		for _, f := range splits {
			if !strings.Contains(f, ".app") {
				if len(splits) == 2 {
					path = path + f
				} else {
					path = path + f + "/"
				}
			}
		}
		if len(path) > 0 {
			filename = path
		}
	}

	return filename
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
	allData = append([]string{"Right version, Size, , Left version, Size, Diff\n"}, allData...)

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

func readDir(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		var empty []os.FileInfo
		return empty
	}
	return files
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.Contains(a, e) {
			return true
		}
	}
	return false
}

func merging(oldest []os.FileInfo, newest []os.FileInfo) []os.FileInfo {
	oldFilePaths := stringFileInfos(oldest)
	var mergedFileInfos []os.FileInfo = oldest

	for _, fileInfo := range newest {
		namePath := fileInfo.Name()
		if !contains(oldFilePaths, namePath) {
			mergedFileInfos = append(mergedFileInfos, fileInfo)
		}
	}
	return mergedFileInfos
}

func stringFileInfos(paths []os.FileInfo) []string {
	filePaths := make([]string, 0)
	for _, filepath := range paths {
		namePath := filepath.Name()
		filePaths = append(filePaths, namePath)
	}
	return filePaths
}
