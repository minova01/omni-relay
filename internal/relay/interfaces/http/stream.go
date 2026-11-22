package relayhttp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

const maxSSEEventBytes = 1 << 20

type streamUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
}
type streamChunk struct {
	Usage *streamUsage `json:"usage"`
}

func (h *Handler) stream(w http.ResponseWriter, r *http.Request, response *http.Response) {
	if !strings.HasPrefix(strings.ToLower(response.Header.Get("Content-Type")), "text/event-stream") {
		h.fail(w, r, errors.New("upstream did not return SSE"))
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.fail(w, r, errors.New("response does not support flushing"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	reader := bufio.NewReader(response.Body)
	var usage *streamUsage
	var sent, done bool
	for {
		event, err := readSSEEvent(reader, maxSSEEventBytes)
		if err != nil {
			if r.Context().Err() == nil {
				h.logger.WarnContext(r.Context(), "stream interrupted", "error", err, "sent", sent, "usage_available", usage != nil)
				if sent {
					_, _ = io.WriteString(w, "event: error\ndata: {\"error\":\"stream_interrupted\"}\n\n")
					flusher.Flush()
				} else {
					h.fail(w, r, err)
				}
			}
			return
		}
		if len(event) == 0 {
			continue
		}
		payload := sseData(event)
		if len(payload) > 0 && !bytes.Equal(payload, []byte("[DONE]")) {
			var chunk streamChunk
			if json.Unmarshal(payload, &chunk) != nil {
				h.logger.WarnContext(r.Context(), "invalid upstream SSE event")
				if sent {
					_, _ = io.WriteString(w, "event: error\ndata: {\"error\":\"stream_interrupted\"}\n\n")
					flusher.Flush()
				} else {
					h.fail(w, r, errors.New("invalid SSE event"))
				}
				return
			}
			if chunk.Usage != nil {
				usage = chunk.Usage
			}
		}
		if _, err = w.Write(event); err != nil {
			return
		}
		flusher.Flush()
		sent = true
		if bytes.Equal(payload, []byte("[DONE]")) {
			done = true
			break
		}
	}
	h.logger.InfoContext(r.Context(), "stream completed", "done", done, "usage_available", usage != nil)
}

func readSSEEvent(reader *bufio.Reader, limit int) ([]byte, error) {
	var event []byte
	for {
		line, err := reader.ReadSlice('\n')
		if len(event)+len(line) > limit {
			return nil, errors.New("SSE event too large")
		}
		event = append(event, line...)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(bytes.TrimSpace(line), nil) {
			return event, nil
		}
	}
}

func sseData(event []byte) []byte {
	var data []byte
	for _, line := range bytes.Split(event, []byte("\n")) {
		line = bytes.TrimSuffix(line, []byte("\r"))
		if bytes.HasPrefix(line, []byte("data:")) {
			if len(data) > 0 {
				data = append(data, '\n')
			}
			data = append(data, bytes.TrimSpace(line[5:])...)
		}
	}
	return bytes.TrimSpace(data)
}
