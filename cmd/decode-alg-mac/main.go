package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Cray-HPE/hms-discovery-service/pkg/algorithmicmac"
)

func main() {
		// CLI arguments
		macString := flag.String("mac", "", "Algorithmic MAC address to decode")
		flag.Parse()
	
		if *macString == "" {
			fmt.Println("Error: provided MAC address is empty")
			os.Exit(1)
		}
	
		xname, err := algorithmicmac.DecodeMACAddress(*macString)
		if err != nil {
			panic(err)
		}
	
		fmt.Println(xname)
}