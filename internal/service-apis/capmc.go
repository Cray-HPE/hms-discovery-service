/*
 * MIT License
 *
 * (C) Copyright [2022] Hewlett Packard Enterprise Development LP
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

package service_apis

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

const DefaultCapmcUrl string = "http://cray-capmc/capmc/v1"

type CAPMC struct {
	url         *url.URL
	client      *retryablehttp.Client
	serviceName string
}

type capmcXnameStatus struct {
	Filter string   `json:"filter"`
	Source string   `json:"source"`
	Xnames []string `json:"xnames"`
}

type CapmcStatusResp struct {
	E      int      `json:"e"`
	ErrMsg string   `json:"err_msg"`
	On     []string `json:"on"`
	Off    []string `json:"off"`
}

type capmcXnameOn struct {
	Reason    string   `json:"reason"`
	Xnames    []string `json:"xnames"`
	Continue  bool     `json:"continue"`
}

type CapmcXnameOnResp struct {
	E      int              `json:"e"`
	ErrMsg string           `json:"err_msg"`
	Xnames []CapmcXnameResp `json:"xnames"`
}

type CapmcXnameResp struct {
	E      int    `json:"e"`
	ErrMsg string `json:"err_msg"`
	Xname  string `json:"xname"`
}

// Allocate and initialize new CAPMC struct.
func NewCAPMC(capmcURL string, httpClient *retryablehttp.Client, serviceName string) *CAPMC {
	var err error
	capmc := new(CAPMC)
	if capmc.url, err = url.Parse(capmcURL); err != nil {
		capmc.url, _ = url.Parse(DefaultCapmcUrl)
	} else {
		// Default to using http if not specified
		if len(capmc.url.Scheme) == 0 {
			capmc.url.Scheme = "http"
		}
	}
	capmc.serviceName = serviceName
	if capmc.serviceName == "" {
		capmc.serviceName, err = os.Hostname()
		if err != nil {
			serviceName = "Service_API"
		}
	}

	// Create an httpClient if one was not given
	if httpClient == nil {
		capmc.client = retryablehttp.NewClient()
		capmc.client.HTTPClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		capmc.client.RetryMax = 5
		capmc.client.HTTPClient.Timeout = time.Second * 40
		//turn off the http client loggin!
		tmpLogger := logrus.New()
		tmpLogger.SetLevel(logrus.PanicLevel)
		capmc.client.Logger = tmpLogger
	} else {
		capmc.client = httpClient
	}
	return capmc
}

// Ping CAPMC to see if it is ready.
func (capmc *CAPMC) IsReady() (bool, error) {
	if capmc.url == nil {
		return false, fmt.Errorf("CAPMC struct has no URL")
	}
	uri := capmc.url.String() + "/readiness"
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return false, err
	}
	body, rCode, err := doRequest(req, capmc.client, capmc.serviceName)
	if err != nil {
		return false, err
	}

	if rCode != 204 {
		return false, fmt.Errorf("%s", body)
	}

	return true, nil
}

// Uses get the hardware power state from CAPMC. This gathers power state directly from redfish.
func (capmc *CAPMC) GetXnameStatus(xnames []string) (*CapmcStatusResp, error) {
	var resp CapmcStatusResp
	if capmc.url == nil {
		return nil, fmt.Errorf("CAPMC struct has no URL")
	}
	uri := capmc.url.String() + "/get_xname_status"
	payload := capmcXnameStatus{
		Filter: "show_all",
		Source: "Redfish",
		Xnames: xnames,
	}
	payloadbytes, merr := json.Marshal(payload)
	if merr != nil {
		err := fmt.Errorf("failed to marshal payload: %s", merr)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(payloadbytes))
	if err != nil {
		return nil, err
	}
	body, rCode, err := doRequest(req, capmc.client, capmc.serviceName)
	if err != nil {
		return nil, err
	}

	if rCode != 200 {
		return nil, fmt.Errorf("%s", body)
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Power on target xnames via CAPMC.
func (capmc *CAPMC) XnameOn(xnames []string, reason string) (*CapmcXnameOnResp, error) {
	var resp CapmcXnameOnResp
	if capmc.url == nil {
		return nil, fmt.Errorf("CAPMC struct has no URL")
	}
	if len(xnames) == 0 {
		return nil, fmt.Errorf("CAPMC XnameOn() no xnames for request")
	}
	uri := capmc.url.String() + "/xname_on"
	payload := capmcXnameOn{
		Reason:   reason,
		Xnames:   xnames,
		Continue: true,
	}
	payloadbytes, merr := json.Marshal(payload)
	if merr != nil {
		err := fmt.Errorf("failed to marshal payload: %s", merr)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(payloadbytes))
	if err != nil {
		return nil, err
	}
	body, rCode, err := doRequest(req, capmc.client, capmc.serviceName)
	if err != nil {
		return nil, err
	}

	if rCode != 200 {
		return nil, fmt.Errorf("%s", body)
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}