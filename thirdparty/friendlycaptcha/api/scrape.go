package api

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	http "github.com/bogdanfinn/fhttp"
)

var (
	reContextBase64    = regexp.MustCompile(`window\.SOLVE_CONTEXT_BASE64\s*=\s*"([^"]+)"`)
	reContextVersion   = regexp.MustCompile(`window\.SOLVE_CONTEXT_VERSION\s*=\s*(\d+)`)
	reContextTimestamp = regexp.MustCompile(`window\.SOLVE_CONTEXT_TIMESTAMP\s*=\s*(\d+)`)
	reContextNonce     = regexp.MustCompile(`window\.SOLVE_CONTEXT_NONCE\s*=\s*(\d+)`)

	VERSION = "0.1.36"
)

// ScrapeContext fetches the agent iframe page and extracts the solve context variables.
func (c *Client) ScrapeContext() (SolveContext, error) {
	ts := time.Now().UnixMilli()

	agentURL := fmt.Sprintf(
		"%s/agent?origin=%s&sess_id=%s&sess_c=1&comm_id=%s&sdk_v=%s&v=1&agent_id=%s&ts=%d",
		baseURL,
		"https%3A%2F%2Fglobal.frcapi.com",
		c.sessionID,
		c.agentID,
		VERSION,
		c.agentID,
		ts,
	)

	req, err := http.NewRequest(http.MethodGet, agentURL, nil)
	if err != nil {
		return SolveContext{}, fmt.Errorf("create request: %w", err)
	}

	req.Header = http.Header{
		"sec-ch-ua":                 {"\"Chromium\";v=\"146\", \"Not-A.Brand\";v=\"24\", \"Google Chrome\";v=\"146\""},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {"\"Windows\""},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"},
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"sec-fetch-site":            {"cross-site"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-dest":            {"iframe"},
		"sec-fetch-storage-access":  {"active"},
		"accept-encoding":           {"gzip, deflate, br, zstd"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"priority":                  {"u=0, i"},
		http.HeaderOrderKey:         {"sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "sec-fetch-storage-access", "accept-encoding", "accept-language", "priority"},
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return SolveContext{}, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SolveContext{}, fmt.Errorf("read body: %w", err)
	}

	html := string(body)

	b64Match := reContextBase64.FindStringSubmatch(html)
	if b64Match == nil {
		return SolveContext{}, fmt.Errorf("SOLVE_CONTEXT_BASE64 not found in agent page")
	}

	versionMatch := reContextVersion.FindStringSubmatch(html)
	if versionMatch == nil {
		return SolveContext{}, fmt.Errorf("SOLVE_CONTEXT_VERSION not found in agent page")
	}

	tsMatch := reContextTimestamp.FindStringSubmatch(html)
	if tsMatch == nil {
		return SolveContext{}, fmt.Errorf("SOLVE_CONTEXT_TIMESTAMP not found in agent page")
	}

	nonceMatch := reContextNonce.FindStringSubmatch(html)
	if nonceMatch == nil {
		return SolveContext{}, fmt.Errorf("SOLVE_CONTEXT_NONCE not found in agent page")
	}

	timestamp, err := strconv.ParseInt(tsMatch[1], 10, 64)
	if err != nil {
		return SolveContext{}, fmt.Errorf("parse timestamp: %w", err)
	}

	nonce, err := strconv.ParseUint(nonceMatch[1], 10, 32)
	if err != nil {
		return SolveContext{}, fmt.Errorf("parse nonce: %w", err)
	}

	return SolveContext{
		Version:   versionMatch[1],
		Value:     b64Match[1],
		Nonce:     uint32(nonce),
		Timestamp: timestamp,
	}, nil
}
