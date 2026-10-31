package api

type APIError struct {
	ErrorCode string `json:"error_code"`
	Detail    string `json:"detail"`
	Data      any    `json:"data"`
}

type ActivateRequest struct {
	Sitekey        string `json:"sitekey"`
	WidgetID       string `json:"widget_id"`
	WidgetAnnounce int64  `json:"widget_announce"`
	TS             int64  `json:"ts"`
	Signals        string `json:"signals"`
	SolveContext   string `json:"solve_context"`
}

type ActivateResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Activation string `json:"activation"`
		Details    struct {
			Mode string `json:"mode"`
		} `json:"details"`
	} `json:"data"`
	Error *APIError `json:"error,omitempty"`
}

type QuoteRequest struct {
	Sitekey        string `json:"sitekey"`
	WidgetID       string `json:"widget_id"`
	WidgetAnnounce int64  `json:"widget_announce"`
	TS             int64  `json:"ts"`
	Signals        string `json:"signals"`
	SolveContext   string `json:"solve_context"`
	Activation     string `json:"activation"`
}

type QuoteResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Quote   string `json:"quote"`
		Details struct {
			Cost int `json:"cost"`
		} `json:"details"`
	} `json:"data"`
	Error *APIError `json:"error,omitempty"`
}

type PaymentEntry struct {
	SolveContext string   `json:"solve_context"`
	Submissions  []string `json:"submissions"`
}

type Payment struct {
	Entries []PaymentEntry `json:"entries"`
}

type RedeemRequest struct {
	Sitekey            string  `json:"sitekey"`
	SolveContext       string  `json:"solve_context"`
	Activation         string  `json:"activation"`
	Quote              string  `json:"quote"`
	Payment            Payment `json:"payment"`
	Signals            string  `json:"signals"`
	WidgetID           string  `json:"widget_id"`
	SolveDurationWall  int     `json:"solve_duration_wall"`
	SolveDurationAccum int     `json:"solve_duration_accumulated"`
	TS                 int64   `json:"ts"`
}

type RedeemResponse struct {
	Success bool `json:"success"`
	Data    struct {
		RedeemToken string `json:"redeem_token"`
		Details     struct {
			ExpiresIn int `json:"expires_in"`
		} `json:"details"`
	} `json:"data"`
	Error *APIError `json:"error,omitempty"`
}

type VerifyRequest struct {
	Response string `json:"response"`
	Sitekey  string `json:"sitekey"`
}

type VerifyResponse struct {
	Success   bool `json:"success"`
	Challenge struct {
		Timestamp string `json:"timestamp"`
		Origin    string `json:"origin"`
		Data      any    `json:"data"`
	} `json:"challenge,omitempty"`
	Error *APIError `json:"error,omitempty"`
}
