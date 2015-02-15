package main

import (
	"io"
	"io/ioutil"
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

func (c *Client) FetchFeed() []byte {
	resp, err := c.Get(c.FeedUrl)
	defer resp.Body.Close()

	if err != nil {
		logger.Fatal("Error: failed to fetch feed: " + err.Error())
		os.Exit(1)
	}

	rss, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logger.Fatal("Error while reading feed: " + err.Error())
	}

	return rss
}

func (c *Client) DownloadFile(url, target string) {
	tmpFile := target + ".part"
	out, err := os.Create(tmpFile)
	defer out.Close()

	if err != nil {
		logger.Error("Error creating file " + target + ": " + err.Error())
		return
	}

	resp, err := c.Get(url)
	defer resp.Body.Close()

	if err != nil {
		logger.Error("Error fetching file" + url + ": " + err.Error())
		return
	}

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		logger.Error("Error copying file: " + err.Error())
		return
	}

	os.Rename(tmpFile, target)

	logger.Debug("Download Complete: " + target + " from " + url)
}
