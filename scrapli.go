// Copyright 2021 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package srlinux

import (
	"github.com/scrapli/scrapligo/driver/base"
	"github.com/scrapli/scrapligo/driver/network"
)

const (
	// PTY dimensions to override scrapli defaults,
	// since width=256 results in too long `---` delimiters in show commands
	defaultPtyWidth  = 140
	defaultPtyHeight = 60
)

// NewSRLinuxDriver global var allows to rewrite the driver initialization function for patched variant in testing
var NewSRLinuxDriver = newSRLinuxDriver

// NewSRLinuxDriver returns a driver setup for operation with Nokia SR Linux devices.
func newSRLinuxDriver(
	host string,
	options ...base.Option,
) (*network.Driver, error) {
	defaultPrivilegeLevels := map[string]*base.PrivilegeLevel{
		"exec": {
			Pattern:        `(?im)^--{(\s\[FACTORY\])?[\+\*\s]{1,}running\s}--\[.+?\]--\s*\n[abcd]:\S+#\s*$`,
			Name:           "exec",
			PreviousPriv:   "",
			Deescalate:     "",
			Escalate:       "",
			EscalateAuth:   false,
			EscalatePrompt: ``,
		},
		// configuration privilege level maps to the exclusive config mode on SR Linux
		"configuration": {
			Pattern:        `(?im)^--{(\s\[FACTORY\])?[\+\*\!\s]{1,}candidate\sprivate\s[\-\w\s]+}--\[.+?\]--\s*\n[abcd]:\S+#\s*$`,
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

	// prepend default pty window size
	options = append([]base.Option{base.WithTransportPtySize(defaultPtyWidth, defaultPtyHeight)}, options...)

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

// NewPatchedSRLinuxDriver returns a new driver and allows to rewrite the default function NewSRLinuxDriver
func NewPatchedSRLinuxDriver(
	host string,
	options ...base.Option,
) (*network.Driver, error) {
	return newSRLinuxDriver(
		host,
		options...,
	)
}
