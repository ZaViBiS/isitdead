package probe

const SecretHeader = "X-Probe-Secret"

type Target struct {
	Region string
	URL    string
}

type CheckRequest struct {
	CheckType string `json:"check_type"`
	URL       string `json:"url"`
	Timeout   int    `json:"timeout"`
}

type CheckResponse struct {
	Region  string `json:"region"`
	Status  string `json:"status"`
	Latency int64  `json:"latency"`
	Error   string `json:"error,omitempty"`
}
