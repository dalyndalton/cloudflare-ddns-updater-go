package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pterm/pterm"
)

var (
	cloudflareURL string = "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s"
)

func main() {
	pterm.EnableDebugMessages()

	// Read in config file
	if len(os.Args) <= 1 {
		pterm.Error.Println("config file not specified")
		os.Exit(1)
	}

	var config Config
	file, err := os.ReadFile(os.Args[1])
	if err != nil {
		pterm.Fatal.Println("Error reading file:", os.Args[1], err)
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		pterm.Fatal.Println("Error reading file:", os.Args[1], err)
	}

	client := http.Client{
		Timeout: time.Second * 2,
	}

	// Get current ip address
	req, _ := http.NewRequest("GET", "http://ifconfig.me/ip", nil)
	resp, err := client.Do(req)
	if err != nil {
		pterm.Fatal.Println("Error getting ip:", err)
	}
	current_ip, err := io.ReadAll(resp.Body)
	if err != nil {
		pterm.Fatal.Println("Error reading ip:", err)
	}
	resp.Body.Close()

	pterm.Info.Println("Current IP:", string(current_ip))

	content, _ := json.Marshal(CloudflareRequest{
		Content: string(current_ip),
		Name:    config.Name,
		Proxied: false,
		Type:    "A",
		TTL:     1,
		Comment: "DDNS updated by goscript, updated at " + time.Now().Format("2006-01-02 15:04:05"),
	})

	req, _ = http.NewRequest("PUT", fmt.Sprintf(cloudflareURL, config.ZoneID, config.DNSRecordID), bytes.NewReader(content))
	req.Header.Add("Authorization", "Bearer "+config.APIToken)

	resp, err = client.Do(req)
	if err != nil {
		pterm.Fatal.Println("Error updating cloudflare ip:", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		pterm.Fatal.Println("Error reading cloudflare resp:", err)
	}
	resp.Body.Close()
	var response CloudflareResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		pterm.Fatal.Println("Unable to parse cloudflare resp:", err)
	}
	if response.Success {
		pterm.Info.Println("Record updated successfully")
	} else {
		pterm.Warning.Println("Cloudflare returned an error:", string(body))
	}

}

type Config struct {
	Name        string `json:"name"`
	APIToken    string `json:"api_token"`
	ZoneID      string `json:"zone_id"`
	DNSRecordID string `json:"dns_record_id"`
}

type CloudflareRequest struct {
	Content string   `json:"content"`
	Name    string   `json:"name"`
	Proxied bool     `json:"proxied,omitempty"`
	Type    string   `json:"type"`
	Comment string   `json:"comment,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	TTL     int      `json:"ttl,omitempty"`
}

type CloudflareResponse struct {
	Errors   []interface{}    `json:"errors,omitempty"`
	Messages []interface{}    `json:"messages,omitempty"`
	Result   CloudflareResult `json:"result"`
	Success  bool             `json:"success"`
}

type CloudflareResult struct {
	Content    string   `json:"content"`
	Name       string   `json:"name"`
	Proxied    bool     `json:"proxied"`
	Type       string   `json:"type"`
	Comment    string   `json:"comment"`
	CreatedOn  string   `json:"created_on"`
	ID         string   `json:"id"`
	Locked     bool     `json:"locked"`
	Meta       Meta     `json:"meta"`
	ModifiedOn string   `json:"modified_on"`
	Proxiable  bool     `json:"proxiable"`
	Tags       []string `json:"tags"`
	TTL        int64    `json:"ttl"`
	ZoneID     string   `json:"zone_id"`
	ZoneName   string   `json:"zone_name"`
}

type Meta struct {
	AutoAdded bool   `json:"auto_added"`
	Source    string `json:"source"`
}
