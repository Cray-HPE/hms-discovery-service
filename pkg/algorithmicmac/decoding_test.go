package algorithmicmac

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	base "stash.us.cray.com/HMS/hms-base"
	"stash.us.cray.com/HMS/hms-smd/pkg/sm"
)

type DecodingTestSuite struct {
	suite.Suite
}

func (suite *DecodingTestSuite) TestChassisBMC() {
	testData := []struct {
		macAddress string
		expectedXname string 
	}{{
		macAddress: "02:03:e8:00:00:00",
		expectedXname: "x1000c0b0",
	}, {
		macAddress: "02:03:e8:01:00:00",
		expectedXname: "x1000c1b0",		
	}, {
		macAddress: "02:03:e8:02:00:00",
		expectedXname: "x1000c2b0",		
	}, {
		macAddress: "02:03:e8:03:00:00",
		expectedXname: "x1000c3b0",
	}, {
		macAddress: "02:03:e8:04:00:00",
		expectedXname: "x1000c4b0",
	}, {
		macAddress: "02:03:e8:05:00:00",
		expectedXname: "x1000c5b0",
	}, {
		macAddress: "02:03:e8:06:00:00",
		expectedXname: "x1000c6b0",
	}, {
		macAddress: "02:03:e8:07:00:00",
		expectedXname: "x1000c7b0",
	}, {
		macAddress: "02:23:28:01:00:00",
		expectedXname: "x9000c1b0",
	}, {
		macAddress: "02:23:28:03:00:00",
		expectedXname: "x9000c3b0",
	}}

	for _, test := range testData {
		macAddress := test.macAddress
		// Lowercase MAC Address with colons
		xname, err := DecodeMACAddress(macAddress) 
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")

		// MAC Address without colons
		xname, err = DecodeMACAddress(strings.ReplaceAll(macAddress, ":", ""))
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")

		// Uppercase MAC address with colons
		xname, err = DecodeMACAddress(strings.ToUpper(macAddress))
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")

		// Uppercase MAC address with colons
		xname, err = DecodeMACAddress(strings.ToUpper(strings.ReplaceAll(macAddress, ":", "")))
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")		
	}
}

func (suite *DecodingTestSuite) TestNodeBMC() {
	testData := []struct {
		macAddress string
		expectedXname string 
	}{{
		macAddress: "02:23:28:00:60:00",
		expectedXname: "x9000c0r0b0",
	}, {
		macAddress: "02:23:28:00:63:00",
		expectedXname: "x9000c0r3b0",
	}, {
		macAddress: "02:23:28:00:67:00",
		expectedXname: "x9000c0r7b0",
	},{
		macAddress: "02:23:28:03:60:00",
		expectedXname: "x9000c3r0b0",
	}, {
		macAddress: "02:23:28:03:63:00",
		expectedXname: "x9000c3r3b0",
	}, {
		macAddress: "02:23:28:03:67:00",
		expectedXname: "x9000c3r7b0",
	}}

	for _, test := range testData {
		macAddress := test.macAddress
		// Lowercase MAC Address with colons
		xname, err := DecodeMACAddress(macAddress) 
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")

		// MAC Address without colons
		xname, err = DecodeMACAddress(strings.ReplaceAll(macAddress, ":", ""))
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")

		// Uppercase MAC address with colons
		xname, err = DecodeMACAddress(strings.ToUpper(macAddress))
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")

		// Uppercase MAC address with colons
		xname, err = DecodeMACAddress(strings.ToUpper(strings.ReplaceAll(macAddress, ":", "")))
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")
	}
}

func (suite *DecodingTestSuite) TestRouterBMC() {
	testData := []struct {
		macAddress string
		expectedXname string 
	}{{
		macAddress: "02:23:28:00:60:00",
		expectedXname: "x9000c0r0b0",
	}, {
		macAddress: "02:23:28:00:63:00",
		expectedXname: "x9000c0r3b0",
	}, {
		macAddress: "02:23:28:00:67:00",
		expectedXname: "x9000c0r7b0",
	},{
		macAddress: "02:23:28:03:60:00",
		expectedXname: "x9000c3r0b0",
	}, {
		macAddress: "02:23:28:03:63:00",
		expectedXname: "x9000c3r3b0",
	}, {
		macAddress: "02:23:28:03:67:00",
		expectedXname: "x9000c3r7b0",
	}}

	for _, test := range testData {
		macAddress := test.macAddress
		// Lowercase MAC Address with colons
		xname, err := DecodeMACAddress(macAddress) 
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")

		// MAC Address without colons
		xname, err = DecodeMACAddress(strings.ReplaceAll(macAddress, ":", ""))
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")

		// Uppercase MAC address with colons
		xname, err = DecodeMACAddress(strings.ToUpper(macAddress))
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")

		// Uppercase MAC address with colons
		xname, err = DecodeMACAddress(strings.ToUpper(strings.ReplaceAll(macAddress, ":", "")))
		suite.NoError(err)

		suite.Equal(test.expectedXname, xname, "decoded xname does not match")
	}
}

func (suite *DecodingTestSuite) TestBladeController() {
	macAddress := "02:03:e8:00:90:00"
	xname, err := DecodeMACAddress(macAddress)
	suite.Empty(xname)
	suite.ErrorIs(err, ErrUnsupportedBladeController)
}


func (suite *DecodingTestSuite) TestInvalidMACAddresses() {
	tests := []struct {
		mac string
		expectedErrorMsg string
	}{{
		mac: "00:40:a6:83:3a44",
		expectedErrorMsg: "failed to parse MAC address: address 00:40:a6:83:3a44: invalid MAC address",
	}, {
		mac: "94:40:c9:5b:da",
		expectedErrorMsg: "failed to parse MAC address: address 94:40:c9:5b:da: invalid MAC address",
	}, {
		mac: "02:03:g8:00:00:00",
		expectedErrorMsg: "failed to parse MAC address: address 02:03:g8:00:00:00: invalid MAC address",
	}}

	for _, test := range tests {
		xname, err := DecodeMACAddress(test.mac)
		suite.Empty(xname)
		suite.EqualError(err, test.expectedErrorMsg)
	}
}

func (suite *DecodingTestSuite) TestingNonAlgorithmicMACs() {
	macs := []string{
		"00:40:a6:83:3a:44",
		"94:40:c9:5b:da:30",
	}

	for _, mac := range macs {
		xname, err := DecodeMACAddress(mac)
		suite.Empty(xname)
		suite.EqualError(err, ErrNotAnAlgorithmicMACAddress.Error())
	}
}

func (suite *DecodingTestSuite) TestInvalidSlotValues() {
	tests := []struct {
		mac string
		expectedErrorString string
	}{{
		// ChassisBMC
		mac: "02:03:e8:00:01:00",
		expectedErrorString: "slot for ChassisBMC was not 0, but 1", 
	}, {
		// RouterBMC
		mac: "02:23:28:00:6f:00",
		expectedErrorString: "slot for RouterBMC was greater than 7, was 15", 
	}, {
		// NodeBMC
		mac: "02:23:28:00:3f:00",
		expectedErrorString: "slot for NodeBMC was greater than 7, was 15", 
	}}

	for _, test :=  range tests {
		xname, err := DecodeMACAddress(test.mac)
		suite.Empty(xname)
		suite.EqualError(err, test.expectedErrorString)
	}
}

func (suite *DecodingTestSuite) TestInvalidSubComponentIndexes() {
	tests := []struct {
		mac string
		expectedErrorString string
	}{{
		// ChasisBMC
		mac: "02:03:e8:00:00:20",
		expectedErrorString: "device type ChassisBMC should have only 1 component, but has 2", 
	}, {
		// RouterBMC
		mac: "02:23:28:00:60:20",
		expectedErrorString: "device type RouterBMC should have only 1 component, but has 2", 
	}}

	for _, test :=  range tests {
		xname, err := DecodeMACAddress(test.mac)
		suite.Empty(xname)
		suite.EqualError(err, test.expectedErrorString)
	}
}

func (suite *DecodingTestSuite) TestInvalidBase() {
	testMacs := []string {
		// ChassisBMC
		"02:03:eb:02:00:01",
		// NodeBMC
		"02:03:eb:00:30:11",
		// RouterBMC
		"02:23:28:01:63:01",
	}

	for _, mac := range testMacs {
		xname, err := DecodeMACAddress(mac)
		suite.Empty(xname)
		suite.EqualError(err, "controller base for algorthmic MAC was not 0, but 1")
	}
}

func (suite *DecodingTestSuite) TestAgainstExistingMACData() {
	files := []string{
		"testdata/hillMacs.json",
		"testdata/mountainMacs.json",
	}

	for _, fileName := range files {
		ethernetInterfacesRaw, err := ioutil.ReadFile(fileName)
		suite.NoError(err)

		var ethernetInterfaces []sm.CompEthInterfaceV2
		err = json.Unmarshal(ethernetInterfacesRaw, &ethernetInterfaces)
		suite.NoError(err)

		for _, ei := range ethernetInterfaces {
			// Try the MAC address with colons
			xname, err := DecodeMACAddress(ei.MACAddr)
			suite.NoError(err)

			suite.Equal(ei.CompID, xname, "decoded xname does not match")
			suite.Equal(ei.Type, base.GetHMSType(xname).String(), "decoded xname HMS type does not match")

			// Try the normalized MAC address without colons
			xname, err = DecodeMACAddress(ei.ID)
			suite.NoError(err)

			suite.Equal(ei.CompID, xname, "decoded xname does not match")
			suite.Equal(ei.Type, base.GetHMSType(xname).String(), "decoded xname HMS type does not match")
		}
	}
}

func TestDecodingTestSuite(t *testing.T) {
    suite.Run(t, new(DecodingTestSuite))
}

