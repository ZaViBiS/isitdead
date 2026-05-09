package checker

import (
	"net/http"
	"time"
)

// Check виконує запит до URL і повертає статус та затримку в мілісекундах
func Check(url string) (status string, latency int64) {
	start := time.Now()

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return err.Error(), elapsed
	}
	defer resp.Body.Close()

	return resp.Status, elapsed
}

// TODO: зробити перевірку на сторінки сайту 40x/50x
