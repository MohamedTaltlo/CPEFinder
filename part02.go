// SPDX-License-Identifier: MIT
// Copyright (c) 2026 CPEFinder contributors

//go:build windows

package main

import (
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	iphlpapi = syscall.NewLazyDLL("iphlpapi.dll")

	pRegisterClassExW     = user32.NewProc("RegisterClassExW")
	pCreateWindowExW      = user32.NewProc("CreateWindowExW")
	pDefWindowProcW       = user32.NewProc("DefWindowProcW")
	pShowWindow           = user32.NewProc("ShowWindow")
	pUpdateWindow         = user32.NewProc("UpdateWindow")
	pGetMessageW          = user32.NewProc("GetMessageW")
	pTranslateMessage     = user32.NewProc("TranslateMessage")
	pDispatchMessageW     = user32.NewProc("DispatchMessageW")
	pPostQuitMessage      = user32.NewProc("PostQuitMessage")
	pDestroyWindow        = user32.NewProc("DestroyWindow")
	pSendMessageW         = user32.NewProc("SendMessageW")
	pPostMessageW         = user32.NewProc("PostMessageW")
	pBeginPaint           = user32.NewProc("BeginPaint")
	pEndPaint             = user32.NewProc("EndPaint")
	pFillRect             = user32.NewProc("FillRect")
	pInvalidateRect       = user32.NewProc("InvalidateRect")
	pGetClientRect        = user32.NewProc("GetClientRect")
	pMoveWindow           = user32.NewProc("MoveWindow")
	pSetTimer             = user32.NewProc("SetTimer")
	pKillTimer            = user32.NewProc("KillTimer")
	pLoadCursorW          = user32.NewProc("LoadCursorW")
	pCreateIcon           = user32.NewProc("CreateIcon")
	pSetWindowTextW       = user32.NewProc("SetWindowTextW")
	pGetWindowTextW       = user32.NewProc("GetWindowTextW")
	pGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	pEnableWindow         = user32.NewProc("EnableWindow")
	pSetWindowLongPtrW    = user32.NewProc("SetWindowLongPtrW")
	pGetWindowLongPtrW    = user32.NewProc("GetWindowLongPtrW")
	pOpenClipboard        = user32.NewProc("OpenClipboard")
	pEmptyClipboard       = user32.NewProc("EmptyClipboard")
	pSetClipboardData     = user32.NewProc("SetClipboardData")
	pCloseClipboard       = user32.NewProc("CloseClipboard")
	pDrawTextW            = user32.NewProc("DrawTextW")
	pSetTextColor         = gdi32.NewProc("SetTextColor")
	pSetBkMode            = gdi32.NewProc("SetBkMode")
	pCreateSolidBrush     = gdi32.NewProc("CreateSolidBrush")
	pDeleteObject         = gdi32.NewProc("DeleteObject")
	pRoundRect            = gdi32.NewProc("RoundRect")
	pCreatePen            = gdi32.NewProc("CreatePen")
	pSelectObject         = gdi32.NewProc("SelectObject")
	pCreateFontW          = gdi32.NewProc("CreateFontW")
	pGetStockObject       = gdi32.NewProc("GetStockObject")
	pSetDCBrushColor      = gdi32.NewProc("SetDCBrushColor")
	pSetDCPenColor        = gdi32.NewProc("SetDCPenColor")
	pGlobalAlloc          = kernel32.NewProc("GlobalAlloc")
	pGlobalLock           = kernel32.NewProc("GlobalLock")
	pGlobalUnlock         = kernel32.NewProc("GlobalUnlock")
	pGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	pShellExecuteW        = shell32.NewProc("ShellExecuteW")
	pInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	pGetIpNetTable        = iphlpapi.NewProc("GetIpNetTable")
)

func rgb(r, g, b byte) uintptr { return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16) }

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func loWord(v uintptr) uint16 { return uint16(v & 0xFFFF) }

func hiWord(v uintptr) uint16 { return uint16((v >> 16) & 0xFFFF) }

type Adapter struct {
	Name       string
	Index      int
	MAC        string
	IPv4       []net.IP
	Broadcasts []net.IP
	Up         bool
}

type Device struct {
	Key         string
	IP          string
	MAC         string
	Vendor      string
	Model       string
	Hostname    string
	Firmware    string
	SSID        string
	Method      string
	Description string
	LastSeen    time.Time
}

type Scanner struct {
	mu       sync.Mutex
	running  atomic.Bool
	stop     chan struct{}
	adapter  Adapter
	events   chan Device
	devices  map[string]Device
	lastUI   map[string]time.Time
	ubntSeen map[string]time.Time
}

func NewScanner(events chan Device) *Scanner {
	return &Scanner{events: events, devices: make(map[string]Device), lastUI: make(map[string]time.Time), ubntSeen: make(map[string]time.Time)}
}

func (s *Scanner) IsRunning() bool { return s.running.Load() }

func (s *Scanner) Start(a Adapter) {
	if !s.running.CompareAndSwap(false, true) {
		return
	}

	s.mu.Lock()
	s.stop = make(chan struct{})
	s.adapter = a
	stop := s.stop
	s.mu.Unlock()

	go func() {
		defer s.running.Store(false)
		var wg sync.WaitGroup
		wg.Add(3)
		go func() { defer wg.Done(); s.ubiquitiLoop(stop) }()
		go func() { defer wg.Done(); s.arpLoop(stop) }()
		go func() { defer wg.Done(); s.genericUDP(stop) }()
		wg.Wait()
	}()
}

func (s *Scanner) Stop() {
	s.mu.Lock()
	if s.stop != nil {
		select {
		case <-s.stop:
		default:
			close(s.stop)
		}
	}
	s.mu.Unlock()
}

func (s *Scanner) ClearCache() {
	s.mu.Lock()
	s.devices = make(map[string]Device)
	s.lastUI = make(map[string]time.Time)
	s.ubntSeen = make(map[string]time.Time)
	s.mu.Unlock()
}

func normalizeMAC(m string) string {
	m = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(m, "-", ":"), ".", ""))
	if !strings.Contains(m, ":") && len(m) == 12 {
		var p []string
		for i := 0; i < 12; i += 2 {
			p = append(p, m[i:i+2])
		}
		m = strings.Join(p, ":")
	}
	return m
}

func validIPv4(ip string) bool {
	x := net.ParseIP(ip)
	return x != nil && x.To4() != nil && ip != "0.0.0.0" && ip != "255.255.255.255"
}

func ipScore(ip string) int {
	x := net.ParseIP(ip)
	if x == nil || x.To4() == nil {
		return -100
	}
	b := x.To4()
	if b[0] == 0 || b[0] >= 224 {
		return -50
	}
	if b[0] == 169 && b[1] == 254 {
		return 10
	}
	return 100
}

func chooseIP(oldIP, newIP string) string {
	if !validIPv4(newIP) {
		return oldIP
	}
	if !validIPv4(oldIP) || ipScore(newIP) > ipScore(oldIP) {
		return newIP
	}
	if ipScore(newIP) == ipScore(oldIP) && oldIP == "" {
		return newIP
	}
	return oldIP
}

func vendorFromMAC(mac string) string {
	m := strings.ToUpper(normalizeMAC(mac))
	if len(m) < 8 {
		return ""
	}
	p := m[:8]
	if _, ok := ubntOUIs[p]; ok {
		return "Ubiquiti"
	}
	if _, ok := mimosaOUIs[p]; ok {
		return "Mimosa"
	}
	if _, ok := cambiumOUIs[p]; ok {
		return "Cambium"
	}
	return ""
}
