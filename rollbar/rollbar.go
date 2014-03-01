package rollbar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

type Client struct {
	AppName    string
	AppVersion string
	Endpoint   string
	Token      string
}

func (c *Client) Report(message string) error {
	body := c.buildBody()
	data := body["data"].(map[string]interface{})
	data["body"] = c.errorBody(message, 2)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.Endpoint, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 { // 200, 201, 202, etc
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) errorBody(message string, skip int) map[string]interface{} {
	return map[string]interface{}{
		"trace": map[string]interface{}{
			"frames": c.stacktraceFrames(3 + skip),
			"exception": map[string]interface{}{
				"class":   "",
				"message": message,
			},
		},
	}
}

func (c *Client) stacktraceFrames(skip int) []map[string]interface{} {
	frames := []map[string]interface{}{}

	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		fname := "unknown"
		if f != nil {
			fname = f.Name()
		}
		frames = append(frames, map[string]interface{}{
			"filename": file,
			"lineno":   line,
			"method:":  fname,
		})
	}
	return frames
}

func (c *Client) buildBody() map[string]interface{} {
	timestamp := time.Now().UTC().Unix()

	return map[string]interface{}{
		"access_token": c.Token,
		"data": map[string]interface{}{
			"environment": "production",
			"timestamp":   timestamp,
			"platform":    "client",
			"language":    "go",
			"notifier": map[string]interface{}{
				"name":    c.AppName,
				"version": c.AppVersion,
			},
			"client": map[string]interface{}{
				"arch":     runtime.GOARCH,
				"platform": runtime.GOOS,
			},
		},
	}
}
