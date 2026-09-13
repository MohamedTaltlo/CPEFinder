// SPDX-License-Identifier: MIT
// Copyright (c) 2026 CPEFinder contributors

//go:build windows

package main

import (
	"net"
	"sort"
	"strings"
	"time"
)

var ubntOUIs = makeSet([]string{
	"00:15:6D", "00:27:22", "04:18:D6", "0C:EA:14", "18:E8:29", "1C:0B:8B", "1C:6A:1B", "24:5A:4C", "24:A4:3C", "28:70:4E", "2C:E5:BD", "44:D9:E7", "58:D6:1F", "60:22:32", "68:2E:3C", "68:72:51", "68:D7:9A", "6C:63:F8", "70:A7:41", "74:83:C2", "74:AC:B9", "74:F9:2C", "74:FA:29", "78:45:58", "78:8A:20", "80:2A:A8", "84:78:48", "8C:30:66", "8C:ED:E1", "90:41:B2", "94:2A:6F", "9C:05:D6", "A4:F8:FF", "A8:9C:6C", "AC:8B:A9", "B4:FB:E4", "CC:35:D9", "D0:21:F9", "D4:89:C1", "D8:B3:70", "D8:C2:62", "DC:9F:DB", "E0:63:DA", "E4:38:83", "F0:9F:C2", "F4:92:BF", "F4:E2:C6", "FC:EC:DA",
})

var mimosaOUIs = makeSet([]string{
	"20:B5:C6", "84:9C:A4", "8C:B6:C5", "90:70:BF", "CC:54:FE", "D0:B0:DF",
	"00:A0:0A", "00:01:AA",
})

var cambiumOUIs = makeSet([]string{
	"00:04:56", "30:CB:C7", "58:C1:7A", "90:14:AF", "90:6D:62", "B4:A2:5C", "BC:A9:93", "BC:E6:7C", "FC:11:65", "0A:00:3E",
})

func makeSet(a []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range a {
		m[v] = struct{}{}
	}
	return m
}

func methodScore(m string) int {
	switch m {
	case "Ubiquiti Discovery":
		return 100
	case "Cambium MNDP":
		return 95
	case "LLDP", "CDP":
		return 90
	case "DHCP":
		return 80
	case "SSDP", "WS-Discovery", "mDNS":
		return 70
	case "HTTP":
		return 65
	case "ARP مباشر":
		return 60
	case "ARP Probe":
		return 55
	case "ARP":
		return 50
	case "IPv4":
		return 40
	default:
		if strings.HasPrefix(m, "HTTP") {
			return 65
		}
		return 0
	}
}

func (s *Scanner) merge(d Device) {
	if d.MAC != "" {
		d.MAC = normalizeMAC(d.MAC)
	}
	if d.Vendor == "" && d.MAC != "" {
		d.Vendor = vendorFromMAC(d.MAC)
	}
	if d.Vendor == "" {
		t := strings.ToLower(d.Model + " " + d.Hostname + " " + d.Firmware + " " + d.Description)
		if strings.Contains(t, "ubiquiti") || strings.Contains(t, "airmax") || strings.Contains(t, "uisp") || strings.Contains(t, "wave") {
			d.Vendor = "Ubiquiti"
		}
		if strings.Contains(t, "mimosa") {
			d.Vendor = "Mimosa"
		}
		if strings.Contains(t, "cambium") || strings.Contains(t, "epmp") || strings.Contains(t, "ptp550") || strings.Contains(t, "ptp 550") || strings.Contains(t, "pmp450") || strings.Contains(t, "pmp 450") || strings.Contains(t, "cnpilot") {
			d.Vendor = "Cambium"
		}
	}
	if d.MAC != "" {
		d.Key = "mac:" + d.MAC
	} else if d.IP != "" {
		d.Key = "ip:" + d.IP
	} else {
		return
	}
	if d.LastSeen.IsZero() {
		d.LastSeen = time.Now()
	}

	s.mu.Lock()
	old, ok := s.devices[d.Key]
	changed := !ok
	if d.Method == "Ubiquiti Discovery" {
		last := s.ubntSeen[d.Key]
		if ok && !last.IsZero() && d.LastSeen.Sub(last) > 7*time.Second {
			old.Firmware = ""
			old.SSID = ""
			old.Model = ""
			old.Hostname = ""
			old.Description = ""
			changed = true
		}
		s.ubntSeen[d.Key] = d.LastSeen
	}
	if ok {
		oldIP := old.IP
		old.IP = chooseIP(old.IP, d.IP)
		if old.IP != oldIP { changed = true }
		if d.MAC != "" && d.MAC != old.MAC { old.MAC = d.MAC; changed = true }
		if d.Vendor != "" && d.Vendor != old.Vendor { old.Vendor = d.Vendor; changed = true }
		if d.Model != "" && d.Model != old.Model { old.Model = d.Model; changed = true }
		if d.Hostname != "" && d.Hostname != old.Hostname { old.Hostname = d.Hostname; changed = true }
		if d.Firmware != "" && d.Firmware != old.Firmware { old.Firmware = d.Firmware; changed = true }
		if d.SSID != "" && d.SSID != old.SSID { old.SSID = d.SSID; changed = true }
		if d.Description != "" && d.Description != old.Description { old.Description = d.Description; changed = true }
		if d.Method != "" && d.Method != old.Method && (old.Method == "" || methodScore(d.Method) >= methodScore(old.Method)) { old.Method = d.Method; changed = true }
		old.LastSeen = d.LastSeen
		d = old
	}
	s.devices[d.Key] = d
	last := s.lastUI[d.Key]
	emit := changed || time.Since(last) > 30*time.Second
	if emit { s.lastUI[d.Key] = time.Now() }
	s.mu.Unlock()
	if emit {
		select { case s.events <- d: default: }
	}
}

func enumerateAdapters() []Adapter {
	ifs, _ := net.Interfaces()
	out := []Adapter{}
	for _, i := range ifs {
		if i.Flags&net.FlagLoopback != 0 || len(i.HardwareAddr) == 0 { continue }
		a := Adapter{Name: i.Name, Index: i.Index, MAC: normalizeMAC(i.HardwareAddr.String()), Up: i.Flags&net.FlagUp != 0}
		addrs, _ := i.Addrs()
		for _, ad := range addrs {
			ipnet, ok := ad.(*net.IPNet)
			if !ok { continue }
			ip := ipnet.IP.To4()
			if ip == nil { continue }
			a.IPv4 = append(a.IPv4, append(net.IP(nil), ip...))
			mask := ipnet.Mask
			if len(mask) >= 4 {
				b := net.IPv4(ip[0]|^mask[0], ip[1]|^mask[1], ip[2]|^mask[2], ip[3]|^mask[3])
				a.Broadcasts = append(a.Broadcasts, b)
			}
		}
		out = append(out, a)
	}
	adapterScore := func(a Adapter) int {
		n := strings.ToLower(a.Name); sc := 0
		if a.Up { sc += 100 }
		if strings.Contains(n, "ethernet") || strings.Contains(n, "lan") { sc += 40 }
		if strings.Contains(n, "wi-fi") || strings.Contains(n, "wifi") || strings.Contains(n, "wlan") { sc += 10 }
		if strings.Contains(n, "virtual") || strings.Contains(n, "vmware") || strings.Contains(n, "hyper-v") || strings.Contains(n, "tailscale") || strings.Contains(n, "wireguard") || strings.Contains(n, "bluetooth") || strings.Contains(n, "loopback") { sc -= 80 }
		return sc
	}
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := adapterScore(out[i]), adapterScore(out[j])
		if si != sj { return si > sj }
		return out[i].Name < out[j].Name
	})
	return out
}

func (s *Scanner) ubiquitiLoop(stop <-chan struct{}) {
	for {
		select { case <-stop: return; default: }
		s.ubiquitiRound(stop)
		for i := 0; i < 20; i++ {
			select { case <-stop: return; case <-time.After(250 * time.Millisecond): }
		}
	}
}
