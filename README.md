# SR Linux convenience functions for Scrapligo

This module contains convenience functions for working with [Scrapligo](https://github.com/scrapli/scrapligo) and [SR Linux](https://www.nokia.com/networks/products/sr-linux/).

## AddSelfSignedServerTLSProfile

The `AddSelfSignedServerTLSProfile` function uses SR Linux CLI command to generate a self-signed certificate and key. It uses those two artifacts to create a server TLS profile for securing gNMI, HTTPS, etc.

This function intends to simplify TLS certificate provisioning in a lab setting, where self-signed certificates are a norm. This function is usually called the first thing after SR Linux boots to ensure that the gNMI service can be enabled.

This function takes the following arguments:

* `*network.Driver` - an initialized network driver.
* `profileName` - a name of the server profile that will be created for the generated key/cert. If empty, the default name `self-signed-tls-profile` will be used.
* `authClient` - a boolean value indicating if the server TLS profile should have authenticate-client option set to true or false.

Note, that the `network.Driver` should already be opened before calling this function and with the `PtyWidth` sufficient to accommodate certificate/key size.

## WaitSRLMgmtSrvReady

This function blocks until the SR Linux management server fully boots.
