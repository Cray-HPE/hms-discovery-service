package algorithmicmac

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	base "stash.us.cray.com/HMS/hms-base"
)

// ErrInvalidXNameType is returned when a algorthmic MAC address cannot be generated for a
// specified XName
var ErrInvalidXNameType = errors.New("invalid xname type")

// MACPrefixString is the expected algorithmic MAC address prefix
var MACPrefixString = "02"

// MACPrefix is the expected algorithmic MAC address prefix
var MACPrefix = 0x02

var chassisBMCRegex = regexp.MustCompile("x([0-9]+)c([0-9]+)b0")
var nodeBMCRegex = regexp.MustCompile("x([0-9]+)c([0-9]+)s([0-9]+)b([0-9]+)")
var routerBMCRegex = regexp.MustCompile("x([0-9]+)c([0-9]+)r([0-9]+)b([0-9]+)")

// EncodeXName will encode the provided xname into its correndsponding
// algorithmic MAC address form
func EncodeXName(xname string) (string , error) {
	switch (base.GetHMSType(xname)) {
	case base.ChassisBMC:
		matches := chassisBMCRegex.FindStringSubmatch(xname)
		if len(matches) != 3 {
			return "", errors.New("failed to parse ChassisBMC xname")
		}

		rack, err := strconv.Atoi(matches[1])
		if err != nil {
			return "", err
		}

		chassis, err := strconv.Atoi(matches[2])
		if err != nil {
			return "", err
		}


		return GenerateMACChassisBMC(MACPrefixString, rack, chassis), nil	

	case base.NodeBMC:
		matches := nodeBMCRegex.FindStringSubmatch(xname)
		if len(matches) != 5 {
			return "", errors.New("failed to parse NodeBMC xname")
		}
		
		rack, err := strconv.Atoi(matches[1])
		if err != nil {
			return "", err
		}

		chassis, err := strconv.Atoi(matches[2])
		if err != nil {
			return "", err
		}

		slot, err := strconv.Atoi(matches[3])
		if err != nil {
			return "", err
		}

		subComponentIndex, err := strconv.Atoi(matches[4])
		if err != nil {
			return "", err
		}

		return GenerateMACNodeBMC(MACPrefixString, rack, chassis, slot, subComponentIndex), nil	

	case base.RouterBMC:
		matches := routerBMCRegex.FindStringSubmatch(xname)
		if len(matches) != 5 {
			return "", errors.New("failed to parse RouterBMC xname")
		}

		rack, err := strconv.Atoi(matches[1])
		if err != nil {
			return "", err
		}

		chassis, err := strconv.Atoi(matches[2])
		if err != nil {
			return "", err
		}

		slot, err := strconv.Atoi(matches[3])
		if err != nil {
			return "", err
		}

		return GenerateMACRouterBMC(MACPrefixString, rack, chassis, slot), nil	
	default:
		return "", ErrInvalidXNameType
	}
}

// GenerateMAC will generate a generic algorthmic MAC address
func GenerateMAC(mp string, rack int, chassis int, slt int, idx int) string {
	return fmt.Sprintf("%s:%02x:%02x:%02x:%02x:%02x", mp,
		(rack>>8)&0xFF, rack&0xFF, chassis&0xFF, slt&0xFF, (idx<<4)&0xFF)
}

// GenerateMACNodeBMC will generate an algorthmic MAC address for a Node BMC
func GenerateMACNodeBMC(mp string, rack int, chassis int, slt int, idx int) string {
	return GenerateMAC(mp, rack, chassis, slt+48, idx)
}

// GenerateMACRouterBMC will generate an algorthmic MAC address for a Router BMC
func GenerateMACRouterBMC(mp string, rack int, chassis int, slt int) string {
	return GenerateMAC(mp, rack, chassis, slt+96, 0)

}

// GenerateMACChassisBMC will generate an algorthmic MAC address for a Chassis BMC
func GenerateMACChassisBMC(mp string, rack int, chassis int) string {
	return GenerateMAC(mp, rack, chassis, 0, 0)
}