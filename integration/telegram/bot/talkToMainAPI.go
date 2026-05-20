// NOTE: цей увесь файл не повинен тут існувати, треба щось з ним робити, але це вже потім
package bot

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var (
	ErrMissingBaseURL = errors.New("BASE_URL is not configured")
	ErrLinkRejected   = errors.New("telegram link was rejected")
)

func sendToken(token string, chatID int64) error {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return ErrMissingBaseURL
	}
	parsedURL, err := url.Parse(baseURL + "/api/telegram/token/" + strconv.FormatInt(chatID, 10) + "/" + token)
	if err != nil {
		return err
	}

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
			return ErrLinkRejected
		}
		return fmt.Errorf("telegram link request failed with status %s", resp.Status)
	}
	return nil
}
