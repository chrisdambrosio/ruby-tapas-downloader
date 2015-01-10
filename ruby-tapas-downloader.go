package main

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Channel struct {
	XMLName xml.Name `xml:"rss"`
	Title   string   `xml:"channel>title"`
	Items   []Item   `xml:"channel>item"`
}

type Item struct {
	Title     string    `xml:"title"`
	Enclosure Enclosure `xml:"enclosure"`
}

type Enclosure struct {
	Url string `xml:"url,attr"`
}

func (e Enclosure) Filename() string {
	tokens := strings.Split(e.Url, "/")
	return tokens[len(tokens)-1]
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

func (client Client) downloadFile(item Item) {
	fileUrl := item.Enclosure.Url
	filename := item.Enclosure.Filename()

	out, err := os.Create(filename)
	defer out.Close()

	if err != nil {
		log.Println("Error: error copying file ", filename, " - ", err)
	}

	resp, err := client.Get(fileUrl)
	defer resp.Body.Close()

	if err != nil {
		log.Println("Error: failed to fetch file", fileUrl, " - ", err)
	}

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		log.Fatal("Error: ", err)
	}

	log.Println("Downloaded: ", fileUrl)
}

func main() {
	username := "username"
	password := "password"
	url := "https://rubytapas.dpdcart.com/feed"

	client := Client{username: username, password: password, feedUrl: url}

	rss := client.fetchFeed()

	var c Channel
	xml.Unmarshal(rss, &c)

	client.downloadFile(c.Items[0])
}
