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

	src, dst, pathSeperator, name, chunkSize, keepUnknownFileTypes, discordMode := readArgs()

	if len(src) < 1 || len(dst) < 1 {
		printHelp()
		fmt.Printf("src: %v", src)
		fmt.Printf("dst: %v", dst)
		log.Fatal("Invalid paths.")
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(src)

	if err != nil {
		fmt.Println("There are no files in this directory.")
		printHelp()
		fmt.Println("PROGRAM STOPPED")
		os.Exit(2)
	}

	sourceBasePath := src + pathSeperator

	chunkList := []os.FileInfo{}
	var calculatedFiles int = 0

	fmt.Println("Copying files.")

	for num, f := range files {

		if discordMode {
			if !strings.Contains(f.Name(), "data") && !strings.Contains(f.Name(), "index") && !strings.Contains(f.Name(), ".") {
				chunkList = append(chunkList, f)
			}
		} else {
			if !strings.Contains(f.Name(), ".") {
				chunkList = append(chunkList, f)
			}
		}

		if len(chunkList)%chunkSize == 0 && num != 0 {
			wg.Add(1)
			go fileArrayCopy(chunkList, dst, sourceBasePath, name, calculatedFiles, keepUnknownFileTypes)
			chunkList = []os.FileInfo{}
			calculatedFiles += chunkSize
		}
	}

	if len(chunkList) > 0 {
		wg.Add(1)
		go fileArrayCopy(chunkList, dst, sourceBasePath, name, calculatedFiles, keepUnknownFileTypes)
	}

	wg.Wait()

}

func readArgs() (src string, dst string, pathSeperator string, name string, chunkSize int, keepUnknownFileTypes bool, discordMode bool) {

	if len(os.Args[1:]) < 1 {
		printHelp()
		log.Fatal("Missing arguments!")
	}

	argsWithoutProg := os.Args[1:] // Argument Input

	// STANDARD PARAMETER
	dst, _ = os.Getwd()
	chunkSize = 10
	keepUnknownFileTypes = false
	pathSeperator = "\\"
	name = "Picture_"
	discordMode = true

	if runtime.GOOS != "windows" {
		pathSeperator = "/"
	}

	if len(argsWithoutProg) == 1 { // Accept execution with one parameter without defining it with -src
		src = argsWithoutProg[0]

	} else {
		for i := 0; i < len(argsWithoutProg); i++ {

			if argsWithoutProg[i] == "-src" {
				if pathSeperator == "\\" {
					src = strings.Replace(argsWithoutProg[i+1], "\\", "\\\\", -1)
				} else {
					src = argsWithoutProg[i+1]
				}
				continue
			}

			if argsWithoutProg[i] == "-dst" {
				if pathSeperator == "\\" {
					dst = strings.Replace(argsWithoutProg[i+1], "\\", "\\\\", -1)
				} else {
					dst = argsWithoutProg[i+1]
				}
				continue
			}

			if argsWithoutProg[i] == "-cs" {
				tmp, err := strconv.Atoi(argsWithoutProg[i+1])
				if err != nil || tmp < 1 {
					chunkSize = 5
					log.Printf("Error: %v \n Set chunkSize to %v.", err, chunkSize)
				} else {
					chunkSize = tmp
				}
				continue
			}

			if argsWithoutProg[i] == "-tc" {
				threadCount, err := strconv.Atoi(argsWithoutProg[i+1])

				if err != nil || threadCount < 1 {
					log.Printf("Error: %v \n threadCount was not set.", err)
				} else {
					runtime.GOMAXPROCS(threadCount)
				}
				continue
			}

			if argsWithoutProg[i] == "-s" {
				pathSeperator = argsWithoutProg[i+1]
			}

			if argsWithoutProg[i] == "-k" {
				keepUnknownFileTypes = true
				continue
			}

			if argsWithoutProg[i] == "-n" {
				name = argsWithoutProg[i+1]
			}

			if argsWithoutProg[i] == "-!d" {
				discordMode = false
			}
		}
	}

	if src == "" || dst == "" { // Error detection
		log.Fatal("Necessary parameters are missing.")
	}
	if string(dst[len(dst)-1:]) != pathSeperator { // Make sure the destination is a Folder.
		dst += pathSeperator
	}

	return
}

func printHelp() {
	fmt.Println("HELP PAGE")
	fmt.Println("-src [Path] - enter Discord path *necessary")
	fmt.Println("-dst [Path] - enter path to save")
	fmt.Println("-cs [Num] - enter how big the chunk for each thread should be")
	fmt.Println("-tc [Num] - How many threads should run at once")
	fmt.Println("-k - keep files with unknown filetype")
	fmt.Println("-s [/ or \\\\] - seperator used by your file system")
}

func fileArrayCopy(files []os.FileInfo, dst string, orig string, name string, startName int, keepUnknownFileTypes bool) {

	defer wg.Done()

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
					err = copy(orig+f.Name(), dst+name+strconv.Itoa(startName)+"."+fileType)

					if err != nil {
						log.Printf("Unable to copy: %v \n Error: %v", f, err)
					} else {
						startName++
					}
				}
			}
		}
	}
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
