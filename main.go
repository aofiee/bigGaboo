package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/labstack/gommon/log"
)

var argDownloaderURL string
var argSavePath string

//ImagesStruct structure
type ImagesStruct struct {
	imageURL  string
	imageName string
}

func main() {
	args := os.Args
	argumentReciever(args)
}

func argumentReciever(args []string) {
	argDownloaderURL = ""
	argSavePath = ""
	help := `
	Images Downloder from URL
	Usage:
		go run main.go [arguments]
		or go build 
		./main [arguments]

	-u	URL of content
	-f	path to save
	`
	fmt.Println(help)
	for index, arg := range args {
		if arg == "-u" {
			argDownloaderURL = args[index+1]
			if !isValidURL(argDownloaderURL) {
				log.Error("URL is invalid")
				os.Exit(1)
			}
		}
		if arg == "-f" {
			argSavePath = args[index+1]
		}
	}
	if argDownloaderURL != "" && argSavePath != "" {
		if _, err := os.Stat(argSavePath); err != nil {
			extractImagesFromURL(argDownloaderURL, argSavePath)
		} else {
			log.Error("Found folder")
		}
	}
}
func extractImagesFromURL(URL string, dirName string) {
	os.Mkdir(dirName, 0777)
	result := getURLContent(URL)
	imgTags := getImagesURL(result)
	for _, item := range imgTags {
		dirSavePath := dirName + "/" + item.imageName
		goDownload(dirSavePath, item.imageURL)
	}
}
func goDownload(filename string, URL string) {
	err := DownloadFile(filename, URL)
	if err != nil {
		log.Error(err)
	}
}
func getURLContent(URL string) (convertToString string) {
	res, err := http.Get(URL)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	contents, err := ioutil.ReadAll(res.Body)
	convertToString = string(contents)
	return
}

func getImagesURL(html string) []ImagesStruct {
	var imgRex = regexp.MustCompile(`<img[^>]+\bsrc=["']([^"']+)["']`)
	imgTags := imgRex.FindAllStringSubmatch(html, -1)
	output := make([]ImagesStruct, len(imgTags))
	for index := range output {
		output[index].imageURL = imgTags[index][1]
		output[index].imageName = path.Base(output[index].imageURL)
	}
	return output
}

//WriteCounter struct
type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

//PrintProgress func
func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

//DownloadFile func
func DownloadFile(filepath string, url string) error {

	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}
	fmt.Print("\n")
	err = os.Rename(filepath+".tmp", filepath)
	if err != nil {
		return err
	}
	return nil
}

//isValidURL check
func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}
	return true
}
