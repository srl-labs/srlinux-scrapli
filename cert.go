package srlinux

import (
	"errors"
	"fmt"
	"strings"

	"github.com/scrapli/scrapligo/driver/network"
	"github.com/scrapli/scrapligo/driver/opoptions"
)

const (
	defaultSelfSignedTLSCmd = "tools system tls generate-self-signed"
	keyStartMarker          = "-----BEGIN PRIVATE KEY-----"
	keyEndMarker            = "-----END PRIVATE KEY-----"
	certStartMarker         = "-----BEGIN CERTIFICATE-----"
	certEndMarker           = "-----END CERTIFICATE-----"
	DefaultTLSPRofileName   = "self-signed-tls-profile"
)

// AddSelfSignedServerTLSProfile creates a self-signed certificate with key and creates
// a TLS server profile that references those artifacts. TLS profile name defaults to "self-signed-tls-profile"
// the CLI commands to achieve that are carried with scrapligo *network.Driver that must not be opened prior to calling this func
func AddSelfSignedServerTLSProfile(d *network.Driver, profileName string, authClient bool) error {
	if !d.Transport.IsAlive() {
		return errors.New("device driver should be opened prior calling AddSelfSignedServerTLSProfile function")
	}

	if profileName == "" {
		profileName = DefaultTLSPRofileName
	}

	resp, err := d.SendCommand(defaultSelfSignedTLSCmd)
	if err != nil {
		return fmt.Errorf("failed sending generate-self-signed command: %v", err)
	}

	key, cert, err := extractKeyAndCert(resp.Result)
	if err != nil {
		return err
	}

	err = configureTLSProfile(d, profileName, key, cert)
	return err
}

func extractKeyAndCert(in string) (cert, key string, err error) {
	key, found := getStringInBetween(in, keyStartMarker, keyEndMarker, true)
	if !found {
		return "", "", errors.New("failed to get the key string")
	}
	cert, found = getStringInBetween(in, certStartMarker, certEndMarker, true)
	if !found {
		return "", "", errors.New("failed to get the cert string")
	}
	return key, cert, nil
}

func configureTLSProfile(d *network.Driver, profileName, key, cert string) error {

	configs := []string{
		fmt.Sprintf("set / system tls server-profile %s", profileName),
		fmt.Sprintf("set / system tls server-profile %s authenticate-client false", profileName),
	}

	_, err := d.SendConfigs(configs)
	if err != nil {
		return err
	}
	// key and cert are send outside of sendconfigs, because it was not working properly with `eager` option
	_, err = d.SendConfig(fmt.Sprintf("set / system tls server-profile %s key \"%s\"", profileName, key),
		opoptions.WithEager(),
	)
	if err != nil {
		return err
	}
	_, err = d.SendConfig(fmt.Sprintf("set / system tls server-profile %s certificate \"%s\"", profileName, cert),
		opoptions.WithEager())
	if err != nil {
		return err
	}

	_, err = d.SendConfig("commit save")

	return err
}

// GetStringInBetween returns a string between the start/end markers with markers either included or excluded
func getStringInBetween(str string, start, end string, include bool) (result string, found bool) {
	// start index
	sidx := strings.Index(str, start)
	if sidx == -1 {
		return "", false
	}

	// forward start index if we don't want to include the markers
	if !include {
		sidx += len(start)
	}

	newS := str[sidx:]

	// end index
	eidx := strings.Index(newS, end)
	if eidx == -1 {
		return "", false
	}
	// to include the end marker, increment the end index up till its length
	if include {
		eidx += len(end)
	}

	return newS[:eidx], true
}
