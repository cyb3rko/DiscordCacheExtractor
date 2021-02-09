package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var wg = sync.WaitGroup{}

func main() {

	src, dst, chunkSize, keepUnknownFileTypes := readArgs()

	if len(src) < 1 || len(dst) < 1 {
		printHelp()
		fmt.Printf("src: %v", src)
		fmt.Printf("dst: %v", dst)
		log.Fatal("Invalid paths.")
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(src)
	sourceBasePath := src + "\\"

	if err != nil {
		fmt.Println("There are no files in this directory.")
		printHelp()
		fmt.Println("PROGRAM STOPPED")
		os.Exit(2)
	}

	chunkList := []os.FileInfo{}
	var calculatedFiles int = 0

	fmt.Println("Copying files.")

	for num, f := range files {

		if !strings.Contains(f.Name(), "data") && !strings.Contains(f.Name(), "index") {
			chunkList = append(chunkList, f)
		}

		if len(chunkList)%chunkSize == 0 && num != 0 {
			wg.Add(1)
			go fileArrayCopy(chunkList, dst, sourceBasePath, calculatedFiles, keepUnknownFileTypes)
			chunkList = []os.FileInfo{}
			calculatedFiles += chunkSize
		}
	}

	if len(chunkList) > 0 {
		wg.Add(1)
		go fileArrayCopy(chunkList, dst, sourceBasePath, calculatedFiles, keepUnknownFileTypes)
	}

	wg.Wait()

}

func readArgs() (string, string, int, bool) {

	argsWithoutProg := os.Args[1:] // Argument Input

	// STANDARD PARAMETER
	dst := ""
	src := ""
	chunkSize := 10
	keepUnknownFileTypes := false

	for i := 0; i < len(argsWithoutProg); i++ {

		if argsWithoutProg[i] == "-src" {
			src = strings.Replace(argsWithoutProg[i+1], "\\", "\\\\", -1)
			continue
		}

		if argsWithoutProg[i] == "-dst" {
			dst = strings.Replace(argsWithoutProg[i+1], "\\", "\\\\", -1)
			continue
		}

		if argsWithoutProg[i] == "-cs" {
			tmp, err := strconv.Atoi(argsWithoutProg[i+1])
			if err != nil || chunkSize < 1 {
				chunkSize = 5
				log.Printf("Error: %v \n Set chunkSize to %v.", err, chunkSize)
			} else {
				chunkSize = tmp
			}
			continue
		}

		if argsWithoutProg[i] == "-t" {
			threadCount, err := strconv.Atoi(argsWithoutProg[i+1])

			if err != nil || threadCount < 1 {
				log.Printf("Error: %v \n threadCount was not set.", err)
			} else {
				runtime.GOMAXPROCS(threadCount)
			}
			continue
		}

		if argsWithoutProg[i] == "-k" {
			keepUnknownFileTypes = true
		}
	}

	return src, dst, chunkSize, keepUnknownFileTypes
}

func printHelp() {
	fmt.Println("HELP PAGE") // todo: Design Help Page
	fmt.Println("-src [Path] - enter Discord path")
	fmt.Println("-dst [Path] - enter path to save")
	fmt.Println("-cs [Num] - enter how big the chunk for each thread should be")
	fmt.Println("-t [Num] - How many threads should run at once")
	fmt.Println("-k - keep files with unknown filetype")
}

func fileArrayCopy(files []os.FileInfo, dst string, orig string, startName int, keepUnknownFileTypes bool) {

	for _, f := range files {
		file, err := os.Open(orig + f.Name())

		if err != nil {
			log.Printf("Can not open file: %v \n Error: %v", f, err)

		} else {
			defer file.Close()

			fileTypeRaw, err := getFileContentType(file)
			if err != nil {
				log.Printf("Unable to get filetype of: %v \n Error: %v", f, err)
			} else {

				fileTypeRaw = strings.ToLower(fileTypeRaw) // image/png
				fileType := strings.Split(fileTypeRaw, "/")[1]

				if fileType != "octet-stream" || keepUnknownFileTypes { // Skip Files with unknown type
					err = copy(orig+f.Name(), dst+f.Name()+"."+fileType)

					if err != nil {
						log.Printf("Unable to copy: %v \n Error: %v", f, err)
					}
				}
			}
		}
	}
	wg.Done()
}

func copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func getFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}