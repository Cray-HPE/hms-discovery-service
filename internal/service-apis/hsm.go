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

	base "github.com/Cray-HPE/hms-base"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

const DefaultHsmUrl string = "http://cray-smd/hsm/v2"

type HSM struct {
	url         *url.URL
	client      *retryablehttp.Client
	serviceName string
}

type HSMQueryFilter struct {
	ComponentIDs []string `json:"ComponentIDs"`
	Type         []string `json:"type"`
	State        []string `json:"state"`
	Enabled      []string `json:"enabled"`
}

// Allocate and initialize new HSM struct.
func NewHSM(hsmURL string, httpClient *retryablehttp.Client, serviceName string) *HSM {
	var err error
	hsm := new(HSM)
	if hsm.url, err = url.Parse(hsmURL); err != nil {
		hsm.url, _ = url.Parse(DefaultSlsUrl)
	} else {
		// Default to using http if not specified
		if len(hsm.url.Scheme) == 0 {
			hsm.url.Scheme = "http"
		}
	}
	hsm.serviceName = serviceName
	if hsm.serviceName == "" {
		hsm.serviceName, err = os.Hostname()
		if err != nil {
			serviceName = "Service_API"
		}
	}

	// Create an httpClient if one was not given
	if httpClient == nil {
		hsm.client = retryablehttp.NewClient()
		hsm.client.HTTPClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		hsm.client.RetryMax = 5
		hsm.client.HTTPClient.Timeout = time.Second * 40
		//turn off the http client loggin!
		tmpLogger := logrus.New()
		tmpLogger.SetLevel(logrus.PanicLevel)
		hsm.client.Logger = tmpLogger
	} else {
		hsm.client = httpClient
	}
	return hsm
}

// Ping HSM to see if it is ready.
func (hsm *HSM) IsReady() (bool, error) {
	if hsm.url == nil {
		return false, fmt.Errorf("HSM struct has no URL")
	}
	uri := hsm.url.String() + "/service/ready"
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return false, err
	}
	body, rCode, err := doRequest(req, hsm.client, hsm.serviceName)
	if err != nil {
		return false, err
	}

	if rCode != 200 {
		return false, fmt.Errorf("%s", body)
	}

	return true, err
}

// Uses 'POST /State/Components/Query' instead of `GET /State/Components` because the former handles parameter negation (i.e. State!=empty).
func (hsm *HSM) GetStateComponents(filter HSMQueryFilter) (*base.ComponentArray, error) {
	var comps base.ComponentArray
	if hsm.url == nil {
		return nil, fmt.Errorf("HSM struct has no URL")
	}
	uri := hsm.url.String() + "/State/Components/Query"
	if len(filter.ComponentIDs) == 0 {
		// HSM doesn't accept an empty ID array. "s0" means everything.
		filter.ComponentIDs = []string{"s0"}
	}
	payload, merr := json.Marshal(filter)
	if merr != nil {
		err := fmt.Errorf("failed to marshal filter: %s", merr)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	body, rCode, err := doRequest(req, hsm.client, hsm.serviceName)
	if err != nil {
		return nil, err
	}

	if rCode != 200 {
		return nil, fmt.Errorf("%s", body)
	}

	err = json.Unmarshal(body, &comps)
	if err != nil {
		return nil, err
	}

	return &comps, nil
}
