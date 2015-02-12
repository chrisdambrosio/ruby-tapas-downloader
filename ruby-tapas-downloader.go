package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"github.com/PuerkitoBio/goquery"
	"github.com/chrisdambrosio/sanitize"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
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
		log.Fatal(err)
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

type Client struct {
	Username   string
	Password   string
	feedUrl    string
	httpClient *http.Client
}

func NewClient(username, password string) *Client {
	cookieJar, _ := cookiejar.New(nil)
	client := &Client{
		Username:   username,
		Password:   password,
		feedUrl:    FeedUrl,
		httpClient: &http.Client{Jar: cookieJar},
	}
	return client
}

func (c Client) Get(url string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.httpClient.Do(req)

	return resp, err
}

func (c Client) Login() {
	c.httpClient.PostForm(LoginUrl,
		url.Values{
			"username": {c.Username},
			"password": {c.Password},
		},
	)
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

const (
	FeedUrl  = "https://rubytapas.dpdcart.com/feed"
	LoginUrl = "https://rubytapas.dpdcart.com/subscriber/login"
)

func main() {
	var username = flag.String("u", "", "login username")
	var password = flag.String("p", "", "login password")
	var dir = flag.String("d", "", "target directory")
	flag.Parse()

	client := NewClient(*username, *password)

	client.Login()

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
