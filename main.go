package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

type SiteInfo struct {
	UID  string `json:"uid"`
	Host string `json:"host"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s URL FOLDER\n", os.Args[0])
		os.Exit(1)
	}

	url := os.Args[1]
	folder := os.Args[2]

	mainPage, err := fetchPage(url)
	if err != nil {
		fmt.Printf("Error fetching main page: %v\n", err)
		os.Exit(1)
	}

	siteInfo, err := extractSiteInfo(mainPage)
	if err != nil {
		fmt.Printf("Unable to extract siteInfo: %v\n", err)
		os.Exit(1)
	}

	cacheData, err := fetchCache(siteInfo.Host, siteInfo.UID)
	if err != nil {
		fmt.Printf("Error fetching cache: %v\n", err)
		os.Exit(1)
	}

	downloadFiles(siteInfo.Host, siteInfo.UID, folder, cacheData)
}

func fetchPage(url string) (string, error) {
	// #nosec G107 -- URL is validated by user input
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func extractSiteInfo(mainPage string) (*SiteInfo, error) {
	re := regexp.MustCompile(`window\.siteInfo\s*=\s*({[^}]+})`)
	matches := re.FindStringSubmatch(mainPage)

	if len(matches) == 0 {
		return nil, fmt.Errorf("siteInfo not found")
	}

	var siteInfo SiteInfo
	err := json.Unmarshal([]byte(matches[1]), &siteInfo)
	if err != nil {
		return nil, err
	}

	return &siteInfo, nil
}

func fetchCache(host, uid string) (map[string]interface{}, error) {
	cacheURL := fmt.Sprintf("https://%s/cache/%s", host, uid)
	// #nosec G107 -- URL is constructed from validated site info
	resp, err := http.Get(cacheURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cacheData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&cacheData)
	if err != nil {
		return nil, err
	}

	return cacheData, nil
}

func downloadFiles(host, uid, folder string, cacheData map[string]interface{}) {
	total := len(cacheData)
	current := 0

	for key := range cacheData {
		current++
		fmt.Printf("Downloading %d/%d: %s\n", current, total, key)

		accessURL := fmt.Sprintf("https://%s/access/%s/%s", host, uid, key)
		// #nosec G107 -- URL is constructed from validated site info
		resp, err := http.Get(accessURL)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", key, err)
			continue
		}

		path := filepath.Join(folder, key)
		parentFolder := filepath.Dir(path)

		// #nosec G301 -- Directory permissions are appropriate for user downloads
		err = os.MkdirAll(parentFolder, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", parentFolder, err)
			if closeErr := resp.Body.Close(); closeErr != nil {
				fmt.Printf("Error closing response body: %v\n", closeErr)
			}
			continue
		}

		// #nosec G304 -- Path is constructed from validated key and user-provided folder
		file, err := os.Create(path)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", path, err)
			if closeErr := resp.Body.Close(); closeErr != nil {
				fmt.Printf("Error closing response body: %v\n", closeErr)
			}
			continue
		}

		_, err = io.Copy(file, resp.Body)
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Error closing file: %v\n", closeErr)
		}

		if err != nil {
			fmt.Printf("Error writing file %s: %v\n", path, err)
			if removeErr := os.Remove(path); removeErr != nil {
				fmt.Printf("Error removing file %s: %v\n", path, removeErr)
			}
		}
	}

	fmt.Printf("Downloaded %d files to %s\n", total, folder)
} 