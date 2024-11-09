package config_monitor

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/afrianjunior/statx/internal/pkg"
)

type listResponse struct {
	Total    int                     `json:"total"`
	Monitors []*pkg.ConfigMonitorDTO `json:"monitors"`
}

func MutationHandler(configMonitorSvc ConfigMonitorService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload pkg.ConfigMonitorDTO
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			pkg.JsonResponse(w, pkg.BaseResponse{
				Success: false,
				Message: "invalid req body",
				Data:    nil,
			}, http.StatusBadRequest)
			return
		}
		id, err := configMonitorSvc.MutateConfigMonitor(r.Context(), &payload)

		pkg.JsonResponse(w, pkg.BaseResponse{
			Success: true,
			Message: "good",
			Data:    id,
		}, http.StatusOK)
	}
}

func ListHandler(configMonitorSvc ConfigMonitorService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GOTHCA")
		monitors, total, err := configMonitorSvc.GetListConfigMonitors(r.Context(), 100, 0)

		if err != nil {
			pkg.JsonResponse(w, pkg.BaseResponse{
				Success: false,
				Message: err.Error(),
				Data:    nil,
			}, http.StatusBadRequest)
			return
		}
		fmt.Println(monitors)
		pkg.JsonResponse(w, pkg.BaseResponse{
			Success: true,
			Message: "good",
			Data: listResponse{
				Total:    total,
				Monitors: monitors,
			},
		}, http.StatusOK)
	}
}
