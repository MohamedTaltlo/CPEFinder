// SPDX-License-Identifier: MIT
// Copyright (c) 2026 CPEFinder contributors

//go:build windows

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func (s *Scanner) ubiquitiRound(stop <-chan struct{}) {
	locals := s.adapter.IPv4
	if len(locals) == 0 {
		locals = []net.IP{net.IPv4zero}
	}
	for _, lip := range locals {
		select {
		case <-stop:
			return
		default:
		}
		addr := &net.UDPAddr{IP: lip, Port: 0}
		c, err := net.ListenUDP("udp4", addr)
		if err != nil {
			continue
		}
		_ = enableBroadcast(c)
		targets := []net.IP{net.IPv4bcast, net.IPv4(233, 89, 188, 1)}
		targets = append(targets, s.adapter.Broadcasts...)
		for _, t := range targets {
			for _, q := range [][]byte{{1, 0, 0, 0}, {2, 0, 0, 0}} {
				_, _ = c.WriteToUDP(q, &net.UDPAddr{IP: t, Port: 10001})
			}
		}
		_ = c.SetReadDeadline(time.Now().Add(1800 * time.Millisecond))
		buf := make([]byte, 8192)
		for {
			n, src, e := c.ReadFromUDP(buf)
			if e != nil {
				break
			}
			d := parseUBNT(buf[:n])
			if d != nil {
				if d.IP == "" {
					d.IP = src.IP.String()
				}
				d.Vendor = "Ubiquiti"
				d.Method = "Ubiquiti Discovery"
				d.LastSeen = time.Now()
				s.merge(*d)
			}
		}
		c.Close()
	}
}

func enableBroadcast(c *net.UDPConn) error {
	rc, e := c.SyscallConn()
	if e != nil {
		return e
	}
	var se error
	e = rc.Control(func(fd uintptr) {
		se = syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	})
	if e != nil {
		return e
	}
	return se
}

func parseUBNT(data []byte) *Device {
	if len(data) < 4 || (data[0] != 1 && data[0] != 2) {
		return nil
	}
	d := &Device{}
	ips := []string{}
	for o := 4; o+3 <= len(data); {
		typ := data[o]
		ln := int(binary.BigEndian.Uint16(data[o+1 : o+3]))
		o += 3
		if ln < 0 || o+ln > len(data) {
			break
		}
		v := data[o : o+ln]
		o += ln
		switch typ {
		case 0x01:
			if len(v) >= 6 {
				d.MAC = macString(v[:6])
			}
		case 0x02:
			for p := 0; p+10 <= len(v); p += 10 {
				m := macString(v[p : p+6])
				ip := net.IP(v[p+6 : p+10]).String()
				if d.MAC == "" {
					d.MAC = m
				}
				if validIPv4(ip) {
					ips = append(ips, ip)
				}
			}
		case 0x03:
			d.Firmware = cleanText(v)
		case 0x05:
			_ = v
		case 0x0B:
			d.Hostname = cleanText(v)
		case 0x0C:
			if d.Model == "" {
				d.Model = cleanText(v)
			}
		case 0x0D:
			d.SSID = cleanText(v)
		case 0x14, 0x15:
			if x := cleanText(v); x != "" {
				d.Model = x
			}
		case 0x1B:
			if d.Firmware == "" {
				d.Firmware = cleanText(v)
			}
		}
	}
	best := ""
	for _, ip := range ips {
		best = chooseIP(best, ip)
	}
	d.IP = best
	if d.MAC == "" && d.IP == "" {
		return nil
	}
	return d
}

func macString(b []byte) string {
	if len(b) < 6 {
		return ""
	}
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", b[0], b[1], b[2], b[3], b[4], b[5])
}

func cleanText(b []byte) string {
	return strings.TrimSpace(strings.Trim(string(bytes.Trim(b, "\x00")), "\x00"))
}

func (s *Scanner) genericUDP(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}
		s.ssdpRound(stop)
		select {
		case <-stop:
			return
		case <-time.After(8 * time.Second):
		}
	}
}

func (s *Scanner) ssdpRound(stop <-chan struct{}) {
	locals := s.adapter.IPv4
	if len(locals) == 0 {
		return
	}
	msg := "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nMAN: \"ssdp:discover\"\r\nMX: 1\r\nST: ssdp:all\r\n\r\n"
	for _, lip := range locals {
		c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: lip, Port: 0})
		if e != nil {
			continue
		}
		_, _ = c.WriteToUDP([]byte(msg), &net.UDPAddr{IP: net.IPv4(239, 255, 255, 250), Port: 1900})
		_ = c.SetReadDeadline(time.Now().Add(1200 * time.Millisecond))
		buf := make([]byte, 4096)
		for {
			n, src, er := c.ReadFromUDP(buf)
			if er != nil {
				break
			}
			txt := string(buf[:n])
			mac := lookupARP(src.IP.String())
			vendor := vendorFromMAC(mac)
			if vendor == "" {
				lower := strings.ToLower(txt)
				if strings.Contains(lower, "ubiquiti") {
					vendor = "Ubiquiti"
				}
				if strings.Contains(lower, "mimosa") {
					vendor = "Mimosa"
				}
				if strings.Contains(lower, "cambium") {
					vendor = "Cambium"
				}
			}
			if vendor != "" {
				s.merge(Device{IP: src.IP.String(), MAC: mac, Vendor: vendor, Description: firstHeader(txt, "SERVER"), Method: "SSDP", LastSeen: time.Now()})
			}
		}
		c.Close()
	}
}

func firstHeader(s, name string) string {
	for _, ln := range strings.Split(s, "\n") {
		p := strings.SplitN(strings.TrimSpace(ln), ":", 2)
		if len(p) == 2 && strings.EqualFold(strings.TrimSpace(p[0]), name) {
			return strings.TrimSpace(p[1])
		}
	}
	return ""
}

func (s *Scanner) arpLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(1800 * time.Millisecond)
	defer ticker.Stop()
	for {
		s.readARPCache()
		select {
		case <-stop:
			return
		case <-ticker.C:
		}
	}
}

type mibIPNetRow struct {
	Index       uint32
	PhysAddrLen uint32
	PhysAddr    [8]byte
	Addr        uint32
	Type        uint32
}

type arpEntry struct {
	Index int
	IP    string
	MAC   string
}

func getARPEntries() []arpEntry {
	var size uint32
	pGetIpNetTable.Call(0, uintptr(unsafe.Pointer(&size)), 0)
	if size < 4 {
		return nil
	}
	buf := make([]byte, size)
	r, _, _ := pGetIpNetTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0)
	if r != 0 || len(buf) < 4 {
		return nil
	}
	n := *(*uint32)(unsafe.Pointer(&buf[0]))
	rowSize := uintptr(unsafe.Sizeof(mibIPNetRow{}))
	base := uintptr(4)
	entries := make([]arpEntry, 0, int(n))
	for i := uint32(0); i < n; i++ {
		off := base + uintptr(i)*rowSize
		if off+rowSize > uintptr(len(buf)) {
			break
		}
		row := (*mibIPNetRow)(unsafe.Pointer(&buf[off]))
		if row.PhysAddrLen < 6 || row.PhysAddrLen > 8 {
			continue
		}
		ip := fmt.Sprintf("%d.%d.%d.%d", byte(row.Addr), byte(row.Addr>>8), byte(row.Addr>>16), byte(row.Addr>>24))
		if !validIPv4(ip) {
			continue
		}
		mac := macString(row.PhysAddr[:6])
		entries = append(entries, arpEntry{Index: int(row.Index), IP: ip, MAC: mac})
	}
	return entries
}

func (s *Scanner) readARPCache() {
	for _, e := range getARPEntries() {
		if e.Index != s.adapter.Index {
			continue
		}
		v := vendorFromMAC(e.MAC)
		if v == "" {
			continue
		}
		s.merge(Device{IP: e.IP, MAC: e.MAC, Vendor: v, Method: "ARP", LastSeen: time.Now()})
	}
}

func lookupARP(ip string) string {
	for _, e := range getARPEntries() {
		if e.IP == ip {
			return e.MAC
		}
	}
	return ""
}
