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

	OutputsDirName   = "outputs"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: appdiff <new_app> <old_app> <level_folder>")
		return
	}

	newFile := os.Args[1]
	oldFile := os.Args[2]

	levelFolder := int(0)
	if len(os.Args) > 3 {
		level := os.Args[3]
		i, _ := strconv.Atoi(level)
		levelFolder = i
	}

	isIpa := isIpaPackage(newFile)

	// Create temp directory when unzip
	newDir, _ := ioutil.TempDir("", "appdiff")
	oldDir, _ := ioutil.TempDir("", "appdiff")

	unzip(newFile, newDir)
	unzip(oldFile, oldDir)

	// Using goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	var logdiff []string

	go func() {
		defer wg.Done()
		logdiff = diffFilesToRecords(newDir, oldDir, "", levelFolder, isIpa)
	}()
	wg.Wait()

	// Remove temp directory when unzip
	os.RemoveAll(newDir)
	os.RemoveAll(oldDir)

	appdiffDir := getwd()

	// Copy result to clipboard
	replacer := strings.NewReplacer(".ipa", "", ".apk", "")
	newFileName := replacer.Replace(newFile)
	oldFileName := replacer.Replace(oldFile)

	copyToClipboard(newFileName, oldFileName, append(logdiff))
	createFile(appdiffDir, newFileName, oldFileName, append(logdiff))
}

func getLastSlice(value string, separator string) string {
	s := strings.Split(value, separator)
	return s[len(s)-1]
}

func diffFilesToRecords(newDir string, oldDir string, folderName string, levelFolder int, isIpa bool) []string {

	newFiles := readDir(newDir)
	oldFiles := readDir(oldDir)

	records := make([]string, len(oldFiles))

	var merged []os.FileInfo
	merged = oldFiles
	merged = merging(merged, newFiles)

	for _, f := range merged {

		name := f.Name()
		filename := name
		if len(folderName) > 0 {
			filename = folderName + "/" + filename
		}

		logFilename := trimLog(name, folderName)

		var oldDirFileName = filepath.Join(oldDir, name)
		var oldSize = getSize(oldDirFileName)
		var newSize = getSize(filepath.Join(newDir, name))

		records = append(records, fmt.Sprintf("%s, %d, , %s, %d, %d\n", logFilename, newSize, logFilename, oldSize, newSize-oldSize))

		if shouldPrintLog(name, isIpa) {
			if newSize > oldSize {
				fmt.Printf(Increase, logFilename, newSize, oldSize)
			} else if newSize < oldSize {
				fmt.Printf(Decrease, logFilename, newSize, oldSize)
			} else {
				fmt.Printf(Same, logFilename, newSize, oldSize)
			}
		}

		// Check level folder
		if levelFolder > 0 {
			splits := strings.Split(filename, "/")
			level := levelFolder
			if isIpa {
				level = level + 2
			}

			if len(splits) >= level {
				continue
			}
		}

		// Handling nested folder
		if f.IsDir() {
			// Remove folder size
			records = records[:len(records)-1]

			subPath := filepath.Join(newDir, name)
			subSecondPath := filepath.Join(oldDir, name)
			subRecords := diffFilesToRecords(subPath, subSecondPath, filename, levelFolder, isIpa)

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

	// Remove path 'Payload' (iOS)
	if strings.Contains(filename, "Payload/") {
		filename = strings.Trim(filename, "Payload/")
	}

	// Remove path 'x.app' (iOS)
	if strings.Contains(filename, ".app/") {
		path := ""
		splits := strings.Split(filename, "/")
		for _, f := range splits {
			if !strings.HasSuffix(f, ".app") {
				if len(splits) == 2 {
					path = path + f
				} else {
					path = path + f + "/"
				}
			}
		}

		path = strings.TrimSuffix(path, "/")

		if len(path) > 0 {
			filename = path
		}
	}

	return filename
}

func shouldPrintLog(name string, isIpa bool) bool {
	shouldPrint := true
	if isIpa {
		if strings.Contains(name, "Payload") || filepath.Ext(name) == ".app" {
			shouldPrint = false
		}
	}
	return shouldPrint
}
func isIpaPackage(filename string) bool {
	return strings.Contains(filename, ".ipa")
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

func copyToClipboard(newFileName string, oldFileName string, allData []string) {
	// Header columns
	allData = append([]string{newFileName, ", Size, , ", oldFileName, ", Size, Diff\n"}, allData...)

	var buffer bytes.Buffer

	for _, data := range allData {
		buffer.WriteString(data)
	}

	clipboard.WriteAll(buffer.String())

	fmt.Println("\n\nAll data has been copied to clipboard!")
}

func createFile(dir string, newFileName string, oldFileName string, allData []string) {
	fileName := fmt.Sprintf("%s_%s_%s", getLastSlice(dir, "/"), newFileName, oldFileName)

	chdir(dir)

	os.RemoveAll(OutputsDirName)

	errDir := os.Mkdir(OutputsDirName, 0755)
	checkIfError(errDir)

	chdir(OutputsDirName)

	f, err := os.Create(fmt.Sprintf("%s.txt", fileName))
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	allData = append([]string{newFileName, ", Size, , ", oldFileName, ", Size, Diff\n"}, allData...)

	for _, v := range allData {
		fmt.Fprint(f, v)
	}
	err = f.Close()
	checkIfError(err)

	path := getwd()

	fmt.Printf("File %s.txt has been created at %s/%s.txt\n", fileName, path, fileName)
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

func chdir(dir string) {
	err := os.Chdir(dir)
	checkIfError(err)
}

func getwd() string {
	path, err := os.Getwd()
	checkIfError(err)
	return path
}

func checkIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(-1)
}