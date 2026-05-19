// NOTE: цей увесь файл не повинен тут існувати, треба щось з ним робити, але це вже потім
package bot

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func sendToken(token string, chatID int64) error {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return fmt.Errorf("where is my url?!")
	}
	parsedURL, err := url.Parse(baseURL + "/api/telegram/token")
	if err != nil {
		return err
	}

	q := parsedURL.Query()
	q.Set("token", token)
	q.Set("chat_id", strconv.FormatInt(chatID, 10))
	parsedURL.RawQuery = q.Encode()

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed! url: %s with status %s", parsedURL.String(), resp.Status)
	}
	return nil
}
