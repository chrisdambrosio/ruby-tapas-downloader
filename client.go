package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

type Client struct {
	Username   string
	Password   string
	FeedUrl    string
	HttpClient *http.Client
}

func NewClient() *Client {
	cookieJar, _ := cookiejar.New(nil)
	client := &Client{
		FeedUrl:    FeedUrl,
		HttpClient: &http.Client{Jar: cookieJar},
	}
	return client
}

func (c *Client) Get(url string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.HttpClient.Do(req)

	return resp, err
}

func (c *Client) Login(username, password string) {
	c.HttpClient.PostForm(LoginUrl,
		url.Values{
			"username": {username},
			"password": {password},
		},
	)
	c.Username = username
	c.Password = password
}

func (c *Client) fetchFeed() []byte {
	resp, err := c.Get(c.FeedUrl)
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

func (c *Client) downloadFile(url, target string) {
	tmpFile := target + ".part"
	out, err := os.Create(tmpFile)
	defer out.Close()

	if err != nil {
		log.Println("Error: error copying file", target, "-", err)
	}

	resp, err := c.Get(url)
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
