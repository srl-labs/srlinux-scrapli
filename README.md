# SR Linux driver for Scrapligo
This module adds [Nokia SR Linux](https://www.nokia.com/networks/products/service-router-linux-NOS/) platform support for [scrapligo](https://github.com/scrapli/scrapligo) project.

## How to use?
In order to use SR Linux driver for scrapligo users need to import this module and create a new SR Linux driver. The following example demonstrates how this driver can be used:

```go
package main

import (
	"fmt"

	"github.com/scrapli/scrapligo/driver/base"
	"github.com/scrapli/scrapligo/transport"
	"github.com/srl-labs/srlinux-scrapli"
)

func main() {
	d, err := srlinux.NewSRLinuxDriver(
		"clab-scrapli-srlinux",
		base.WithAuthStrictKey(false),
		base.WithAuthUsername("admin"),
		base.WithAuthPassword("admin"),
		base.WithTransportType(transport.StandardTransportName),
	)

    // uncomment to enable debug log
	// logging.SetDebugLogger(log.Print)

	if err != nil {
		fmt.Printf("failed to create driver; error: %+v\n", err)
		return
	}

	err = d.Open()
	if err != nil {
		fmt.Printf("failed to open driver; error: %+v\n", err)
		return
	}

	fmt.Println("Sending 'show version' command...")
	r, err := d.SendCommand("show version")
	if err != nil {
		fmt.Printf("failed to send commands; error: %+v\n", err)
		return
	}

	fmt.Println(r.Result)

	fmt.Println("Sending configuration commands...")
	configs := []string{
		"system information location scrapligo",
		"commit now",
	}

	_, err = d.SendConfigs(configs)
	if err != nil {
		fmt.Printf("failed to send configs; error: %+v\n", err)
		return
	}

	fmt.Println("Checking that configuration has been applied successfully...")

	r, err = d.SendCommand("info from running /system information location")
	if err != nil {
		fmt.Printf("failed to send commands; error: %+v\n", err)
		return
	}

	fmt.Println(r.Result)

	err = d.Close()
	fmt.Println("closing connection...")
	if err != nil {
		fmt.Printf("failed to close driver; error: %+v\n", err)
	}
}
```

## Doesn't work?
If the driver doesn't work, it is quite likely that the prompt has been changed. SR Linux driver relies on regular expressions defined in `scrapli.go` file for the relevant Privilege Levels. Check if those regular expressions match your prompt, and if not, create an issue or propose a change.