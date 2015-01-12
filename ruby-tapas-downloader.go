package main

import (
	"encoding/xml"
	"flag"
	"github.com/chrisdambrosio/sanitize"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Feed struct {
	XMLName  xml.Name  `xml:"rss"`
	Title    string    `xml:"channel>title"`
	Episodes []Episode `xml:"channel>item"`
}

type Episode struct {
	Title       string      `xml:"title"`
	EpisodeFile EpisodeFile `xml:"enclosure"`
}

type EpisodeFile struct {
	Url string `xml:"url,attr"`
}

type Client struct {
	username string
	password string
	feedUrl  string
}

func (c Client) Get(url string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(c.username, c.password)

	resp, err := client.Do(req)

	return resp, err
}

func (client Client) fetchFeed() []byte {
	resp, err := client.Get(client.feedUrl)
	defer resp.Body.Close()

	if err != nil {
		log.Fatal("Error: failed to fetch feed - ", err)
	}

	rss, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Error: while reading feed - ", err)
	}

	return rss
}

func (client Client) downloadFile(url, target string) {
	tmpFile := target + ".part"
	out, err := os.Create(tmpFile)
	defer out.Close()

	if err != nil {
		log.Println("Error: error copying file", target, "-", err)
	}

	resp, err := client.Get(url)
	defer resp.Body.Close()

	if err != nil {
		log.Println("Error: failed to fetch file", url, "-", err)
	}

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		log.Fatal("Error:", err)
	}

	os.Rename(tmpFile, target)

	log.Println("Downloaded:", url)
}

func main() {
	var username = flag.String("u", "", "login username")
	var password = flag.String("p", "", "login password")
	flag.Parse()

	url := "https://rubytapas.dpdcart.com/feed"

	client := Client{username: *username, password: *password, feedUrl: url}

	rss := client.fetchFeed()

	var feed Feed
	xml.Unmarshal(rss, &feed)

	for _, episode := range feed.Episodes {
		filename := sanitize.BaseName(episode.Title) + ".mp4"

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Printf("Downloading file: %s", filename)
			client.downloadFile(episode.EpisodeFile.Url, filename)
		} else {
			log.Printf("File found, skipping: %s", filename)
		}
	}
}
