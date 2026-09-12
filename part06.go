// SPDX-License-Identifier: MIT
// Copyright (c) 2026 CPEFinder contributors

//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

func (a *App) createUI() {
	a.font = createFont(10, 400)
	a.fontBold = createFont(10, 600)
	a.fontTitle = createFont(22, 700)
	a.combo = createCtrl(WS_EX_LAYOUTRTL|WS_EX_RTLREADING, "COMBOBOX", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST|CBS_HASSTRINGS, 0, 0, 100, 200, a.hwnd, ID_ADAPTER)
	a.btnScan = a.button("بدء الفحص", ID_SCAN)
	a.btnStop = a.button("إيقاف", ID_STOP)
	a.btnRefresh = a.button("تحديث الكروت", ID_REFRESH)
	a.btnClear = a.button("مسح", ID_CLEAR)
	a.btnCopy = a.button("نسخ IP", ID_COPY)
	a.btnOpen = a.button("فتح الجهاز", ID_OPEN)
	a.list = createCtrl(WS_EX_CLIENTEDGE|WS_EX_LAYOUTRTL|WS_EX_RTLREADING, "SysListView32", "", WS_CHILD|WS_VISIBLE|WS_TABSTOP|LVS_REPORT|LVS_SINGLESEL|LVS_SHOWSELALWAYS|LVS_NOSORTHEADER, 0, 0, 100, 100, a.hwnd, 2000)
	pSendMessageW.Call(a.list, LVM_SETEXTENDEDLISTVIEWSTYLE, LVS_EX_FULLROWSELECT|LVS_EX_GRIDLINES|LVS_EX_DOUBLEBUFFER, LVS_EX_FULLROWSELECT|LVS_EX_GRIDLINES|LVS_EX_DOUBLEBUFFER)
	cols := []struct {
		name  string
		width int32
	}{{"العنوان IP", 140}, {"MAC", 150}, {"الشركة", 95}, {"الطراز", 155}, {"اسم الجهاز", 145}, {"SSID", 165}, {"الإصدار", 180}, {"الاكتشاف", 125}, {"آخر ظهور", 85}}
	for i, c := range cols {
		txt := u16(c.name)
		lc := LVCOLUMN{Mask: LVCF_FMT | LVCF_WIDTH | LVCF_TEXT | LVCF_SUBITEM, Fmt: LVCFMT_RIGHT, Cx: c.width, PszText: txt, ISubItem: int32(i)}
		pSendMessageW.Call(a.list, LVM_INSERTCOLUMNW, uintptr(i), uintptr(unsafe.Pointer(&lc)))
	}
	for _, h := range []uintptr{a.combo, a.btnScan, a.btnStop, a.btnRefresh, a.btnClear, a.btnCopy, a.btnOpen, a.list} {
		pSendMessageW.Call(h, WM_SETFONT, a.font, 1)
	}
	a.refreshAdaptersSync()
	pEnableWindow.Call(a.btnStop, 0)
	a.layout()
}

func createFont(size int, weight int) uintptr {
	r, _, _ := pCreateFontW.Call(uintptr(uint32(int32(-size*96/72))), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16("Segoe UI"))))
	return r
}

func createCtrl(ex uintptr, cls, text string, style uintptr, x, y, w, h int, parent uintptr, id int) uintptr {
	r, _, _ := pCreateWindowExW.Call(ex, uintptr(unsafe.Pointer(u16(cls))), uintptr(unsafe.Pointer(u16(text))), style, uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, uintptr(id), 0, 0)
	return r
}

func (a *App) button(text string, id int) uintptr {
	return createCtrl(0, "BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP, 0, 0, 100, 34, a.hwnd, id)
}

func (a *App) setAdapters(list []Adapter) {
	a.adapters = list
	pSendMessageW.Call(a.combo, CB_RESETCONTENT, 0, 0)
	for _, ad := range a.adapters {
		state := ""
		if !ad.Up {
			state = " — غير متصل"
		}
		txt := fmt.Sprintf("%s   %s%s", ad.Name, ad.MAC, state)
		pSendMessageW.Call(a.combo, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(txt))))
	}
	if len(a.adapters) > 0 {
		pSendMessageW.Call(a.combo, CB_SETCURSEL, 0, 0)
	}
}

func (a *App) refreshAdaptersSync() { a.setAdapters(enumerateAdapters()) }

func (a *App) refreshAdaptersAsync() {
	pEnableWindow.Call(a.btnRefresh, 0)
	go func() {
		list := enumerateAdapters()
		a.pendingMu.Lock()
		a.pendingAdapters = list
		a.pendingMu.Unlock()
		pPostMessageW.Call(a.hwnd, WM_APP_ADAPTERS, 0, 0)
	}()
}

func (a *App) applyPendingAdapters() {
	a.pendingMu.Lock()
	list := a.pendingAdapters
	a.pendingAdapters = nil
	a.pendingMu.Unlock()
	a.setAdapters(list)
	pEnableWindow.Call(a.btnRefresh, 1)
}

func (a *App) selectedAdapter() (Adapter, bool) {
	r, _, _ := pSendMessageW.Call(a.combo, CB_GETCURSEL, 0, 0)
	i := int(int32(r))
	if i < 0 || i >= len(a.adapters) {
		return Adapter{}, false
	}
	return a.adapters[i], true
}

func (a *App) startScan() {
	ad, ok := a.selectedAdapter()
	if !ok {
		return
	}
	if a.scanner.IsRunning() {
		return
	}
	a.scanning = true
	a.scanner.Start(ad)
	pEnableWindow.Call(a.btnScan, 0)
	pEnableWindow.Call(a.btnStop, 1)
	a.invalidate()
}

func (a *App) stopScan() {
	if !a.scanning && !a.scanner.IsRunning() {
		return
	}
	a.scanning = false
	pEnableWindow.Call(a.btnScan, 1)
	pEnableWindow.Call(a.btnStop, 0)
	a.invalidate()
	go a.scanner.Stop()
}

func (a *App) clearRows() {
	go a.scanner.ClearCache()
	a.rows = map[string]int{}
	a.rowKeys = nil
	a.devices = map[string]Device{}
	a.count = 0
	pSendMessageW.Call(a.list, LVM_DELETEALLITEMS, 0, 0)
	a.invalidate()
}

func (a *App) drainEvents() {
	if a.scanning && !a.scanner.IsRunning() {
		a.scanning = false
		pEnableWindow.Call(a.btnScan, 1)
		pEnableWindow.Call(a.btnStop, 0)
		a.invalidate()
	}
	for i := 0; i < 64; i++ {
		select {
		case d := <-a.events:
			a.upsertRow(d)
		default:
			return
		}
	}
}

func (a *App) upsertRow(d Device) {
	a.devices[d.Key] = d
	idx, ok := a.rows[d.Key]
	vals := []string{d.IP, d.MAC, d.Vendor, d.Model, d.Hostname, d.SSID, d.Firmware, d.Method, d.LastSeen.Format("15:04:05")}
	if !ok {
		idx = len(a.rowKeys)
		a.rows[d.Key] = idx
		a.rowKeys = append(a.rowKeys, d.Key)
		it := LVITEM{Mask: LVIF_TEXT, IItem: int32(idx), ISubItem: 0, PszText: u16(vals[0])}
		pSendMessageW.Call(a.list, LVM_INSERTITEMW, 0, uintptr(unsafe.Pointer(&it)))
		a.count++
		a.invalidate()
	}
	for c := 0; c < len(vals); c++ {
		it := LVITEM{Mask: LVIF_TEXT, IItem: int32(idx), ISubItem: int32(c), PszText: u16(vals[c])}
		pSendMessageW.Call(a.list, LVM_SETITEMW, 0, uintptr(unsafe.Pointer(&it)))
	}
}

func (a *App) selectedDevice() (Device, bool) {
	r, _, _ := pSendMessageW.Call(a.list, LVM_GETNEXTITEM, ^uintptr(0), LVNI_SELECTED)
	i := int(int32(r))
	if i < 0 || i >= len(a.rowKeys) {
		return Device{}, false
	}
	d, ok := a.devices[a.rowKeys[i]]
	return d, ok
}

func (a *App) copySelected() {
	if d, ok := a.selectedDevice(); ok && d.IP != "" {
		setClipboard(d.IP)
	}
}

func (a *App) openSelected() {
	if d, ok := a.selectedDevice(); ok && d.IP != "" {
		ip := d.IP
		go pShellExecuteW.Call(0, uintptr(unsafe.Pointer(u16("open"))), uintptr(unsafe.Pointer(u16("http://"+ip))), 0, 0, SW_SHOWNORMAL)
	}
}

func setClipboard(s string) {
	if r, _, _ := pOpenClipboard.Call(0); r == 0 {
		return
	}
	defer pCloseClipboard.Call()
	pEmptyClipboard.Call()
	w, _ := syscall.UTF16FromString(s)
	sz := uintptr(len(w) * 2)
	h, _, _ := pGlobalAlloc.Call(GMEM_MOVEABLE, sz)
	if h == 0 {
		return
	}
	p, _, _ := pGlobalLock.Call(h)
	if p == 0 {
		return
	}
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(w))
	copy(dst, w)
	pGlobalUnlock.Call(h)
	pSetClipboardData.Call(CF_UNICODETEXT, h)
}

func (a *App) invalidate() { pInvalidateRect.Call(a.hwnd, 0, 0) }
