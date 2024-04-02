package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	baseURL := "https://access.redhat.com/security/data/csaf/beta/vex/2022/"

	// Make GET request to the base URL
	resp, err := http.Get(baseURL)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	// Create a directory to store the downloaded files
	err = os.MkdirAll("", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	// Decode the HTML response to find the links
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading response:", err)
			return
		}
		if n == 0 {
			break
		}

		// Find links ending with ".json"
		links := findJSONLinks(buf[:n])
		for _, link := range links {
			filename := filepath.Join("downloaded_files", filepath.Base(link))
			err := downloadFile(baseURL+link, filename)
			if err != nil {
				fmt.Println("Error downloading file:", err)
			} else {
				fmt.Println("File downloaded:", filename)
			}
		}
	}

	// Define the path to the "2022" folder
	folder := "2022"

	// Create the index.txt file
	err = createIndexFile(folder)
	if err != nil {
		fmt.Println("Error creating index.txt file:", err)
	} else {
		fmt.Println("index.txt file created successfully")
	}

	// Create the changes.csv file
	err = createChangesCSV(folder)
	if err != nil {
		fmt.Println("Error creating changes.csv file:", err)
	} else {
		fmt.Println("changes.csv file created successfully")
	}
}

// Function to find links ending with ".json"
func findJSONLinks(data []byte) []string {
	var links []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, ".json") {
			parts := strings.Split(line, `"`)
			for _, part := range parts {
				if strings.HasSuffix(part, ".json") {
					links = append(links, part)
				}
			}
		}
	}
	return links
}

// Function to download a file from a URL and save it to disk
func downloadFile(url, filename string) error {
	// Make GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// Function to create index.txt file and write rows for each file in the folder
func createIndexFile(folder string) error {
	// Open the 2022 folder
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return err
	}

	// Create index.txt file
	indexFile, err := os.Create("index.txt")
	if err != nil {
		return err
	}
	defer indexFile.Close()

	// Write rows for each file in the folder to index.txt
	for _, file := range files {
		if !file.IsDir() {
			// Write the folder name and file name to index.txt
			_, err := indexFile.WriteString(fmt.Sprintf("%s/%s\n", folder, file.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Function to create changes.csv file and write rows for each file in the folder
func createChangesCSV(folder string) error {
	// Open the 2022 folder
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return err
	}

	// Create changes.csv file
	changesFile, err := os.Create("changes.csv")
	if err != nil {
		return err
	}
	defer changesFile.Close()

	// Write rows for each file in the folder to changes.csv
	for _, file := range files {
		if !file.IsDir() {
			// Get the last modified time of the file
			modTime := file.ModTime().Format(time.RFC3339)

			// Write the folder name, file name, and last modified time to changes.csv
			_, err := changesFile.WriteString(fmt.Sprintf("\"%s/%s\",\"%s\"\n", folder, file.Name(), modTime))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
