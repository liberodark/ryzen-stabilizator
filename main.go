// Copyright 2018 Sergio Correia <sergio@correia.cc>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/klauspost/cpuid"
	"github.com/qrwteyrutiyoup/ryzen-stabilizator/aslr"
	"github.com/qrwteyrutiyoup/ryzen-stabilizator/boosting"
	"github.com/qrwteyrutiyoup/ryzen-stabilizator/c6"
)

const (
	program   = "Ryzen Stabilizator Tabajara"
	copyright = "Copyright (C) 2018 Sergio Correia <sergio@correia.cc>"

	// The family number for Zen processors.
	amdZenFamily = 0x17
)

var (
	version = "unspecified/git version"
)

// rsSettings contains definitions for C6 C-state, processor boosting and
// address space layout randomization (ASLR). All these parameters are "string"
// and accept as values `enabled' and `disabled'.
type rsSettings struct {
	C6       string `toml:"c6"`
	Boosting string `toml:"boosting"`
	ASLR     string `toml:"aslr"`
}

// sanityCheck performs a few checks to be sure we should be running this
// program.
func sanityCheck() error {
	switch {
	// Check if we are running on an AMD processor.
	case cpuid.CPU.VendorID != cpuid.AMD:
		return fmt.Errorf("this is not an AMD processor")
	// Check if it is the right family, 17h (Zen).
	case cpuid.CPU.Family != amdZenFamily:
		return fmt.Errorf("wrong family of AMD processors; expected 23 (17h), got %d", cpuid.CPU.Family)
	// Check if we are running as root.
	case os.Geteuid() != 0:
		return fmt.Errorf("you need to be root to use this program")
	}
	return nil
}

// disableC6 disables C6 C-state.
func disableC6() {
	fmt.Printf("Disabling C6 C-state:   ")
	err := c6.Disable()
	if err != nil {
		fmt.Printf("oops: %v\n", err)
		return
	}
	fmt.Println("SUCCESS")
}

// enableC6 enables C6 C-state.
func enableC6() {
	fmt.Printf("Enabling C6 C-state:   ")
	err := c6.Enable()
	if err != nil {
		fmt.Printf("oops: %v\n", err)
		return
	}
	fmt.Println("SUCCESS")
}

// disableBoosting disables processor boosting.
func disableBoosting() {
	fmt.Printf("Disabling processor boosting:   ")
	err := boosting.Disable()
	if err != nil {
		fmt.Printf("oops: %v\n", err)
		return
	}
	fmt.Println("SUCCESS")
}

// enableBoosting enables processor boosting.
func enableBoosting() {
	fmt.Printf("Enabling processor boosting:   ")
	err := boosting.Enable()
	if err != nil {
		fmt.Printf("oops: %v\n", err)
		return
	}
	fmt.Println("SUCCESS")
}

// disableASLR disables address space layout randomization (ASLR).
func disableASLR() {
	fmt.Printf("Disabling address space layout randomization (ASLR):   ")
	err := aslr.Disable()
	if err != nil {
		fmt.Printf("oops: %v\n", err)
		return
	}
	fmt.Println("SUCCESS")
}

// enableASLR enables address space layout randomization (ASLR).
func enableASLR() {
	fmt.Printf("Enabling address space layout randomization (ASLR):   ")
	err := aslr.Enable()
	if err != nil {
		fmt.Printf("oops: %v\n", err)
		return
	}
	fmt.Println("SUCCESS")
}

// showStatus displays the current status of both C6 C-state and processor
// boosting.
func showStatus() {
	c6Status := "C6 C-state is DISABLED."
	c6Enabled, err := c6.Enabled()
	if err == nil {
		if c6Enabled {
			c6Status = "C6 C-state is ENABLED."
		}
	} else {
		c6Status = fmt.Sprintf("Error while obtaining status of C6 C-state: %v", err)
	}
	fmt.Printf("\n%s\n", c6Status)

	aslrStatus := "ASLR is DISABLED."
	aslrEnabled, err := aslr.Enabled()
	if err == nil {
		if aslrEnabled {
			aslrStatus = "ASLR is ENABLED."
		}
	} else {
		aslrStatus = fmt.Sprintf("Error while obtaining status of ASLR: %v", err)
	}
	fmt.Println(aslrStatus)

	boostingEnabled, err := boosting.Enabled()
	boostingStatus := "Processor boosting is DISABLED."
	if err == nil {
		if boostingEnabled {
			boostingStatus = "Processor boosting is ENABLED."
		}
	} else {
		boostingStatus = fmt.Sprintf("Error while obtaining status of processor boosting: %v", err)
	}
	fmt.Println(boostingStatus)
}

func handleConfigurationFile(configFile string) {
	settings := rsSettings{}

	// Reading and parsing the configuration file provided.
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error: unable to read contents of config file %q: %v.\n", configFile, err)
		return
	}

	if _, err = toml.Decode(string(buf), &settings); err != nil {
		fmt.Printf("Error: problem parsing config file %q: %v.\n\n", configFile, err)
		return
	}

	// Now we perform the actions indicated by the config file.
	fmt.Printf("Config file: %q\n", configFile)
	switch strings.ToLower(settings.Boosting) {
	case "enable":
		enableBoosting()
	case "disable":
		disableBoosting()
	}
	switch strings.ToLower(settings.C6) {
	case "enable":
		enableC6()
	case "disable":
		disableC6()
	}
	switch strings.ToLower(settings.ASLR) {
	case "enable":
		enableASLR()
	case "disable":
		disableASLR()
	}

	// Current status of both C6 C-state and processor boosting.
	showStatus()
}

func main() {
	fmt.Printf("%s %s\n%s\n\n", program, version, copyright)

	err := sanityCheck()
	if err != nil {
		fmt.Printf("Error: %v.\n", err)
		return
	}

	configFilePtr := flag.String("config", "", "ryzen-stabilizator config file")
	enableC6Ptr := flag.Bool("enable-c6", false, "Enable C6 C-state")
	disableC6Ptr := flag.Bool("disable-c6", false, "Disable C6 C-state")
	enableBoostingPtr := flag.Bool("enable-boosting", false, "Enable processor boosting")
	disableBoostingPtr := flag.Bool("disable-boosting", false, "Disable processor boosting")
	enableASLRPtr := flag.Bool("enable-aslr", false, "Enable address space layout randomization (ASLR)")
	disableASLRPtr := flag.Bool("disable-aslr", false, "Disable address space layout randomization (ASLR)")

	flag.Parse()

	// Handle config file with associated profile.
	if *configFilePtr != "" {
		handleConfigurationFile(*configFilePtr)
		return
	}

	// Regular handling of command-line arguments, if we are not using config
	// file with predefined profiles.
	// C6.
	switch {
	case *disableC6Ptr:
		disableC6()
	case *enableC6Ptr:
		enableC6()
	}

	// Boosting.
	switch {
	case *disableBoostingPtr:
		disableBoosting()
	case *enableBoostingPtr:
		enableBoosting()
	}

	// ASLR.
	switch {
	case *disableASLRPtr:
		disableASLR()
	case *enableASLRPtr:
		enableASLR()
	}

	// Current status of both C6 C-state and processor boosting.
	showStatus()
}
