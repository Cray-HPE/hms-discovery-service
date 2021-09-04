package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/Cray-HPE/hms-discovery-service/pkg/algorithmicmac"
	base "stash.us.cray.com/HMS/hms-base"
)

func main() {
	// CLI arguments
	macString := flag.String("mac", "", "Algorithmic MAC address to decode")
	flag.Parse()

	if *macString == "" {
		fmt.Println("Error: provided MAC address is empty")
		os.Exit(1)
	}

	// Decode the Algorithmic MAC address
	macAddress, err := net.ParseMAC(*macString)
	if err != nil {
		fmt.Println("Error: failed to parse MAC address -",err)
		os.Exit(1)
	}

	macPoolPrefix :=  macAddress[0]
	rack := int32(macAddress[1]) << 8 + int32(macAddress[2])
	chassis := int32(macAddress[3])
	slotAndOffset := int32(macAddress[4])
	subComponentIndex := int32(macAddress[5] & 0xf0) >> 4
	controllerBase := int32(macAddress[5] & 0x0f)

	if macPoolPrefix != 0x02 {
		fmt.Println("Error: provided MAC address is not algorithmic. Prefix is:", macPoolPrefix)
		os.Exit(1)
	}

	deviceType := base.HMSTypeInvalid
	var slot, offset int32
	if (0 <= slotAndOffset && slotAndOffset < 48) {
		// cC - Chassis Controller
		offset = 0
		slot = slotAndOffset

		if slot != 0 {
			fmt.Println("Slot for ChassisController was not 0, is :", slot)
			os.Exit(1)
		}

		deviceType = base.ChassisBMC
	} else if (48 <= slotAndOffset && slotAndOffset < 96) {
		// nC - Node Controller
		offset = 48
		slot = slotAndOffset - offset

		if (7 < slot) {
			fmt.Println("Slot for NodeController was greater than 7, is :", slot)
			os.Exit(1)
		}

		deviceType = base.NodeBMC
	} else if (96 <= slotAndOffset && slotAndOffset < 144) {
		// sS - Switch Controller
		offset = 96
		slot = slotAndOffset - offset

		if (7 < slot) {
			fmt.Println("Slot for SwitchController was greater than 7, is :", slot)
			os.Exit(1)
		}

		deviceType = base.RouterBMC
	} else {
		// kC - Blade Controller
		offset = 144
		slot := slotAndOffset - offset

		if (7 < slot) {
			fmt.Println("Slot for kC was greater than 7, is :", slot)
			os.Exit(1)
		}

		// No xname exists for blade controllers:
		fmt.Println("Error: No xname exists for blade controllers")
	}

	// Index 
	// Only NodeControllers and BladeControllers currently have more than 1 component per blade
	if !(deviceType == "NodeController" || deviceType == "BladeController") && subComponentIndex > 0 {
		fmt.Println("Error DeviceType: ", deviceType, "should have only 1 component, but has", subComponentIndex)
		os.Exit(1)
	} 

	// This should be `0` as the base address of 0 is resevered for the primage managemnet interface of the controller. The Vaults 1-15 are reserved for future use. 
	if controllerBase != 0 {
		fmt.Println("ControllerBase for algorthmic MAC was not zero, but: ", controllerBase)
		os.Exit(1)
	}

	fmt.Println(macAddress.String())
	fmt.Println("|| || || || || ||")
	fmt.Println(`|| || || || || |\- Controller defined 'base' :`, controllerBase)
	fmt.Println(`|| || || || || \-- Sub component 'index'     :`, subComponentIndex)
	fmt.Println(`|| || || || \----- Slot + offset             : Slot:`, slot, "Offset:", offset, "DeviceType:", deviceType)
	fmt.Println(`|| || || \-------- Chassis                   :`, chassis)
	fmt.Println(`|| \-------------- Rack                      :`, rack)
	fmt.Println(`\----------------- MAC pool prefix           :`, macPoolPrefix)

	xname, err := algorithmicmac.DecodeMACAddress(*macString)
	if err != nil {
		panic(err)
	}

	fmt.Println("Xname:", xname, "Type:", base.GetHMSType(xname))
}