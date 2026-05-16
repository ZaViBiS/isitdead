// package
package api

type registerRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type serverRequest struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	CheckType     string `json:"check_type"`
	CheckInterval int    `json:"check_interval"`
	Timeout       int    `json:"timeout"`
	SlowThreshold int    `json:"slow_threshold"`
}

type dashboardServerResponse struct {
	ID             uint     `json:"id"`
	Name           string   `json:"name"`
	URL            string   `json:"url"`
	CheckType      string   `json:"check_type"`
	CheckInterval  int      `json:"check_interval"`
	Timeout        int      `json:"timeout"`
	SlowThreshold  int      `json:"slow_threshold"`
	CheckCount30d  int64    `json:"check_count_30d"`
	Uptime30d      float64  `json:"uptime_30d"`
	AvgLatency30d  int64    `json:"avg_latency_30d"`
	CurrentStatus  string   `json:"current_status"`
	CurrentLatency int64    `json:"current_latency"`
	HourlyBuckets  []string `json:"hourly_buckets"`
}
