package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/go-oauth/oauth"
)

type tweet struct {
	Text string
}

var (
	authClient *oauth.Client
	creds      *oauth.Credentials
)

var (
	conn   net.Conn
	reader io.ReadCloser
)

var httpClient *http.Client
var authSetupOnce sync.Once

func makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	authSetupOnce.Do(func() {
		setupTwitterAuth()
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: dial,
			},
		}
	})
	formEnc := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))
	req.Header.Set("Authorization", authClient.AuthorizationHeader(creds, "POST", req.URL, params))
	return httpClient.Do(req)
}

func setupTwitterAuth() {
	creds = &oauth.Credentials{
		Token:  spTwitterAccessToken,
		Secret: spTwitterAccessSecret,
	}

	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  spTwitterKey,
			Secret: spTwitterSecret,
		},
	}
}

func dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	conn = netc
	return netc, nil
}

func closeConnection() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

func readFromTwitter(votes chan<- vote, options []string, quit <-chan struct{}, wg *sync.WaitGroup) {
	//options, err := loadOptions()
	//if err != nil {
	//	fmt.Println("failed to load options:", err)
	//	return
	//}
	u, _ := url.Parse("https://stream.twitter.com/1.1/statuses/filter.json")
	query := make(url.Values)
	query.Set("track", strings.Join(options, ","))
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(query.Encode()))
	if err != nil {
		fmt.Println("creating filter failed: ", err)
	}
	resp, err := makeRequest(req, query)
	if err != nil {
		fmt.Println("making failed: ", err)
	}
	reader := resp.Body
	decoder := json.NewDecoder(reader)
	fmt.Println(decoder)

	defer wg.Done()

	for {
		var tweet tweet
		if err := decoder.Decode(&tweet); err != nil {
			fmt.Println(err)
			break
		}
		for _, option := range options {
			if strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(option)) {
				select {
				case <-quit:
					return
				case votes <- vote{Vote: option, Tweet: tweet}:
				}
			}
		}
	}
}
