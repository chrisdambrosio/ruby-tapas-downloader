package main

import (
	"bytes"
	"encoding/xml"
	"github.com/PuerkitoBio/goquery"
	"os"
)

type Feed struct {
	XMLName  xml.Name  `xml:"rss"`
	Title    string    `xml:"channel>title"`
	Episodes []Episode `xml:"channel>item"`
}

type Episode struct {
	Title       string `xml:"title"`
	Description []byte `xml:"description"`
}

func (e Episode) Files() []EpisodeFile {
	reader := bytes.NewReader(e.Description)
	doc, err := goquery.NewDocumentFromReader(reader)

	if err != nil {
		logger.Fatal("Could not parse episode files: " + err.Error())
		os.Exit(1)
	}

	var files []EpisodeFile

	doc.Find("h3:contains('Attached Files')").
		NextFiltered("ul").Find("li a").Each(
		func(i int, s *goquery.Selection) {
			name := s.Text()
			url, _ := s.Attr("href")
			files = append(files, EpisodeFile{Url: url, Name: name})
		})
	return files
}

type EpisodeFile struct {
	Url  string
	Name string
}
