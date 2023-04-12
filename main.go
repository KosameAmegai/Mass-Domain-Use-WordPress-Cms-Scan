package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("script by kosameamegai / for usage: go run main.go <list.txt>")
		os.Exit(1)
	}

	filename := os.Args[1]
	targets, err := readLines(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	sem := make(chan bool, 200)

	for _, target := range targets {
		sem <- true
		wg.Add(1)
		go func(site string) {
			defer func() {
				<-sem
				wg.Done()
			}()
			checkWordPress(site)
		}(target)
	}

	wg.Wait()
}

func readLines(filename string) ([]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	return lines, nil
}

func checkWordPress(site string) {
	url := normalizeURL(site)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf(" --| %s --> [Not WordPress]\n", site)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf(" --| %s --> [Not WordPress]\n", site)
		return
	}

	src := string(body)
	re := regexp.MustCompile(`<meta name="generator" content="(.*)" />`)
	found := re.FindStringSubmatch(src)

	if found != nil && strings.Contains(found[1], "WordPress") {
		fmt.Printf(" --| %s --> [WordPress]\n", site)
		appendToFile("wp.txt", site+"/\n")
	} else if strings.Contains(src, "wp-content/themes") {
		fmt.Printf(" --| %s --> [WordPress]\n", site)
		appendToFile("wp.txt", site+"/\n")
	} else {
		fmt.Printf(" --| %s --> [Not WordPress]\n", site)
	}
}

func normalizeURL(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	return url
}

func appendToFile(filename, text string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(text)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}
