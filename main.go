package main

import (
	"encoding/xml"
	"flag"
	"github.com/chrisdambrosio/sanitize"
	"github.com/jcelliott/lumber"
	"os"
	"path"
)

const (
	FeedUrl  = "https://rubytapas.dpdcart.com/feed"
	LoginUrl = "https://rubytapas.dpdcart.com/subscriber/login"
)

var (
	logger *lumber.MultiLogger
)

func main() {
	var username = flag.String("u", "", "login username")
	var password = flag.String("p", "", "login password")
	var dir = flag.String("d", "", "target directory")
	flag.Parse()

	logger = lumber.NewMultiLogger()
	logger.AddLoggers(lumber.NewConsoleLogger(lumber.INFO))

	client := NewClient()

	client.Login(*username, *password)

	rss := client.FetchFeed()

	var feed Feed
	xml.Unmarshal(rss, &feed)

	for _, episode := range feed.Episodes {
		episodeDir := sanitize.BaseName(episode.Title)
		episodePath := path.Join(*dir, episodeDir)

		if _, err := os.Stat(episodePath); os.IsNotExist(err) {
			err = os.Mkdir(episodePath, 0755)

			if err != nil {
				logger.Fatal(err.Error())
				os.Exit(1)
			}
		}

		for _, file := range episode.Files() {
			filepath := path.Join(episodePath, file.Name)

			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				logger.Info("Downloading file: %s", path.Join(episodeDir, file.Name))
				client.DownloadFile(file.Url, filepath)
			} else {
				logger.Debug("File found, skipping: %s", path.Join(episodeDir, file.Name))
			}
		}
	}
}
