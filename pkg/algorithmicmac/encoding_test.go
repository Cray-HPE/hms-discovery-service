package algorithmicmac

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"
	"stash.us.cray.com/HMS/hms-smd/pkg/sm"
)

type EncodingTestSuite struct {
	suite.Suite
}

func (suite *EncodingTestSuite) TestChassisBMC() {
	testData := []struct {
		xname      string
		macAddress string
	}{{
		macAddress: "02:03:e8:00:00:00",
		xname:      "x1000c0b0",
	}, {
		macAddress: "02:03:e8:01:00:00",
		xname:      "x1000c1b0",
	}, {
		macAddress: "02:03:e8:02:00:00",
		xname:      "x1000c2b0",
	}, {
		macAddress: "02:03:e8:03:00:00",
		xname:      "x1000c3b0",
	}, {
		macAddress: "02:03:e8:04:00:00",
		xname:      "x1000c4b0",
	}, {
		macAddress: "02:03:e8:05:00:00",
		xname:      "x1000c5b0",
	}, {
		macAddress: "02:03:e8:06:00:00",
		xname:      "x1000c6b0",
	}, {
		macAddress: "02:03:e8:07:00:00",
		xname:      "x1000c7b0",
	}, {
		macAddress: "02:23:28:01:00:00",
		xname:      "x9000c1b0",
	}, {
		macAddress: "02:23:28:03:00:00",
		xname:      "x9000c3b0",
	}}

	for _, test := range testData {
		macAddress, err := EncodeXName(test.xname)
		suite.NoError(err)

		suite.Equal(test.macAddress, macAddress, "encoded MAC address does not match")
	}
}

func (suite *EncodingTestSuite) TestNodeBMC() {
	testData := []struct {
		macAddress string
		xname      string
	}{{
		macAddress: "02:23:28:00:60:00",
		xname:      "x9000c0r0b0",
	}, {
		macAddress: "02:23:28:00:63:00",
		xname:      "x9000c0r3b0",
	}, {
		macAddress: "02:23:28:00:67:00",
		xname:      "x9000c0r7b0",
	}, {
		macAddress: "02:23:28:03:60:00",
		xname:      "x9000c3r0b0",
	}, {
		macAddress: "02:23:28:03:63:00",
		xname:      "x9000c3r3b0",
	}, {
		macAddress: "02:23:28:03:67:00",
		xname:      "x9000c3r7b0",
	}}

	for _, test := range testData {
		macAddress, err := EncodeXName(test.xname)
		suite.NoError(err)

		suite.Equal(test.macAddress, macAddress, "encoded MAC address does not match")
	}
}

func (suite *EncodingTestSuite) TestRouterBMC() {
	testData := []struct {
		macAddress string
		xname      string
	}{{
		macAddress: "02:23:28:00:60:00",
		xname:      "x9000c0r0b0",
	}, {
		macAddress: "02:23:28:00:63:00",
		xname:      "x9000c0r3b0",
	}, {
		macAddress: "02:23:28:00:67:00",
		xname:      "x9000c0r7b0",
	}, {
		macAddress: "02:23:28:03:60:00",
		xname:      "x9000c3r0b0",
	}, {
		macAddress: "02:23:28:03:63:00",
		xname:      "x9000c3r3b0",
	}, {
		macAddress: "02:23:28:03:67:00",
		xname:      "x9000c3r7b0",
	}}

	for _, test := range testData {
		macAddress, err := EncodeXName(test.xname)
		suite.NoError(err)

		suite.Equal(test.macAddress, macAddress, "encoded MAC address does not match")
	}
}

func (suite *EncodingTestSuite) TestAgainstExistingMACData() {
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
			mac, err := EncodeXName(ei.CompID)
			suite.NoError(err)

			suite.Equal(ei.MACAddr, mac, "encoded mac address does not match")
		}
	}
}

func (suite *EncodingTestSuite) TestInvalidXNames() {
	xnames := []string{
		// Valid Xnames, but they do not support algorithmic MACs
		"d0",           // "comptype_cdu",                // dD
		"x1",           // "comptype_cabinet",            // xX
		"x1c1r1t1f1",   //"x1c1r1T1f1",   // "comptype_rtr_tor_fpga",
		"x1c1h1s1",     // "comptype_hl_switch",             // xXcChHsS
		"d0w1",         // "comptype_cdu_mgmt_switch",    // dDwW
		"x1d1",         // "comptype_cab_cdu",            // xXdD
		"x1m1p0",       // "comptype_cab_pdu",            // xXmMpP
		"x1m1i1",       // "comptype_cab_pdu_nic",        // xXmMiI
		"x1m1p0j1",     // "comptype_cab_pdu_outlet",     // xXmMpPjJ DEPRECATED
		"x1m1p0v1",     // "comptype_cab_pdu_pwr_connector",     // xXmMpPvV
		"x1c1",         // "comptype_chassis",                 // xXcC
		"x1c1t0",       // "comptype_cmm_rectifier",           // xXcCtT
		"x1c1f0",       // "comptype_cmm_fpga",                // xXcCfF
		"x1e1",         // "comptype_cec",                     // xXeE
		"x1c1s1",       // "comptype_compmod",                 // xXcCsS
		"x1c1r1",       // "comptype_rtrmod",                  // xXcCrR
		"x1c1s1b1i1",   // "comptype_bmc_nic",                 // xXcCsSbBiI
		"x1c1s1e1",     // "comptype_node_enclosure",          // xXcCsSeE
		"x1c1s1v1",     // "comptype_compmod_power_connector", // xXcCsSvV
		"x1c1s1b1n1",   // "comptype_node",                    // xXcCsSbBnN
		"x1c1s1b1n1p1", // "comptype_node_processor",          // xXcCsSbBnNpP
		"x1c1s1b1n1i1", // "comptype_node_nic",                // xXcCsSbBnNiI
		"x1c1s1b1n1h1", // "comptype_node_hsn_nic",            // xXcCsSbBnNhH
		"x1c1s1b1n1d1", // "comptype_dimm",                    // xXcCsSbBnNdD
		"x1c1s1b1n1a1", // "comptype_node_accel",              // xXcCsSbBnNaA
		"x1c1s1b1f0",   // "comptype_node_fpga",               // xXcCsSbBfF
		"x1c1r1a1",     // "comptype_hsn_asic",                // xXcCrRaA
		"x1c1r1f1",     // "comptype_rtr_fpga",                // xXcCrRfF
		"x1c1r1b1i1",   // "comptype_rtr_bmc_nic",             // xXcCrRbBiI
		"x1c1r1e1",     // "comptype_hsn_board",             // xXcCrReE
		"x1c1r1a1l1",   // "comptype_hsn_link",              // xXcCrRaAlL
		"x1c1r1j1",     // "comptype_hsn_connector",         // xXcCrRjJ
		"x1c1r1j1p1",   // "comptype_hsn_connector_port",    // xXcCrRjJpP
		"x1c1w1",       // "comptype_mgmt_switch",           // xXcCwW
		"x1c1w1j1",     // "comptype_mgmt_switch_connector", // xXcCwWjJ

		// Garbage xnames
		"foo",
		"bar",
		"z1000",
	}

	for _, xname := range xnames {
		mac, err := EncodeXName(xname)
		suite.Empty(mac)
		suite.ErrorIs(err, ErrInvalidXNameType)
	}
}

func TestEncodingTestSuite(t *testing.T) {
	suite.Run(t, new(EncodingTestSuite))
}
