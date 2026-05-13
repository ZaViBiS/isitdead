// Package checker відповідає за виконання перевірок серверів.
package checker

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	maxLinkCheckPages        = 20
	maxLinkCheckReferences   = 120
	maxBrokenLinkExamples    = 5
	monitorUserAgent         = "isitdead.cc monitor/1.0 (+https://isitdead.cc)"
	defaultConnectionTimeout = 10 * time.Second
)

type linkReference struct {
	target string
	source string
	attr   string
}

type brokenReference struct {
	target string
	source string
	status string
}

// Check виконує перевірку залежно від типу.
func Check(checkType, target string, timeoutSeconds int) (status string, latency int64) {
	timeout := connectionTimeout(timeoutSeconds)
	switch checkType {
	case "ping":
		return TCPPing(target, timeout)
	case "links":
		return LinkCheck(target, timeout)
	default:
		return HttpCheck(target, timeout)
	}
}

// HttpCheck виконує запит до URL і повертає статус та затримку
func HttpCheck(url string, timeout time.Duration) (status string, latency int64) {
	start := time.Now()

	client := http.Client{
		Timeout: timeout,
	}

	req, err := newMonitorRequest(http.MethodGet, url)
	if err != nil {
		return err.Error(), time.Since(start).Milliseconds()
	}

	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return err.Error(), elapsed
	}
	defer resp.Body.Close()

	return resp.Status, elapsed
}

// LinkCheck crawls a site entry page and reports broken pages/assets with sources.
func LinkCheck(rawURL string, timeout time.Duration) (status string, latency int64) {
	start := time.Now()

	baseURL, err := parseHTTPURL(rawURL)
	if err != nil {
		return err.Error(), time.Since(start).Milliseconds()
	}

	client := http.Client{Timeout: timeout}
	visitedPages := map[string]bool{}
	queuedPages := map[string]bool{baseURL.String(): true}
	queue := []*url.URL{baseURL}
	checkedTargets := map[string]string{}
	broken := []brokenReference{}
	referenceCount := 0

	for len(queue) > 0 && len(visitedPages) < maxLinkCheckPages && referenceCount < maxLinkCheckReferences {
		pageURL := queue[0]
		queue = queue[1:]
		pageKey := pageURL.String()
		if visitedPages[pageKey] {
			continue
		}
		visitedPages[pageKey] = true

		req, err := newMonitorRequest(http.MethodGet, pageKey)
		if err != nil {
			broken = append(broken, brokenReference{target: pageKey, source: "crawl", status: err.Error()})
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			broken = append(broken, brokenReference{target: pageKey, source: "crawl", status: err.Error()})
			continue
		}

		if resp.StatusCode >= http.StatusBadRequest {
			broken = append(broken, brokenReference{target: pageKey, source: "crawl", status: resp.Status})
			resp.Body.Close()
			continue
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			resp.Body.Close()
			continue
		}

		refs, err := extractLinkReferences(resp.Body, pageURL)
		resp.Body.Close()
		if err != nil {
			broken = append(broken, brokenReference{target: pageKey, source: "crawl", status: err.Error()})
			continue
		}

		for _, ref := range refs {
			if referenceCount >= maxLinkCheckReferences {
				break
			}
			referenceCount++

			refURL, err := parseHTTPURL(ref.target)
			if err != nil {
				continue
			}

			targetStatus, ok := checkedTargets[refURL.String()]
			if !ok {
				targetStatus = checkReference(&client, refURL.String())
				checkedTargets[refURL.String()] = targetStatus
			}

			if targetStatus != "" {
				broken = append(broken, brokenReference{
					target: refURL.String(),
					source: ref.source,
					status: targetStatus,
				})
			}

			if targetStatus == "" && ref.attr == "href" && sameHost(baseURL, refURL) && !visitedPages[refURL.String()] && !queuedPages[refURL.String()] {
				queuedPages[refURL.String()] = true
				queue = append(queue, refURL)
			}
		}
	}

	elapsed := time.Since(start).Milliseconds()
	if len(broken) == 0 {
		return "200 OK", elapsed
	}

	return formatBrokenReferences(broken), elapsed
}

func parseHTTPURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("missing URL host")
	}
	parsed.Fragment = ""
	return parsed, nil
}

func extractLinkReferences(body io.Reader, pageURL *url.URL) ([]linkReference, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	refs := []linkReference{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				switch attr.Key {
				case "href", "src", "action", "poster":
					if resolved, ok := resolveReference(pageURL, attr.Val); ok {
						refs = append(refs, linkReference{target: resolved, source: pageURL.String(), attr: attr.Key})
					}
				case "srcset":
					for _, candidate := range parseSrcset(attr.Val) {
						if resolved, ok := resolveReference(pageURL, candidate); ok {
							refs = append(refs, linkReference{target: resolved, source: pageURL.String(), attr: attr.Key})
						}
					}
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	return refs, nil
}

func resolveReference(pageURL *url.URL, rawRef string) (string, bool) {
	rawRef = strings.TrimSpace(rawRef)
	if rawRef == "" || strings.HasPrefix(rawRef, "#") {
		return "", false
	}

	lower := strings.ToLower(rawRef)
	if strings.HasPrefix(lower, "mailto:") ||
		strings.HasPrefix(lower, "tel:") ||
		strings.HasPrefix(lower, "javascript:") ||
		strings.HasPrefix(lower, "data:") {
		return "", false
	}

	refURL, err := url.Parse(rawRef)
	if err != nil {
		return "", false
	}

	resolved := pageURL.ResolveReference(refURL)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", false
	}
	resolved.Fragment = ""
	return resolved.String(), true
}

func parseSrcset(srcset string) []string {
	parts := strings.Split(srcset, ",")
	candidates := make([]string, 0, len(parts))
	for _, part := range parts {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) > 0 {
			candidates = append(candidates, fields[0])
		}
	}
	return candidates
}

func checkReference(client *http.Client, target string) string {
	req, err := newMonitorRequest(http.MethodHead, target)
	if err != nil {
		return err.Error()
	}

	resp, err := client.Do(req)
	if err == nil {
		statusCode := resp.StatusCode
		resp.Body.Close()
		if statusCode < http.StatusBadRequest {
			return ""
		}
	}

	req, err = newMonitorRequest(http.MethodGet, target)
	if err != nil {
		return err.Error()
	}

	resp, err = client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return resp.Status
	}
	return ""
}

func newMonitorRequest(method, target string) (*http.Request, error) {
	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", monitorUserAgent)
	return req, nil
}

func sameHost(a, b *url.URL) bool {
	return strings.EqualFold(a.Host, b.Host)
}

func formatBrokenReferences(broken []brokenReference) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Broken links: %d", len(broken))
	for i, ref := range broken {
		if i >= maxBrokenLinkExamples {
			fmt.Fprintf(&b, "; +%d more", len(broken)-i)
			break
		}
		fmt.Fprintf(&b, "; %s %s from %s", ref.status, ref.target, ref.source)
	}
	return b.String()
}

func connectionTimeout(timeoutSeconds int) time.Duration {
	if timeoutSeconds <= 0 {
		return defaultConnectionTimeout
	}
	return time.Duration(timeoutSeconds) * time.Second
}

// TCPPing виконує спробу підключення до TCP порту.
func TCPPing(target string, timeout time.Duration) (status string, latency int64) {
	start := time.Now()

	// Якщо порт не вказано, додаємо за замовчуванням 80
	if !hasPort(target) {
		target = target + ":80"
	}

	conn, err := net.DialTimeout("tcp", target, timeout)
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
