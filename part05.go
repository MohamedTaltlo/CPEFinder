// SPDX-License-Identifier: MIT
// Copyright (c) 2026 CPEFinder contributors

//go:build windows

package main

import (
	"sync"
	"syscall"
	"unsafe"
)

func makeAppIcon(hinst uintptr, size int) uintptr {
	if size < 16 {
		size = 16
	}
	strideMask := ((size + 31) / 32) * 4
	andMask := make([]byte, strideMask*size)
	for i := range andMask {
		andMask[i] = 0xFF
	}
	xor := make([]byte, size*size*4)
	set := func(x, y int, r, g, b byte) {
		if x < 0 || y < 0 || x >= size || y >= size {
			return
		}
		dy := size - 1 - y
		o := (dy*size + x) * 4
		xor[o+0], xor[o+1], xor[o+2], xor[o+3] = b, g, r, 0xFF
		mo := dy*strideMask + x/8
		andMask[mo] &^= 1 << uint(7-(x%8))
	}
	pad := maxInt(1, size/16)
	rad := maxInt(3, size/5)
	for y := pad; y < size-pad; y++ {
		for x := pad; x < size-pad; x++ {
			dx := 0
			dy := 0
			if x < pad+rad {
				dx = pad + rad - x
			} else if x >= size-pad-rad {
				dx = x - (size - pad - rad - 1)
			}
			if y < pad+rad {
				dy = pad + rad - y
			} else if y >= size-pad-rad {
				dy = y - (size - pad - rad - 1)
			}
			if dx*dx+dy*dy <= rad*rad || dx == 0 || dy == 0 {
				set(x, y, 37, 99, 235)
			}
		}
	}
	cx, cy := size/2, size/2
	th := maxInt(1, size/16)
	drawLine := func(x0, y0, x1, y1 int) {
		dx := absInt(x1 - x0)
		sx := -1
		if x0 < x1 {
			sx = 1
		}
		dy := -absInt(y1 - y0)
		sy := -1
		if y0 < y1 {
			sy = 1
		}
		err := dx + dy
		for {
			for yy := -th; yy <= th; yy++ {
				for xx := -th; xx <= th; xx++ {
					set(x0+xx, y0+yy, 255, 255, 255)
				}
			}
			if x0 == x1 && y0 == y1 {
				break
			}
			e2 := 2 * err
			if e2 >= dy {
				err += dy
				x0 += sx
			}
			if e2 <= dx {
				err += dx
				y0 += sy
			}
		}
	}
	nodes := [][2]int{{cx, cy}, {size / 3, size / 3}, {2 * size / 3, size / 3}, {cx, 2 * size / 3}}
	for i := 1; i < len(nodes); i++ {
		drawLine(cx, cy, nodes[i][0], nodes[i][1])
	}
	rn := maxInt(2, size/10)
	for _, n := range nodes {
		for y := -rn; y <= rn; y++ {
			for x := -rn; x <= rn; x++ {
				if x*x+y*y <= rn*rn {
					set(n[0]+x, n[1]+y, 255, 255, 255)
				}
			}
		}
	}
	if len(andMask) == 0 || len(xor) == 0 {
		return 0
	}
	h, _, _ := pCreateIcon.Call(hinst, uintptr(size), uintptr(size), 1, 32, uintptr(unsafe.Pointer(&andMask[0])), uintptr(unsafe.Pointer(&xor[0])))
	return h
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type App struct {
	hwnd                                                     uintptr
	combo, list                                              uintptr
	btnScan, btnStop, btnRefresh, btnClear, btnCopy, btnOpen uintptr
	font, fontBold, fontTitle                                uintptr
	adapters                                                 []Adapter
	scanner                                                  *Scanner
	events                                                   chan Device
	rows                                                     map[string]int
	rowKeys                                                  []string
	devices                                                  map[string]Device
	scanning                                                 bool
	autoStarted                                              bool
	count                                                    int
	pendingMu                                                sync.Mutex
	pendingAdapters                                          []Adapter
	closing                                                  bool
	iconBig, iconSmall                                       uintptr
}

var app *App

func main() {
	ic := INITCOMMONCONTROLSEX{DwSize: uint32(unsafe.Sizeof(INITCOMMONCONTROLSEX{})), DwICC: 0x00004000 | 0x00000001}
	pInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&ic)))
	app = &App{events: make(chan Device, 512), rows: map[string]int{}, devices: map[string]Device{}}
	app.scanner = NewScanner(app.events)
	app.run()
}

func (a *App) run() {
	hinst, _, _ := pGetModuleHandleW.Call(0)
	cls := u16("CPEFinderV9Window")
	cur, _, _ := pLoadCursorW.Call(0, IDC_ARROW)
	a.iconBig = makeAppIcon(hinst, 32)
	a.iconSmall = makeAppIcon(hinst, 16)
	wc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), LpfnWndProc: syscall.NewCallback(wndProc), HInstance: hinst, HIcon: a.iconBig, HIconSm: a.iconSmall, HCursor: cur, LpszClassName: cls}
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	hwnd, _, _ := pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(u16(appTitle+" - V"+appVersion))), WS_OVERLAPPEDWINDOW|WS_VISIBLE|WS_CLIPCHILDREN, 100, 60, 1180, 760, 0, 0, hinst, 0)
	a.hwnd = hwnd
	if a.iconBig != 0 {
		pSendMessageW.Call(hwnd, WM_SETICON, ICON_BIG, a.iconBig)
	}
	if a.iconSmall != 0 {
		pSendMessageW.Call(hwnd, WM_SETICON, ICON_SMALL, a.iconSmall)
	}
	pShowWindow.Call(hwnd, SW_SHOW)
	pUpdateWindow.Call(hwnd)
	pSetTimer.Call(hwnd, 1, 250, 0)
	var msg MSG
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func wndProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		app.hwnd = hwnd
		app.createUI()
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_PAINT:
		app.paint()
		return 0
	case WM_SIZE:
		app.layout()
		return 0
	case WM_TIMER:
		app.drainEvents()
		return 0
	case WM_COMMAND:
		id := int(loWord(wparam))
		code := hiWord(wparam)
		switch id {
		case ID_SCAN:
			app.startScan()
		case ID_STOP:
			app.stopScan()
		case ID_REFRESH:
			app.refreshAdaptersAsync()
		case ID_CLEAR:
			app.clearRows()
		case ID_COPY:
			app.copySelected()
		case ID_OPEN:
			app.openSelected()
		case ID_ADAPTER:
			_ = code
		}
		return 0
	case WM_APP_ADAPTERS:
		app.applyPendingAdapters()
		return 0
	case WM_APP_EXIT:
		pDestroyWindow.Call(hwnd)
		return 0
	case WM_NOTIFY:
		hdr := (*NMHDR)(unsafe.Pointer(lparam))
		if hdr != nil && hdr.HwndFrom == app.list && hdr.Code == NM_DBLCLK {
			app.openSelected()
		}
		return 0
	case WM_CLOSE:
		if app.closing {
			return 0
		}
		app.closing = true
		pKillTimer.Call(hwnd, 1)
		pShowWindow.Call(hwnd, 0)
		go func() {
			app.scanner.Stop()
			pPostMessageW.Call(hwnd, WM_APP_EXIT, 0, 0)
		}()
		return 0
	case WM_DESTROY:
		pPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, uintptr(msg), wparam, lparam)
	return r
}
