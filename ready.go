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

	mgmtServerRdyCmd  = "info from state system app-management application mgmt_server state | grep running"
	commitCompleteCmd = "info from state system configuration commit 1 status | grep complete"
)

var ()

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
			// two commands are checked, first if the mgmt_server is running
			resp, err := d.SendCommand(mgmtServerRdyCmd)
			if err != nil || resp.Failed != nil {
				time.Sleep(retryTimer)
				continue
			}

			if !strings.Contains(resp.Result, "running") {
				time.Sleep(retryTimer)
				continue
			}

			// and then if the initial commit completes
			resp, err = d.SendCommand(commitCompleteCmd)
			if err != nil || resp.Failed != nil {
				time.Sleep(retryTimer)
				continue
			}

			if !strings.Contains(resp.Result, "complete") {
				time.Sleep(retryTimer)
				continue
			}

			return nil
		}
	}
}
