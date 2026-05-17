package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://api.dynadot.com/restful/v2"

type Client struct {
	apiKey    string
	secretKey string
	http      *http.Client
}

func New(apiKey, secretKey string) *Client {
	return &Client{
		apiKey:    apiKey,
		secretKey: secretKey,
		http:      &http.Client{},
	}
}

func (c *Client) sign(path, body string) string {
	msg := c.apiKey + "\n" + path + "\n" + "\n" + body
	mac := hmac.New(sha256.New, []byte(c.secretKey))
	mac.Write([]byte(msg))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (c *Client) do(method, path string, body any) ([]byte, error) {
	var bodyBytes []byte
	var bodyStr string

	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyStr = string(bodyBytes)
	}

	req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Signature", c.sign(path, bodyStr))
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	return respBytes, nil
}

type NameserverResponse struct {
	Data struct {
		Nameservers []string `json:"nameservers"`
	} `json:"data"`
}

func (c *Client) GetNameservers(domain string) ([]string, error) {
	resp, err := c.do("GET", "/domains/"+domain+"/nameserver", nil)
	if err != nil {
		return nil, err
	}

	var result NameserverResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result.Data.Nameservers, nil
}

func (c *Client) SetNameservers(domain string, nameservers []string) error {
	_, err := c.do("PUT", "/domains/"+domain+"/nameserver", map[string]any{
		"nameservers": nameservers,
	})
	return err
}
