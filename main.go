package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

/*
Receives UDP strings from ETC EOS console on port 5000, and subsequently sends HTTP GET to VMIX API on port 8088
(or user defined ports)

Contains some baked-in user defined values:
Examples:
ETC EOS sends "SCN,3" --> triggers VMIX to load Scene index 3 from datasource "Scenes" and then triggers VMIX script "GFXSCENE"
// ** Note Zero Index **

ETC EOS sends "SCENE" --> triggers VMIX script "SCENE" (ie. "Next Scene" for when no scene numbers are defined)
ETC EOS sends "TOP" --> triggers VMIX script "TOP" (ie. "Top of the Show", or 'reset')

*/

// Be sure to define listenPort, cooldownPeriodSeconds, and vmixIP in main()
// Register any additional VMIX scripts with the vmix_register_script() function in main()

var registeredScripts []string
var timeNow time.Time
var timeLastEvent int64
var cooldownPeriodSeconds int64
var vmixAPI_URI string
var vmixIP string

func main() {
	// USER DEFINED VALUES:
	vmix_register_script("SCENE")
	vmix_register_script("TOP")
	listenPort := "5000" // Port to listen for UDP strings from ETC EOS
	vmixPort := "8088"   // Port VMIX API is running on.  IP is defined by runtime CLI argument
	cooldownPeriodSeconds = 6
	// END USER DEFINED VALUES

	timeNow = time.Now()
	fmt.Println(timeNow.Format(time.RFC3339) + " - Starting up")

	vmixIP_ENV := os.Getenv("VMIX_IP")

	if vmixIP_ENV != "" {
		vmixIP = vmixIP_ENV
	} else if len(os.Args) == 2 {
		vmixIP = os.Args[1] // IP at which VMIX resides
	} else {
		fmt.Println("Enter VMIX IP in the format: " + os.Args[0] + " x.x.x.x")
		os.Exit(1)
	}

	vmixAPI_URI = "http://" + vmixIP + ":" + vmixPort + "/api/"

	fmt.Println("VMIX API is at: " + vmixAPI_URI)

	pc, err := net.ListenPacket("udp", ":"+listenPort)
	fmt.Println("Listening on UDP port " + listenPort + "...")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening the socket.")
	}
	defer pc.Close()

	for {
		buf := make([]byte, 24)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to read input buffer.")
		}

		// fmt.Println("tick")

		timeNow = time.Now()
		timeNowUnix := timeNow.Unix()

		fmt.Printf(timeNow.Format(time.RFC3339)+" - packet from: %v", addr)
		fmt.Printf("  %+q\n", buf[:n])
		fmt.Printf("Got UDP string: %+q \n", buf[:n])

		thisPacketContents := string(buf[:n])

		// if we have exceeded the cooldown period since the last event...
		if timeNowUnix > (timeLastEvent + cooldownPeriodSeconds) {

			// Check if we have a comma delimited command
			// example: expecting the format SCN,3 for "SCENE 3"
			if strings.Contains(thisPacketContents, ",") {
				fmt.Println("Caught a comma!")

				// split the comma delimiter
				thisSplitCommand := strings.Split(thisPacketContents, ",")

				// if this is indeed a SCN command...
				if thisSplitCommand[0] == "SCN" {
					if (thisSplitCommand[1] == "") || (len(thisSplitCommand) < 2) {
						fmt.Fprintln(os.Stderr, "No value after comma... skipping...")
						continue
					}

					if !isInteger(thisSplitCommand[1]) {
						fmt.Fprintln(os.Stderr, "Value after comma is not an integer... skipping...")
						continue
					}

					fmt.Printf("This number contained is: ")
					fmt.Println(thisSplitCommand[1])

					// Send VMIX API command for this scene title (this will be a two-step process)
					// clean the trailing crlf
					thisSplitCommand[1] = strings.TrimSuffix(thisSplitCommand[1], "\r\n")
					fmt.Println("Sending VMIX API command DataSourceSelectRow for source Scenes with index: " + thisSplitCommand[1])
					apiCall := vmixAPI_URI + "?Function=DataSourceSelectRow&Value=Scenes," + thisSplitCommand[1]
					fmt.Println("API Call: " + apiCall)
					_, err := http.Get(apiCall)
					timeLastEvent = timeNow.Unix()
					if err != nil {
						fmt.Fprintln(os.Stderr, "Could not issue GET request to VMIX API for scene: "+thisSplitCommand[1])
					}

					time.Sleep(250 * time.Millisecond)
					// wait some time for data source to change	..........
					err = vmix_trigger_script("GFXSCENE")
					if err != nil {
						fmt.Fprintln(os.Stderr, "Could not issue GET request to VMIX API endpoint for GFXSCENE script.")
					}

				} // end SCN,

			} else {
				// We have some other value with no commma delimiter...
				// example "SCENE\r\n"
				// example "TOP\r\n"
				// or even just "TOP"

				// clean the crlf off the input
				thisPacketContents = strings.TrimSuffix(thisPacketContents, "\r\n")
				// We simply pass thru the value if there's a match to registered scripts
				if slices.Contains(registeredScripts, thisPacketContents) {
					err = vmix_trigger_script(thisPacketContents)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Could not issue GET request to VMIX API endpoint for script: "+thisPacketContents)
					}
					timeLastEvent = timeNow.Unix()
				}
			}
		} else { // End OUTER IF (Cooldown)
			fmt.Println("Cooldown not expired.  Ignoring...")
		}
	} // end FOR
} // End MAIN

func vmix_register_script(scriptName string) {
	registeredScripts = append(registeredScripts, scriptName)
}

func vmix_trigger_script(scriptName string) error {
	fmt.Println("Sending VMIX API command for user defined script value: " + scriptName)
	apiCall := vmixAPI_URI + "?Function=ScriptStart&Value=" + scriptName
	fmt.Println("API Call: " + apiCall)
	_, err := http.Get(apiCall)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not issue GET request to VMIX API endpoint user defined script.")
		return err // error
	}
	return nil // success
}

func isInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
