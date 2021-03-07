package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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