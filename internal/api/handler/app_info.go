package handler

import (
	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/metrics"
)

type appInfoResponse struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

func GetAppInfo(c *echo.Context) error {
	version, commit := metrics.GetBuildInfo()
	return ok(c, appInfoResponse{Version: version, Commit: commit})
}
