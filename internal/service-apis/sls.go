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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"time"
)

const DefaultSlsUrl string = "http://cray-sls"

type SLS struct {
	url         *url.URL
	client      *retryablehttp.Client
	serviceName string
}

type slsReady struct {
	Ready  bool   `json:"Ready"`
	Reason string `json:"Reason,omitempty"`
	Code   int    `json:"Code,omitempty"`
}

// Allocate and initialize new SLS struct.
func NewSLS(slsURL string, httpClient *retryablehttp.Client, serviceName string) *SLS {
	var err error
	sls := new(SLS)
	if sls.url, err = url.Parse(slsURL); err != nil {
		sls.url, _ = url.Parse(DefaultSlsUrl)
	} else {
		// Default to using http if not specified
		if len(sls.url.Scheme) == 0 {
			sls.url.Scheme = "http"
		}
	}
	sls.serviceName = serviceName
	if sls.serviceName == "" {
		sls.serviceName, err = os.Hostname()
		if err != nil {
			serviceName = "Service_API"
		}
	}

	// Create an httpClient if one was not given
	if httpClient == nil {
		sls.client = retryablehttp.NewClient()
		sls.client.HTTPClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		sls.client.RetryMax = 5
		sls.client.HTTPClient.Timeout = time.Second * 40
		//turn off the http client loggin!
		tmpLogger := logrus.New()
		tmpLogger.SetLevel(logrus.PanicLevel)
		sls.client.Logger = tmpLogger
	} else {
		sls.client = httpClient
	}
	return sls
}

// Ping SLS to see if it is ready.
func (sls *SLS) IsReady() (bool, error) {
	var ready slsReady
	if sls.url == nil {
		return false, fmt.Errorf("SLS struct has no URL")
	}
	uri := sls.url.String() + "/ready"
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return false, err
	}
	body, rCode, err := doRequest(req, sls.client, sls.serviceName)
	if err != nil {
		return false, err
	}

	if rCode != 200 {
		return false, fmt.Errorf("%s", body)
	}

	err = json.Unmarshal(body, &ready)
	if err != nil {
		return false, err
	}

	if !ready.Ready {
		err = fmt.Errorf("%d - %s", ready.Code, ready.Reason)
	}
	return ready.Ready, err
}