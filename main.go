package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/awesome-gocui/gocui"
)

const (
	cmd = "nmap -sV -n -p %s -oX %s.xml %s"
)

const (
	SUCCEDING = "open"
	FAILING   = "closed"
)

var (
	requiredPorts = []int{80, 443, 53, 25, 8080, 9090, 9001}
	hosts         = []string{"10.1.0.1", "10.1.0.2", "10.1.0.3", "10.1.0.4", "10.1.0.5", "10.1.0.6", "10.1.0.7", "10.1.0.8", "10.1.0.9"}
	//requiredPorts = []int{80, 443, 53, 25}
	//hosts         = []string{"148.88.1.0/24", "148.88.2.0/24", "148.88.3.0/24", "148.88.4.0/24", "148.88.5.0/24", "148.88.6.0/24", "148.88.7.0/24", "148.88.8.0/24", "148.88.9.0/24"}
)

func main() {
	reportUpdates := make(chan report)

	for _, host := range hosts {
		ports := strings.Builder{}
		portLen := len(requiredPorts)
		for index, port := range requiredPorts {
			if index == portLen-1 {
				ports.WriteString(fmt.Sprintf("%d", port))
			} else {
				ports.WriteString(fmt.Sprintf("%d,", port))
			}
		}
		newScanner := &scanner{ipRange: host, nmapCmd: fmt.Sprintf(cmd, ports.String(), host, host)}
		go newScanner.scanEmit(reportUpdates)
	}

	gui, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		fmt.Printf("Failed to instantiate gui: %v\n", err)
		os.Exit(1)
	}
	defer gui.Close()

	gui.SetManagerFunc(layout)
	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}); err != nil {
		os.Exit(0)
	}

	go func() {
		for true {
			select {
			case rp := <-reportUpdates:
				gui.Update(func(g *gocui.Gui) error {
					view, err := g.View(rp.ipRange)
					if err == nil {
						view.Clear()
						for _, requiredPort := range requiredPorts {
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

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	sidePanelWidth := 25
	if sidePanelWidth > maxX/5 {
		sidePanelWidth = maxX / 5
	}

	if view, err := g.SetView("side", -1, -1, sidePanelWidth, maxY, 0); err != nil {
		view.Wrap = true
		fmt.Fprintln(view, "Monitored Hosts:")
		for _, host := range hosts {
			fmt.Fprintf(view, "\t%s\n", host)
		}

		fmt.Fprintln(view, "\nMonitored Ports:")
		for _, port := range requiredPorts {
			fmt.Fprintf(view, "\t%d\n", port)
		}
	}

	baseWidth := 17
	for _, host := range hosts {
		hostLen := len(host) + 3
		if hostLen > baseWidth {
			baseWidth = hostLen
		}
	}

	yoffset := 0
	height := 1 + len(requiredPorts)
	topLeftX := sidePanelWidth + 1
	width := sidePanelWidth + 1 + baseWidth

	for _, host := range hosts {
		if height+yoffset >= maxY {
			topLeftX = width + 1
			width += baseWidth
			yoffset = 0
		}

		if view, err := g.SetView(host, topLeftX, yoffset, width, height+yoffset, 0); err != nil {
			view.Title = fmt.Sprintf("%s\n", host)
		}
		yoffset += height + 1
	}
	return nil
}
