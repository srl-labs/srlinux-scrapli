package srlinux

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/scrapli/scrapligo/driver/network"
)

const (
	readyTimeout = 3 * time.Minute
	retryTimer   = 2 * time.Second

	mgmtServerRdyCmd = "info from state system app-management application mgmt_server state | grep running"
	// readyForConfigCmd checks the output of a file on srlinux which will be populated once the mgmt server is ready to accept config
	readyForConfigCmd = "file cat /etc/opt/srlinux/devices/app_ephemeral.mgmt_server.ready_for_config"
)

// WaitSRLMgmtSrvReady returns when the node boot sequence reaches the stage when it is ready to accept config commands
// returns an error if not ready by readyTimeout.
func WaitSRLMgmtSrvReady(ctx context.Context, d *network.Driver) error {
	ctx, cancel := context.WithTimeout(ctx, readyTimeout)
	defer cancel()

	var err error

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for SR Linux node %s to boot: %v", d.Transport.GetHost(), err)
		default:
			// first check if the mgmt_server app is running
			resp, err := d.SendCommand(mgmtServerRdyCmd)
			if err != nil || resp.Failed != nil {
				time.Sleep(retryTimer)
				continue
			}

			if !strings.Contains(resp.Result, "running") {
				time.Sleep(retryTimer)
				continue
			}

			// then check if mgmt server is fully initialized and ready to accept configs
			resp, err = d.SendCommand(readyForConfigCmd)
			if err != nil || resp.Failed != nil {
				time.Sleep(retryTimer)
				continue
			}

			if !strings.Contains(resp.Result, "loaded initial configuration") {
				time.Sleep(retryTimer)
				continue
			}

			return nil
		}
	}
}
