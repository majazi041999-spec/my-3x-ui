package service

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/mhsanaei/3x-ui/v2/logger"
)

// PanelService provides business logic for panel management operations.
// It handles panel restart, updates, and system-level panel controls.
type PanelService struct {
	settingService SettingService
}

func (s *PanelService) downloadCustomXrayBinary() {
	url, err := s.settingService.getString("customXrayBinaryURL")
	if err != nil || url == "" {
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		logger.Error("failed to download custom xray binary:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logger.Error("failed to download custom xray binary, status:", resp.Status)
		return
	}

	tmpPath := filepath.Join(os.TempDir(), "xray-custom-download")
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		logger.Error("failed to create temporary custom xray binary:", err)
		return
	}
	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		logger.Error("failed to write temporary custom xray binary:", err)
		return
	}
	tmpFile.Close()

	target := "/usr/local/x-ui/bin/xray"
	if err = os.Rename(tmpPath, target); err != nil {
		logger.Error("failed to replace custom xray binary:", err)
	}
}

func (s *PanelService) RestartPanel(delay time.Duration) error {
	s.downloadCustomXrayBinary()
	p, err := os.FindProcess(syscall.Getpid())
	if err != nil {
		return err
	}
	go func() {
		time.Sleep(delay)
		err := p.Signal(syscall.SIGHUP)
		if err != nil {
			logger.Error("failed to send SIGHUP signal:", err)
		}
	}()
	return nil
}
