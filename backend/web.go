package backend

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// FetchWebSignatures fetches HTTP/HTTPS title, server header, and body snippet for node detection.
func FetchWebSignatures(ip string, urlStr string) (title, server, body string) {
	targets := []string{}
	if urlStr != "" {
		targets = append(targets, urlStr)
	} else if ip != "" {
		targets = append(targets, "http://"+ip, "https://"+ip)
	}
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for _, target := range targets {
		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TWSNMP-FK/1.0)")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		rawBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		_ = resp.Body.Close()
		rawStr := string(rawBytes)
		server = resp.Header.Get("Server")
		if realm := resp.Header.Get("WWW-Authenticate"); realm != "" {
			if server != "" {
				server += " " + realm
			} else {
				server = realm
			}
		}
		if doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawStr)); err == nil {
			title = strings.TrimSpace(doc.Find("title").First().Text())
		}
		cleanText := strings.Join(strings.Fields(rawStr), " ")
		if len(cleanText) > 8192 {
			cleanText = cleanText[:8192]
		}
		body = cleanText
		if title != "" || server != "" || body != "" {
			break
		}
	}
	return
}
