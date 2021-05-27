package srlinux

import (
	"fmt"

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
			Pattern:        `(?im)^--{\srunning\s}--\[.+?\]--\s+\n[abcd]:\S+#\s*$`,
			Name:           "exec",
			PreviousPriv:   "",
			Deescalate:     "",
			Escalate:       "",
			EscalateAuth:   false,
			EscalatePrompt: ``,
		},
		// configuration privilege level maps to the exclusive config mode on SR Linux
		"configuration": {
			Pattern:        `(?im)^\*?\(ex\)\[/?\]\n[abcd]:\S+@\S+#\s?$`,
			Name:           "configuration",
			PreviousPriv:   "exec",
			Deescalate:     "quit-config",
			Escalate:       "edit-config exclusive",
			EscalateAuth:   false,
			EscalatePrompt: ``,
		},
		// "configuration-with-path": {
		// 	Pattern:        `(?im)^\*?\(ex\)\[\S{2,}.+\]\n[abcd]:\S+@\S+#\s?$`,
		// 	Name:           "configuration-with-path",
		// 	PreviousPriv:   "configuration",
		// 	Deescalate:     "exit all",
		// 	Escalate:       "",
		// 	EscalateAuth:   false,
		// 	EscalatePrompt: ``,
		// },
	}

	defaultFailedWhenContains := []string{
		"CRITICAL:",
		"MAJOR:",
		"MINOR:",
	}

	const defaultDefaultDesiredPriv = "exec"

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
	fmt.Println("here101")
	err := d.AcquirePriv(d.DefaultDesiredPriv)
	if err != nil {
		return err
	}
	fmt.Println("before sending first command")

	if _, err = d.SendCommand("environment complete-on-space false", nil); err != nil {
		return err
	}
	fmt.Println("after sending first command")

	fmt.Println("finished sending onopen commands")
	return err
}

// SRLinuxOnClose is a default on close callable for SR Linux.
func SRLinuxOnClose(d *network.Driver) error {
	err := d.AcquirePriv(d.DefaultDesiredPriv)
	if err != nil {
		return err
	}

	err = d.Channel.Write([]byte("logout"), false)
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
	if _, err := d.Channel.SendInput("discard /", false, false, -1); err != nil {
		return nil, err
	}

	if _, err := d.Channel.SendInput("exit", false, false, -1); err != nil {
		return nil, err
	}

	_, err := d.Channel.SendInput("quit-config", false, false, -1)

	d.CurrentPriv = "exec"

	return nil, err
}
