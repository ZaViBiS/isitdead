// Package checker відповідає за виконання перевірок серверів.
package checker

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// Check виконує перевірку залежно від типу (http або ping)
func Check(checkType, target string) (status string, latency int64) {
	if checkType == "ping" {
		return TcpPing(target)
	}
	return HttpCheck(target)
}

// HttpCheck виконує запит до URL і повертає статус та затримку
func HttpCheck(url string) (status string, latency int64) {
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

// TcpPing виконує спробу підключення до TCP порту (TCP Ping)
func TcpPing(target string) (status string, latency int64) {
	start := time.Now()
	
	// Якщо порт не вказано, додаємо за замовчуванням 80
	if !hasPort(target) {
		target = target + ":80"
	}

	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return fmt.Sprintf("TCP Connection Error: %v", err), elapsed
	}
	defer conn.Close()

	return "Connected", elapsed
}

func hasPort(target string) bool {
	_, _, err := net.SplitHostPort(target)
	return err == nil
}
