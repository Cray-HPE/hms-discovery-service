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
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"io/ioutil"
	"net/http"

	base "github.com/Cray-HPE/hms-base"
)

// doRequest sends a HTTP request
func doRequest(req *http.Request, client *retryablehttp.Client, serviceName string) ([]byte, int, error) {
	// Error if there is no client defined
	if client == nil {
		return nil, 0, fmt.Errorf("No HTTP Client")
	}

	// Send the request
	base.SetHTTPUserAgent(req, serviceName)
	newRequest, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, 0, err
	}
	newRequest.Header.Set("Content-Type", "application/json")

	rsp, err := client.Do(newRequest)
	if err != nil {
		return nil, 0, err
	}

	// Read the response
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, rsp.StatusCode, nil
}