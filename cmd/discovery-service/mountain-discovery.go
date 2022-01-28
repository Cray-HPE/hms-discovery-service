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

package main

import(
	"time"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-discovery-service/internal/logger"
	svc "github.com/Cray-HPE/hms-discovery-service/internal/service-apis"
	"github.com/sirupsen/logrus"
)

func doMountainDiscovery() {
	var targetedComponents []string
	targetedHardwareTypes := []string{
		base.Chassis.String(),
		base.ComputeModule.String(),
		base.RouterModule.String(),
	}

	/////////////////
	// GET list of targetedHardwareTypes components from HSM that are not empty or disabled.
	/////////////////
	logger.Log.Infof("Retrieving a list of StateComponents from HSM targeting types %v", targetedHardwareTypes)
	filter := svc.HSMQueryFilter{
		Type:    targetedHardwareTypes,
		State:   []string{"!empty"}, // HSM's POST /State/Components/Query handles parameter negations.
		Enabled: []string{"true"},
	}
	comps, err := hsm.GetStateComponents(filter)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{"ERROR": err}).Error("Could not retrieve StateComponents from HSM")
		return
	}
	logger.Log.Debug("Retrieved StateComponents from HSM")
	for _, comp := range comps.Components {
		targetedComponents = append(targetedComponents, comp.ID)
	}
	if len(comps.Components) == 0 {
		logger.Log.Info("No valid targets, exiting")
		return
	}

	/////////////////
	// GET hardware power status for the targetedComponents from CAPMC
	/////////////////
	logger.Log.Infof("Retrieving xname power state from CAPMC: %v", targetedComponents)
	resp, err := capmc.GetXnameStatus(targetedComponents)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{"ERROR": err}).Error("Could not retrieve power state from CAPMC")
		return
	}
	logger.Log.Debugf("Retrieved hardware power state from CAPMC: %v", *resp)
	if resp.E == 400 {
		logger.Log.WithFields(logrus.Fields{"ERROR": resp.ErrMsg}).Error("Received error from CAPMC")
	}

	/////////////////
	// POWER on targeted components that are 'off'
	/////////////////
	if len(resp.Off) == 0 {
		logger.Log.Info("No components are 'off'. Skipping power on attampt.")
		return
	}
	logger.Log.Infof("Attempting to power on components: %v", resp.Off)
	powerResp, err := capmc.XnameOn(resp.Off, "power on to facilitate mountain discovery")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{"ERROR": err}).Error("Could not issue power command to CAPMC")
		return
	}
	if powerResp.E != 0 {
		logger.Log.WithFields(logrus.Fields{"e": powerResp.E, "err_msg": powerResp.ErrMsg}).Warning("CAPMC power operation return an error.")
	}

	/////////////////
	// Verify POWER state for all of the components we attempted to power on
	/////////////////
	logger.Log.Infof("Sleeping for %d seconds before attempting to verify power state", verifyWait)
	time.Sleep(time.Second * time.Duration(*verifyWait))
	respV, err := capmc.GetXnameStatus(resp.Off)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{"ERROR": err}).Error("Could not retrieve power state from CAPMC")
		return
	}
	if respV.E != 0 {
		logger.Log.WithFields(logrus.Fields{"e": respV.E, "err_msg": respV.ErrMsg}).Error("Received error from CAPMC when verifying power state")
		return
	}
	if len(respV.On) != 0 {
		if len(respV.On) != len(resp.Off) {
			logger.Log.Errorf("Not all xnames could be powered on: %v", respV)
		} else {
			logger.Log.Infof("Power on successfully applied to: %v", respV.On)
		}
	} else {
		logger.Log.Warning("Power on of targeted components has not worked as expected. No components powered on.")
	}
	logger.Log.Debugf("Retrieved hardware power state from CAPMC: %v", *resp)
	if resp.E == 400 {
		logger.Log.WithFields(logrus.Fields{"ERROR": resp.ErrMsg}).Error("Received error from CAPMC")
	}

	/////////////////
	// Report
	/////////////////
	logger.Log.Info("Operation Summary:")
	logger.Log.Infof("Targeted hardware types: \t %v", targetedHardwareTypes)
	logger.Log.Infof("HSM identified count of targeted hardware: \t %d", len(targetedComponents))
	logger.Log.Debugf("HSM identified population of targeted hardware: \t %v", targetedComponents)
	logger.Log.Infof("Count of xnames for power on: \t %d", len(resp.Off))
	logger.Log.Debugf("Targeted population of xnames for power on: \t %v", resp.Off)
	logger.Log.Infof("Count of xnames successfully powered on: \t %d", len(respV.On))
	logger.Log.Debugf("Targeted population of xnames successfully powered on: \t %v", respV.On)
	if len(resp.Off) != len(respV.On) {
		logger.Log.Error("FAILURE: Could not power on xnames: %v", respV.Off)
	}
}