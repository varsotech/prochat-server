package imageproxy

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

func (s *Routes) externalHandler(w http.ResponseWriter, r *http.Request) {
	params := strings.SplitN(r.URL.Path, "/", 2)
	if len(params) != 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	signatureParam := params[0]
	inputUrlParam := params[1]

	// Validate input URL
	signature := s.urlSigner.GenerateSignature(inputUrlParam)
	if signature != signatureParam {
		slog.Debug("Signature mismatch", slog.String("signature", signatureParam))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	u, err := url.Parse(inputUrlParam)
	if err != nil {
		slog.Debug("failed to parse url", slog.String("url", inputUrlParam))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if u.Scheme != "https" {
		slog.Debug("unsupported scheme", slog.String("url", inputUrlParam))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// TODO: Serve from cache
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		slog.Debug("failed to create request", slog.String("url", inputUrlParam))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		slog.Debug("failed to send request", slog.String("url", inputUrlParam))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug("bad status code", slog.String("url", inputUrlParam))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// TODO: Content type?
	//w.Header().Set("Content-Type", "image/png")

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		slog.Debug("failed to write response", slog.String("url", inputUrlParam))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
