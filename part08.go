// SPDX-License-Identifier: MIT
// Copyright (c) 2026 CPEFinder contributors

//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

func (s *Scanner) mndpRound(stop <-chan struct{}) {
	locals := s.adapter.IPv4
	if len(locals) == 0 { locals = []net.IP{net.IPv4zero} }
	for _, lip := range locals {
		select { case <-stop: return; default: }
		c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: lip, Port: 0})
		if err != nil { continue }
		_ = enableBroadcast(c)
		targets := []net.IP{net.IPv4bcast}
		targets = append(targets, s.adapter.Broadcasts...)
		for _, t := range targets {
			_, _ = c.WriteToUDP([]byte{0, 0, 0, 0}, &net.UDPAddr{IP: t, Port: 5678})
		}
		_ = c.SetReadDeadline(time.Now().Add(1400 * time.Millisecond))
		buf := make([]byte, 4096)
		for {
			n, src, e := c.ReadFromUDP(buf)
			if e != nil { break }
			d := parseMNDP(buf[:n], src.IP.String())
			if d != nil { s.merge(*d) }
		}
		c.Close()
	}
}

func parseMNDP(data []byte, sourceIP string) *Device {
	if len(data) < 8 { return nil }
	d := &Device{}
	for o := 4; o+4 <= len(data); {
		typ := binary.BigEndian.Uint16(data[o : o+2])
		ln := int(binary.BigEndian.Uint16(data[o+2 : o+4]))
		o += 4
		if ln < 0 || o+ln > len(data) { break }
		v := data[o : o+ln]
		o += ln
		switch typ {
		case 1:
			if len(v) >= 6 { d.MAC = macString(v[:6]) }
		case 5:
			d.Hostname = cleanText(v)
		case 7:
			d.Firmware = cleanText(v)
		case 8:
			d.Description = cleanText(v)
		case 12:
			d.Model = cleanText(v)
		case 17:
			if len(v) == 4 { d.IP = fmt.Sprintf("%d.%d.%d.%d", v[0], v[1], v[2], v[3]) }
		}
	}
	d.MAC = normalizeMAC(d.MAC)
	vendor := vendorFromMAC(d.MAC)
	text := strings.ToLower(d.Description + " " + d.Model + " " + d.Hostname)
	if vendor == "" && (strings.Contains(text, "cambium") || strings.Contains(text, "epmp") || strings.Contains(text, "ptp550") || strings.Contains(text, "ptp 550") || strings.Contains(text, "pmp450") || strings.Contains(text, "pmp 450") || strings.Contains(text, "cnpilot")) { vendor = "Cambium" }
	if vendor != "Cambium" { return nil }
	if !validIPv4(d.IP) && validIPv4(sourceIP) { d.IP = sourceIP }
	d.Vendor = "Cambium"
	d.Method = "Cambium MNDP"
	d.LastSeen = time.Now()
	if d.MAC == "" && d.IP == "" { return nil }
	return d
}

func (s *Scanner) pokeRound(stop <-chan struct{}) {
	for _, lip := range s.adapter.IPv4 {
		v := lip.To4()
		if v == nil { continue }
		c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: lip, Port: 0})
		if err != nil { continue }
		payload := []byte{0}
		for host := 1; host < 255; host++ {
			select { case <-stop: c.Close(); return; default: }
			if byte(host) == v[3] { continue }
			t := net.IPv4(v[0], v[1], v[2], byte(host))
			_, _ = c.WriteToUDP(payload, &net.UDPAddr{IP: t, Port: 9})
		}
		for _, ip := range []string{"192.168.1.20", "192.168.1.10", "169.254.1.1", "192.168.0.20", "192.168.0.1"} {
			if p := net.ParseIP(ip); p != nil { _, _ = c.WriteToUDP(payload, &net.UDPAddr{IP: p, Port: 9}) }
		}
		c.Close()
	}
}
