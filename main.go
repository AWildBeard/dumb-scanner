package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/awesome-gocui/gocui"
	"gopkg.in/yaml.v3"
)

func main() {
	conf := &Config{}

	if configContent, err := ioutil.ReadFile("dumb-scanner.yaml"); err == nil {
		if err := yaml.Unmarshal(configContent, conf); err != nil || len(conf.MonitoredHosts) < 1 || len(conf.MonitoredPorts) < 1 || conf.NmapScanFlags == "" {
			fmt.Printf("Invalid Config!")
			os.Exit(0)
		}
	} else {
		fmt.Println("Config file missing. Generating a good default. Change the config to suit your needs. Run me again!")
		conf.MonitoredHosts = []string{"148.88.1.0/24", "148.88.2.0/24", "148.88.3.0/24", "148.88.4.0/24", "148.88.5.0/24",
			"148.88.6.0/24", "148.88.7.0/24", "148.88.8.0/24", "148.88.9.0/24"}
		conf.MonitoredPorts = []int{22, 25, 53, 80, 443}
		conf.NmapScanFlags = "-sC -sV"
		conf.HistoricBannerChecking = false
		configContent, _ = yaml.Marshal(conf)
		ioutil.WriteFile("dumb-scanner.yaml", configContent, 0660)
		os.Exit(0)
	}

	reportUpdates := make(chan report)

	for _, host := range conf.MonitoredHosts {
		const cmd = "nmap %s -n -p %s -oX %s.xml %s"
		ports := strings.Builder{}
		portLen := len(conf.MonitoredPorts)

		for index, port := range conf.MonitoredPorts {
			if index == portLen-1 {
				ports.WriteString(fmt.Sprintf("%d", port))
			} else {
				ports.WriteString(fmt.Sprintf("%d,", port))
			}
		}

		outputFileName := strings.ReplaceAll(host, "/", "_cidr")
		newScanner := &scanner{
			ipRange:   host,
			fileName:  outputFileName,
			nmapCmd:   fmt.Sprintf(cmd, conf.NmapScanFlags, ports.String(), outputFileName, host),
			smartMode: conf.HistoricBannerChecking,
		}

		go newScanner.scanEmit(reportUpdates)
	}

	gui, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		fmt.Printf("Failed to instantiate gui: %v\n", err)
		os.Exit(1)
	}
	defer gui.Close()

	gui.SetManagerFunc(conf.layout)
	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}); err != nil {
		return
	}

	go func() {
		for true {
			select {
			case rp := <-reportUpdates:
				gui.Update(func(g *gocui.Gui) error {
					view, err := g.View(rp.ipRange)
					if err == nil {
						view.Clear()
						for _, requiredPort := range conf.MonitoredPorts {
							found := false
							for port, _ := range rp.openPorts {
								if port == requiredPort {
									found = true
								}
							}

							if !found {
								fmt.Fprintf(view, "%6d : \033[38;5;196mFAIL\033[m\n", requiredPort)
							} else {
								fmt.Fprintf(view, "%6d : \033[38;5;118mSUCC\033[m\n", requiredPort)
							}
						}
					}

					return err
				})
			}
		}
	}()

	if err := gui.MainLoop(); err != nil && !gocui.IsQuit(err) {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
