package stream

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const DefaultURL = "https://stream.wikimedia.org/v2/stream/recentchange"

// wiki streams data using event|id|data
// we only care abotu data
func dataPayload(line []byte) ([]byte, bool) {
	if !bytes.HasPrefix(line, []byte("data:")) {
		return nil, false
	}

	payload := line[len("data:"):]
	payload = bytes.TrimPrefix(payload, []byte(" "))

	if len(payload) == 0 {
		return nil, false
	}

	return payload, true
}

// client streams recentchange events
type Client struct {
	URL       string
	UserAgent string
}

func (c *Client) Run(ctx context.Context, handle func(data []byte)) error {
	backoff := time.Second
	for {
		if err := c.stream(ctx, handle); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Warn("stream disconnected; reconnecting", "err", err, "backoff", backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}
		return nil
	}
}

// connection until breaks or ctx cancelled
func (c *Client) stream(ctx context.Context, handle func(data []byte)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	slog.Info("connected to stream", "url", c.URL)

	// Splits body into lines, up to 1mb
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)

	for sc.Scan() {
		if payload, ok := dataPayload(sc.Bytes()); ok {
			// Scanner reuses its internal buffer so clone
			handle(bytes.Clone(payload))
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("read stream: %w", err)
	}
	return fmt.Errorf("stream closed by server")
}
