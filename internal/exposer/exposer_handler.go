package exposer

import (
	"fmt"
	"net/http"

	"github.com/afrianjunior/statx/internal/pkg"
)

func StatusHandler(exposerSvc ExposerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		url := r.URL.Query().Get("url")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		duration := r.URL.Query().Get("duration")

		if url == "" {
			pkg.JsonResponse(w, pkg.BaseResponse{
				Success: false,
				Message: "url parameter is required",
				Data:    nil,
			}, http.StatusBadRequest)
			return
		}

		timeRange, err := pkg.ParseTimeRange(start, end, duration)
		if err != nil {
			pkg.JsonResponse(w, pkg.BaseResponse{
				Success: false,
				Message: fmt.Sprintf("invalid time range: %v", err),
				Data:    nil,
			}, http.StatusBadRequest)
			return
		}

		results, err := exposerSvc.QueryUpTimeStatus(ctx, url, timeRange)
		if err != nil {
			pkg.JsonResponse(w, pkg.BaseResponse{
				Success: false,
				Message: err.Error(),
				Data:    nil,
			}, http.StatusInternalServerError)
			return
		}

		pkg.JsonResponse(w, pkg.BaseResponse{
			Success: true,
			Message: "good",
			Data:    results,
		}, http.StatusOK)
	}
}
