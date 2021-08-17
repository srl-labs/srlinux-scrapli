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

## Additional functions
In addition to providing a scrapligo driver for SR Linux, this package contains convenience functions with additional functionality.

### AddSelfSignedServerTLSProfile
The `AddSelfSignedServerTLSProfile` function uses SR Linux ability to generate a self-signed certificate and key and uses those two artifacts to create a server TLS profile that is used for securing gNMI, HTTPS, etc.

The intent behind this function is to simplify TLS certificates provisioning in a lab setting, where self-signed certificates are norm. This function is usually called first thing after SR Linux boot process, to ensure that gNMI service can be enabled.

This function takes the following arguments:

* `*network.Driver` - an initialized network driver, which is a result of `srlinux.NewSRLinuxDriver()`
* `profileName` - a name of the server profile that will be created for the generated key/cert. If empty, the default name `self-signed-tls-profile` will be used.
* `authClient` - a boolean value indicating if the server TLS profile should have authenticate-client option set to true or false.

Note, that the network.Driver that is passed as a first argument should not be opened prior to calling this function, as it will be opened with the specific PTY size inside the function.

## Something doesn't work?
If the driver doesn't work, it is quite likely that the prompt has been changed. SR Linux driver relies on regular expressions defined in `scrapli.go` file for the relevant Privilege Levels. Check if those regular expressions match your prompt, and if not, create an issue or propose a change.