// Copyright 2021 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package srlinux

import (
	"github.com/scrapli/scrapligo/driver/base"
	"github.com/scrapli/scrapligo/driver/network"
)

// NewSRLinuxDriver returns a driver setup for operation with Nokia SR Linux devices.
func NewSRLinuxDriver(
	host string,
	options ...base.Option,
) (*network.Driver, error) {
	defaultPrivilegeLevels := map[string]*base.PrivilegeLevel{
		"exec": {
			Pattern:        `(?im)^--{[\+\*\s]{1,}running\s}--\[.+?\]--\s*\n[abcd]:\S+#\s*$`,
			Name:           "exec",
			PreviousPriv:   "",
			Deescalate:     "",
			Escalate:       "",
			EscalateAuth:   false,
			EscalatePrompt: ``,
		},
		// configuration privilege level maps to the exclusive config mode on SR Linux
		"configuration": {
			Pattern:        `(?im)^--{[\+\*\s]{1,}candidate\sprivate\s[\-\w\s]+}--\[.+?\]--\s*\n[abcd]:\S+#\s*$`,
			Name:           "configuration",
			PreviousPriv:   "exec",
			Deescalate:     "discard now",
			Escalate:       "enter candidate private",
			EscalateAuth:   false,
			EscalatePrompt: ``,
		},
	}

	defaultFailedWhenContains := []string{
		"Error:",
	}

	const defaultDefaultDesiredPriv = "exec"

	// override default terminal width so that srlinux won't print to much `----`
	options = append(options, base.WithTransportPtySize(140, 80))

	d, err := network.NewNetworkDriver(
		host,
		defaultPrivilegeLevels,
		defaultDefaultDesiredPriv,
		defaultFailedWhenContains,
		SRLinuxOnOpen,
		SRLinuxOnClose,
		options...)

	if err != nil {
		return nil, err
	}

	d.Augments["abortConfig"] = SRLinuxAbortConfig

	return d, nil
}

// SRLinuxOnOpen is a default on open callable for SR Linux.
func SRLinuxOnOpen(d *network.Driver) error {
	err := d.AcquirePriv(d.DefaultDesiredPriv)
	if err != nil {
		return err
	}

	if _, err = d.SendCommand("environment cli-engine type basic", nil); err != nil {
		return err
	}

	if _, err = d.SendCommand("environment complete-on-space false", nil); err != nil {
		return err
	}

	return err
}

// SRLinuxOnClose is a default on close callable for SR Linux.
func SRLinuxOnClose(d *network.Driver) error {
	err := d.AcquirePriv(d.DefaultDesiredPriv)
	if err != nil {
		return err
	}

	err = d.Channel.Write([]byte("quit"), false)
	if err != nil {
		return err
	}

	err = d.Channel.SendReturn()
	if err != nil {
		return err
	}

	return nil
}

// SRLinuxAbortConfig aborts SR Linux configuration session.
func SRLinuxAbortConfig(d *network.Driver) (*base.Response, error) {
	_, err := d.Channel.SendInput("discard /", false, false, -1)

	d.CurrentPriv = "exec"

	return nil, err
}
