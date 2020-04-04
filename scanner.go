package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lair-framework/go-nmap"
)

type report struct {
	ipRange   string
	openPorts map[int][]*Host
}

type scanner struct {
	ipRange string
	nmapCmd string
}

type Host struct {
	IpAddress   string
	PortDetails map[int]*Port
}

type Port struct {
	Port           int
	State          string
	DownTime       time.Duration
	ServiceName    string
	ServiceVersion string
	downPoint      time.Time
}

func (scnr *scanner) scanEmit(reportUpdates chan report) {
	storedHosts := map[string]*Host{}
	for true {
		cmds := strings.Split(scnr.nmapCmd, " ")
		if err := exec.Command(cmds[0], cmds[1:]...).Run(); err != nil {
			fmt.Printf("Failed to run command: %v\n", err)
			os.Exit(1)
		}

		if data, err := ioutil.ReadFile(fmt.Sprintf("%s.xml", scnr.ipRange)); err == nil {
			if output, err := nmap.Parse(data); err == nil {
				report := report{}
				report.ipRange = scnr.ipRange
				report.openPorts = map[int][]*Host{}
				for _, scannedHost := range output.Hosts {
					newHost, ok := storedHosts[scannedHost.Addresses[0].Addr]
					if !ok { // new host
						newHost = &Host{}
						newHost.IpAddress = scannedHost.Addresses[0].Addr
						newHost.PortDetails = map[int]*Port{}
						storedHosts[scannedHost.Addresses[0].Addr] = newHost
					}

					for _, prt := range scannedHost.Ports {
						newPort, ok := newHost.PortDetails[prt.PortId]
						if !ok { // new port
							newPort = &Port{}
							newPort.Port = prt.PortId
							newPort.DownTime = 0 * time.Second
							newPort.State = "unknown"
							newPort.downPoint = time.Time{}
							newPort.ServiceName = prt.Service.Name
							newPort.ServiceVersion = prt.Service.Version
							newHost.PortDetails[newPort.Port] = newPort
						}

						if !newPort.downPoint.IsZero() {
							newPort.DownTime += time.Since(newPort.downPoint)
						}

						if prt.State.State != "open" || newPort.ServiceName != prt.Service.Name {
							newPort.downPoint = time.Now()
							newPort.State = FAILING
						} else { // port is now open, record downtime if there was any
							newPort.downPoint = time.Time{}
							newPort.State = SUCCEDING

							if report.openPorts[newPort.Port] == nil {
								report.openPorts[newPort.Port] = make([]*Host, 0)
							}
							report.openPorts[newPort.Port] = append(report.openPorts[newPort.Port], newHost)
						}

					}

				}

				reportUpdates <- report
				if outputData, err := json.MarshalIndent(storedHosts, "", "\t"); err == nil {
					ioutil.WriteFile(fmt.Sprintf("%s.json", scnr.ipRange), outputData, 0660)
				}
			} else {
				fmt.Printf("Failed to parse file: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Failed to read file: %v\n", err)
			os.Exit(1)
		}
	}
}
