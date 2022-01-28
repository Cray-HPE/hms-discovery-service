/*
 * MIT License
 *
 * (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"flag"
	"github.com/Cray-HPE/hms-discovery-service/internal/logger"
	svc "github.com/Cray-HPE/hms-discovery-service/internal/service-apis"
)

var (
	// slsURL = flag.String("sls_url", "http://cray-sls", "System Layout Service URL")
	// hsmURL = flag.String("hsm_url", "http://cray-smd", "State Manager URL")
	// capmcURL = flag.String("capmc_url", "http://cray-capmc", "CAPMC URL")

	// discoverRiver = flag.Bool("discover_river", true, "Discover River nodes?")

	// discoverMountain        = flag.Bool("discover_mountain", true, "Discover Mountain nodes?")
	// mountainDiscoveryScript = flag.String("mountain_discovery_script", "mountain_discovery.py",
		// "Location of the script to give Python to run for Mountain discovery.")
	verifyWait = flag.Int("verify_wait", 30,
		"The amount of time in seconds after a power operation to wait before verifying success.")

	// bootNewRiverNodes = flag.Bool("boot_new_river_nodes", false,
		// "When we discover a new node, should we power it on (to make it run REDS)?")

	// httpClient *retryablehttp.Client

	// atomicLevel zap.AtomicLevel
	// logger      *zap.Logger

	// secureStorage       securestorage.SecureStorage
	// hsmCredentialStore  *compcredentials.CompCredStore
	// redsCredentialStore *switches.RedsCredStore
	// pduCredentialStore  *pdu_credential_store.PDUCredentialStore

	// dhcpdnsClient dns_dhcp.DNSDHCPHelper
	
	sls *svc.SLS
	hsm *svc.HSM
	capmc *svc.CAPMC
)

// func main() {
	// // Parse the arguments.
	// flag.Parse()

	// // Setup our http Client
	// httpClient = retryablehttp.NewClient()
	// transport := &http.Transport{
		// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// httpClient.HTTPClient.Transport = transport
	// httpClient.RetryMax = 2
	// httpClient.RetryWaitMax = time.Second * 2
	// httpLogger := http_logger.NewHTTPLogger(logger)
	// httpClient.Logger = httpLogger
	
	// logger.Info("Beginning HMS discovery service.",
		// zap.String("slsURL", *slsURL),
		// zap.String("hsmURL", *hsmURL),
		// zap.String("capmcURL", *hsmURL),
		// zap.Bool("discoverRiver", *discoverRiver),
		// zap.Bool("discoverMountain", *discoverMountain),
		// zap.Bool("bootNewRiverNodes", *bootNewRiverNodes),
		// zap.String("atomicLevel", atomicLevel.String()),
	// )

	// // Loop waiting for the connection to Vault to work.
	// var err error
	// for true {
		// err = setupVault()
		// if err != nil {
			// logger.Error("Unable to setup Vault!", zap.Error(err))
			// time.Sleep(time.Second * 1)
		// } else {
			// break
		// }
	// }

	// startDiscovery()
// }
func main() {
	logger.Init()
	svc.NewSLS(svc.DefaultSlsUrl, nil, "HMDS")
	svc.NewHSM(svc.DefaultHsmUrl, nil, "HMDS")
	svc.NewCAPMC(svc.DefaultCapmcUrl, nil, "HMDS")
}