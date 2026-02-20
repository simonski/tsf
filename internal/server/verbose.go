package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxLoggedBodyBytes = 4096

func (s *Server) withVerboseHTTPLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var reqBody []byte
		if r.Body != nil {
			body, _ := io.ReadAll(r.Body)
			reqBody = body
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		fmt.Printf("[HTTP Request] %s %s\n", r.Method, r.URL.String())
		for key, values := range r.Header {
			display := strings.Join(values, ",")
			if strings.EqualFold(key, "Authorization") {
				display = "<redacted>"
			}
			fmt.Printf("> %s: %s\n", key, display)
		}
		printBody("> Body: ", reqBody)

		rec := &verboseResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		fmt.Printf("[HTTP Response] %d %s (%s)\n", rec.status, http.StatusText(rec.status), time.Since(start).Round(time.Millisecond))
		for key, values := range rec.Header() {
			fmt.Printf("< %s: %s\n", key, strings.Join(values, ","))
		}
		printBody("< Body: ", rec.body.Bytes())
		fmt.Println()
	})
}

type verboseResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *verboseResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *verboseResponseWriter) Write(data []byte) (int, error) {
	if w.body.Len() < maxLoggedBodyBytes {
		remain := maxLoggedBodyBytes - w.body.Len()
		if len(data) <= remain {
			w.body.Write(data)
		} else {
			w.body.Write(data[:remain])
		}
	}
	return w.ResponseWriter.Write(data)
}

func printBody(prefix string, body []byte) {
	if len(body) == 0 {
		return
	}
	if len(body) > maxLoggedBodyBytes {
		fmt.Printf("%s%s... [truncated]\n", prefix, string(body[:maxLoggedBodyBytes]))
		return
	}
	fmt.Printf("%s%s\n", prefix, string(body))
}
