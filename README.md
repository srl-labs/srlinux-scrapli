# SR Linux driver for Scrapligo

> **Warning**  
> @carlmontanari made the whole process of adding a custom community platform as easy as writing a yaml file. Check out scrapligo v1 and [srlinux platform definition file](https://github.com/scrapli/scrapligo/blob/main/assets/platforms/nokia_srl.yaml).  
> For that reason, the srlinux network driver that is part of this repo is not needed anymore. Additional functions, though, such as Generate TLS profile, are still maintained and used. 

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

Note, that the network.Driver that is passed as a first argument should already be opened prior to calling this function and with the `PtyWidth` sufficient to accommodate certificate/key size. For example:

```go
d, err := srlinux.NewSRLinuxDriver(
	// some config
)
// setting PTY width to 5k chars to accommodate for long strings of key/cert
d.Transport.BaseTransportArgs.PtyWidth = 5000

transport, _ := d.Transport.Impl.(*transport.System)
transport.SetExecCmd("docker")
transport.SetOpenCmd([]string{"exec", "-u", "root", "-it", contName, "sr_cli", "-d"})

_ := d.Open()

_ := srlinux.WaitSRLMgmtSrv(context.TODO(), d)
```

### WaitSRLMgmtSrv
This function return a nil error when we ensure that SR Linux management server is started and ready to accept configuration commands. The main purpose of this function is to ensure that if the SR Linux node has just been started we won't start configuring it before it is ready.

An example could be defined like follows:

```go
// open a driver
_ := d.Open()

// wait till we can proceed with configs
if err := srlinux.WaitSRLMgmtSrv(context.TODO(), d); err != nil {
	log.Fatal(err)
}
// start sending config commands
_ = srlinux.AddSelfSignedServerTLSProfile(d, tlsProfileName, false)
```

## Known issues and limitations
1. scrapligo doesn't assume that a command in configuration context can switch the session to exec priv. level, although this is what SR Linux commands like `commit save` and `commit now` do.  
    For that reason, if you send configs and up using one of the above mentioned commands, use `d.AcquirePriv()` function to get into the desired privilege level if it doesn't match the previously detected one.

## Something doesn't work?
If the driver doesn't work, it is quite likely that the prompt has been changed. SR Linux driver relies on regular expressions defined in `scrapli.go` file for the relevant Privilege Levels. Check if those regular expressions match your prompt, and if not, create an issue or propose a change.
