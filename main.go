package main

import (
	"encoding/xml"
	"flag"
	"github.com/chrisdambrosio/sanitize"
	"log"
	"os"
	"path"
)

const (
	FeedUrl  = "https://rubytapas.dpdcart.com/feed"
	LoginUrl = "https://rubytapas.dpdcart.com/subscriber/login"
)

func main() {
	var username = flag.String("u", "", "login username")
	var password = flag.String("p", "", "login password")
	var dir = flag.String("d", "", "target directory")
	flag.Parse()

	client := NewClient()

	client.Login(*username, *password)

	rss := client.fetchFeed()

	var feed Feed
	xml.Unmarshal(rss, &feed)

	for _, episode := range feed.Episodes {
		episodeDir := path.Join(*dir, sanitize.BaseName(episode.Title))

		if _, err := os.Stat(episodeDir); os.IsNotExist(err) {
			err = os.Mkdir(episodeDir, 0755)

			if err != nil {
				log.Fatal(err)
			}
		}

		for _, file := range episode.Files() {
			filepath := path.Join(episodeDir, file.Name)

			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				log.Printf("Downloading file: %s", filepath)
				client.downloadFile(file.Url, filepath)
			} else {
				log.Printf("File found, skipping: %s", filepath)
			}
		}
	}
}
