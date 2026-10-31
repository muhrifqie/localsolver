package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"

	"github.com/xKiian/friendlycaptcha/signals"
	"github.com/xKiian/friendlycaptcha/solver"
	"github.com/xKiian/friendlycaptcha/util"
)

const baseURL = "https://global.frcapi.com/api/v2/captcha"

type SolveContext struct {
	Version   string
	Value     string
	Nonce     uint32
	Timestamp int64
}

type Client struct {
	http         tlsclient.HttpClient
	sitekey      string
	solveContext SolveContext
	widgetID     string
	agentID      string
	sessionID    string
	rseq         int
	seqOffset    int
	announceTime int64
}

func NewClient(sitekey, proxy string) (*Client, error) {
	options := []tlsclient.HttpClientOption{
		tlsclient.WithTimeoutSeconds(30),
		tlsclient.WithClientProfile(profiles.Chrome_133),
		tlsclient.WithCookieJar(tlsclient.NewCookieJar()),
	}
	if proxy != "" {
		options = append(options, tlsclient.WithProxyUrl(proxy))
	}

	session, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("create tls client: %w", err)
	}

	return &Client{
		http:         session,
		sitekey:      sitekey,
		widgetID:     util.WidgetID(),
		agentID:      util.AgentID(),
		sessionID:    util.SessionID(),
		rseq:         0,
		seqOffset:    9,
		announceTime: time.Now().UnixMilli(),
	}, nil
}

func (c *Client) post(endpoint string, body any) ([]byte, error) {
	c.rseq++

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/"+endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header = http.Header{
		"accept":                        {"*/*"},
		"accept-language":               {"en-US,en;q=0.9"},
		"content-type":                  {"application/json"},
		"device-memory":                 {"8"},
		"downlink":                      {"1.6"},
		"dpr":                           {"1"},
		"ect":                           {"4g"},
		"frc-agent-id":                  {c.agentID},
		"frc-agent-load":                {strconv.FormatInt(c.announceTime, 10)},
		"frc-lib":                       {"friendly-captcha-lib@1.3.1"},
		"frc-p-rseq":                    {strconv.Itoa(c.rseq + c.seqOffset)},
		"frc-rseq":                      {strconv.Itoa(c.rseq)},
		"frc-sdk":                       {"friendly-captcha-sdk@0.1.36"},
		"frc-solve-context-nonce":       {strconv.FormatUint(uint64(c.solveContext.Nonce), 10)},
		"frc-solve-context-version":     {c.solveContext.Version},
		"origin":                        {"https://global.frcapi.com"},
		"rtt":                           {"100"},
		"sec-ch-prefers-reduced-motion": {"no-preference"},
		"sec-ch-ua":                     {`"Not A(Brand";v="8", "Chromium";v="146", "Google Chrome";v="146"`},
		"sec-ch-ua-arch":                {`"x86"`},
		"sec-ch-ua-bitness":             {`"64"`},
		"sec-ch-ua-full-version":        {`"146.0.0.0"`},
		"sec-ch-ua-mobile":              {"?0"},
		"sec-ch-ua-model":               {`""`},
		"sec-ch-ua-platform":            {`"Windows"`},
		"sec-ch-ua-platform-version":    {`"19.0.0"`},
		"sec-ch-ua-wow64":               {"?0"},
		"sec-fetch-dest":                {"empty"},
		"sec-fetch-mode":                {"cors"},
		"sec-fetch-site":                {"same-origin"},
		"user-agent":                    {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"},
		"viewport-width":                {"0"},
		http.HeaderOrderKey:             {"content-length", "sec-ch-ua-platform", "frc-agent-load", "frc-solve-context-version", "frc-solve-context-nonce", "sec-ch-ua", "sec-ch-ua-mobile", "frc-agent-id", "frc-sdk", "frc-lib", "frc-rseq", "content-type", "frc-p-rseq", "user-agent", "accept", "origin", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "sec-fetch-storage-access", "accept-encoding", "accept-language", "priority"},
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return data, nil
}

func (c *Client) genSignals() (string, error) {
	return signals.Generate(c.widgetID, c.sessionID, c.solveContext.Nonce)
}

func (c *Client) Activate() (*ActivateResponse, error) {
	sig, err := c.genSignals()
	if err != nil {
		return nil, fmt.Errorf("generate signals: %w", err)
	}

	body := ActivateRequest{
		Sitekey:        c.sitekey,
		WidgetID:       c.widgetID,
		WidgetAnnounce: c.announceTime,
		TS:             time.Now().UnixMilli(),
		Signals:        sig,
		SolveContext:   c.solveContext.Value,
	}

	data, err := c.post("activate", body)
	if err != nil {
		return nil, err
	}

	var resp ActivateResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode activate response: %w", err)
	}
	if !resp.Success && resp.Error != nil {
		return nil, fmt.Errorf("activate failed: %s (%s)", resp.Error.ErrorCode, resp.Error.Detail)
	}

	return &resp, nil
}

func (c *Client) Quote(activation string) (*QuoteResponse, error) {
	sig, err := c.genSignals()
	if err != nil {
		return nil, fmt.Errorf("generate signals: %w", err)
	}

	body := QuoteRequest{
		Sitekey:        c.sitekey,
		WidgetID:       c.widgetID,
		WidgetAnnounce: c.announceTime,
		TS:             time.Now().UnixMilli(),
		Signals:        sig,
		SolveContext:   c.solveContext.Value,
		Activation:     activation,
	}

	data, err := c.post("quote", body)
	if err != nil {
		return nil, err
	}

	var resp QuoteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode quote response: %w", err)
	}
	if !resp.Success && resp.Error != nil {
		return nil, fmt.Errorf("quote failed: %s (%s)", resp.Error.ErrorCode, resp.Error.Detail)
	}

	return &resp, nil
}

func (c *Client) Redeem(activation, quote string, submissions []string, solveDuration int) (*RedeemResponse, error) {
	sig, err := c.genSignals()
	if err != nil {
		return nil, fmt.Errorf("generate signals: %w", err)
	}

	body := RedeemRequest{
		Sitekey:      c.sitekey,
		SolveContext: c.solveContext.Value,
		Activation:   activation,
		Quote:        quote,
		Payment: Payment{
			Entries: []PaymentEntry{
				{
					SolveContext: c.solveContext.Value,
					Submissions:  submissions,
				},
			},
		},
		Signals:            sig,
		WidgetID:           c.widgetID,
		SolveDurationWall:  solveDuration,
		SolveDurationAccum: 0,
		TS:                 time.Now().UnixMilli(),
	}

	data, err := c.post("redeem", body)
	if err != nil {
		return nil, err
	}

	var resp RedeemResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode redeem response: %w", err)
	}
	if !resp.Success && resp.Error != nil {
		return nil, fmt.Errorf("redeem failed: %s (%s)", resp.Error.ErrorCode, resp.Error.Detail)
	}

	return &resp, nil
}

func (c *Client) Verify(token, apiKey string) (*VerifyResponse, error) {
	body := VerifyRequest{
		Response: token,
		Sitekey:  c.sitekey,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal verify request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/siteverify", baseURL), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header = http.Header{
		"content-type": {"application/json"},
		"x-api-key":    {apiKey},
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result VerifyResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode verify response: %w", err)
	}

	return &result, nil
}

func (c *Client) Solve() (string, error) {
	solveCtx, err := c.ScrapeContext()
	if err != nil {
		return "", fmt.Errorf("scrape context: %w", err)
	}
	c.solveContext = solveCtx

	actResp, err := c.Activate()
	if err != nil {
		return "", fmt.Errorf("activate: %w", err)
	}
	activation := actResp.Data.Activation

	quoteResp, err := c.Quote(activation)
	if err != nil {
		return "", fmt.Errorf("quote: %w", err)
	}

	result, err := solver.SolveAll(solver.PuzzleParams{
		Timestamp:   c.solveContext.Timestamp,
		Nonce:       c.solveContext.Nonce,
		SubmitNonce: util.RandNonce(),
		Cost:        quoteResp.Data.Details.Cost,
	})
	if err != nil {
		return "", fmt.Errorf("solve: %w", err)
	}

	redeemResp, err := c.Redeem(activation, quoteResp.Data.Quote, []string{result.Submission}, result.Duration)
	if err != nil {
		return "", fmt.Errorf("redeem: %w", err)
	}

	return redeemResp.Data.RedeemToken, nil
}
