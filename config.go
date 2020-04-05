package main

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type Config struct {
	MonitoredHosts         []string
	MonitoredPorts         []int
	NmapScanFlags          string
	HistoricBannerChecking bool
}

func (conf *Config) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	sidePanelWidth := 25
	if sidePanelWidth > maxX/5 {
		sidePanelWidth = maxX / 5
	}

	if view, err := g.SetView("side", -1, -1, sidePanelWidth, maxY, 0); err != nil {
		view.Wrap = true
		fmt.Fprintln(view, "\033[1m\033[4mMonitored Hosts:\033[m")
		for _, host := range conf.MonitoredHosts {
			fmt.Fprintf(view, "\t%s\n", host)
		}

		fmt.Fprintln(view, "\n\033[1m\033[4mMonitored Ports:\033[m")
		for _, port := range conf.MonitoredPorts {
			fmt.Fprintf(view, "\t%d\n", port)
		}
	}

	baseWidth := 17
	for _, host := range conf.MonitoredHosts {
		hostLen := len(host) + 3
		if hostLen > baseWidth {
			baseWidth = hostLen
		}
	}

	yoffset := 0
	height := 1 + len(conf.MonitoredPorts)
	topLeftX := sidePanelWidth + 1
	width := sidePanelWidth + 1 + baseWidth

	for _, host := range conf.MonitoredHosts {
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
