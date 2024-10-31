package exposer

import (
  "net/http"
  "github.com/afrianjunior/statx/internal/pkg"
)

func StatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			(w, pkg.BaseResponse{
				Success: false,
				Message: "failed to parse request body",
				Data:    nil,
			}, http.StatusBadRequest)
			s.jsonError(w, "url parameter is required", http.StatusBadRequest)
			return
		}

		timeRange, err := s.parseTimeRange(r)
		if err != nil {
			s.jsonError(w, fmt.Sprintf("invalid time range: %v", err), http.StatusBadRequest)
			return
		}

		results, err := s.QueryStatus(url, timeRange)
		if err != nil {
			s.jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.jsonResponse(w, results)
	}
}
