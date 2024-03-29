package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
)

// File struct from the response, we will only use URL.
type File struct {
	Name string `json:"name"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

// Response struct of the request
type Response struct {
	Success bool   `json:"success"`
	Files   []File `json:"files"`
}

func main() {
	var wg sync.WaitGroup

	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("Please provide arguments. Usage: gong [files]")
	}

	// Upload each files, one per one.
	for _, p := range flag.Args() {
		wg.Add(1)
		go func(file string) {
			if exists(file) {
				// Upload the file
				link, err := upload(file)
				if err != nil {
					fmt.Println(file+":", "upload failed:", err)
				}
				fmt.Println(file+":", link)
			} else {
				fmt.Println("file doesn't exist:", file)
			}
			wg.Done()
		}(p)
	}

	wg.Wait()
}

// Actually upload the file
func upload(filepath string) (string, error) {
	// Open the file, reading only.
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Get file info
	fi, err := os.Stat(filepath)
	if err != nil {
		return "", err
	}

	// Create form
	form, err := w.CreateFormFile("files[]", fi.Name())
	if err != nil {
		return "", err
	}

	// Write to form
	if _, err := io.Copy(form, file); err != nil {
		return "", err
	}
	// Close file and writer.
	file.Close()
	w.Close()

	req, err := http.NewRequest("POST", "https://gang.moe/api/upload", &b)
	if err != nil {
		return "", err
	}

	// Set the correct Content-Type
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Prepare client and call API
	client := &http.Client{}
	// Unmarshal and return response.
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)

	var r Response
	json.Unmarshal(body, &r)

	// Close the body
	resp.Body.Close()

	if r.Success {
		return r.Files[0].URL, nil
	}

	// Unknown error. Check your network.
	return "Upload failed. Check your network.", nil

}

// Check if the file exists
func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
