/*
 * MIT License
 *
 * (C) Copyright [2021-2022] Hewlett Packard Enterprise Development LP
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

package algorithmicmac

import (
	"errors"
	"fmt"
	"net"
	"strings"

	base "github.com/Cray-HPE/hms-base"
)

// ErrNotAnAlgorithmicMACAddress is returned when the provided MAC address is not algorithmic
var ErrNotAnAlgorithmicMACAddress = errors.New("provided MAC address is not algorithmic does not start with 0x02")

// ErrUnsupportedBladeController is returned when the provided MAC address is for a blade controller. These are not currently supported by CSM. 
var ErrUnsupportedBladeController = errors.New("no xname exists for blade controllers")

// DecodeMACAddress will decode the algormitmic of a Mountain or Hill BMC to
// determine the xname of the device.
func DecodeMACAddress(macString string) (string, error) {
	if len(macString) == 12 {
		// This MAC address does not include colons, so lets add them
		macString = strings.Join([]string{
			macString[0:2],
			macString[2:4],
			macString[4:6],
			macString[6:8],
			macString[8:10],
			macString[10:12],
		}, ":")
	}

	macAddress, err := net.ParseMAC(macString)
	if err != nil {
		return "", fmt.Errorf("failed to parse MAC address: %w", err)
	}

	// Structure of Algorithmic MAC addresses
	//  02:00:00:00:00:00
	//  || || || || || ||
	//  || || || || || |\- Controller defined 'base' : [0x00 - 0x0F]
	//  || || || || || \-- Sub component 'index'     : [0x00 - 0x0F]
	//  || || || || \----- Slot + offset             : [0x00 - 0xFF]
	//  || || || \-------- Chassis                   : [0x00 - 0xFF]
	//  || \-------------- Rack                      : [0x0000 - 0xFFFF]
	//  \----------------- MAC pool prefix           : 0x02

	// Break apart the MAC address into its components
	macPoolPrefix := macAddress[0]
	rack := int32(macAddress[1]) << 8 + int32(macAddress[2])
	chassis := int32(macAddress[3])
	slotAndOffset := int32(macAddress[4])
	subComponentIndex := int32(macAddress[5] & 0xf0) >> 4
	controllerBase := int32(macAddress[5] & 0x0f)

	if macPoolPrefix != 0x02 {
		return "", ErrNotAnAlgorithmicMACAddress
	}

	// Determine what device this MAC is for:
	deviceType := base.HMSTypeInvalid
	var slot, offset int32
	if (0 <= slotAndOffset && slotAndOffset < 48) {
		// cC - Chassis Controller
		offset = 0
		slot = slotAndOffset

		if slot != 0 {
			return "", fmt.Errorf("slot for ChassisBMC was not 0, but %d", slot)
		}

		deviceType = base.ChassisBMC
	} else if (48 <= slotAndOffset && slotAndOffset < 96) {
		// nC - Node Controller
		offset = 48
		slot = slotAndOffset - offset

		if (7 < slot) {
			return "", fmt.Errorf("slot for NodeBMC was greater than 7, was %d", slot)
		}

		deviceType = base.NodeBMC
	} else if (96 <= slotAndOffset && slotAndOffset < 144) {
		// sS - Switch Controller
		offset = 96
		slot = slotAndOffset - offset

		if (7 < slot) {
			return "", fmt.Errorf("slot for RouterBMC was greater than 7, was %d", slot)
		}

		deviceType = base.RouterBMC
	} else {
		// kC - Blade Controller
		offset = 144
		slot := slotAndOffset - offset

		if (7 < slot) {
			return "", fmt.Errorf("slot for BladeController was greater than 7, was %d", slot)
		}

		// No xname exists for blade controllers currently and do not currently support discovery of them.
		return "", ErrUnsupportedBladeController
	}

	// Index 
	// Only NodeControllers and BladeControllers currently have more than 1 component per blade
	//  TODO: Currently not supporting BladeControllers
	if deviceType != base.NodeBMC && subComponentIndex > 0 {
		return "", fmt.Errorf("device type %s should have only 1 component, but has %d", deviceType, subComponentIndex)
	} 

	// This should be `0` as the base address of 0 is resevered for the primatry managemnet interface of the controller. The Vaults 1-15 are reserved for future use. 
	if controllerBase != 0 {
		return "", fmt.Errorf("controller base for algorthmic MAC was not 0, but %d", controllerBase)
	}

	// Build up the XName
	var xname string
	switch(deviceType) {
	case base.ChassisBMC:
		// xXcCbB
		xname = fmt.Sprintf("x%dc%db%d", rack, chassis, subComponentIndex) 
	case base.NodeBMC:
		// xXcCsSbB
		xname = fmt.Sprintf("x%dc%ds%db%d", rack, chassis, slot, subComponentIndex) 
	case base.RouterBMC:
		// xXcCrRbB
		xname = fmt.Sprintf("x%dc%dr%db%d", rack, chassis, slot, subComponentIndex) 
	default:
		return "", fmt.Errorf("unknown device type: %s", deviceType)
	}

	// Normalize and validate the decoded xname is valid.
	xname = base.NormalizeHMSCompID(xname)
	if !base.IsHMSCompIDValid(xname) {
		return "", fmt.Errorf("generated invalid xname: %s", xname)
	}
	 
	return xname, nil
}