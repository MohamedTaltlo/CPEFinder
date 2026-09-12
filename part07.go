// SPDX-License-Identifier: MIT
// Copyright (c) 2026 CPEFinder contributors

//go:build windows

package main

import (
	"strconv"
	"syscall"
	"unsafe"
)

func (a *App) layout() {
	if a.hwnd == 0 {
		return
	}
	var r RECT
	pGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&r)))
	w := int(r.Right - r.Left)
	h := int(r.Bottom - r.Top)
	if w < 760 {
		return
	}
	if h < 540 {
		return
	}
	comboX := 530
	comboW := w - comboX - 35
	if comboW < 220 {
		comboW = 220
	}
	pMoveWindow.Call(a.combo, uintptr(comboX), 188, uintptr(comboW), 34, 1)
	pMoveWindow.Call(a.btnClear, 35, 188, 95, 34, 1)
	pMoveWindow.Call(a.btnStop, 145, 188, 105, 34, 1)
	pMoveWindow.Call(a.btnScan, 265, 188, 120, 34, 1)
	pMoveWindow.Call(a.btnRefresh, 400, 188, 115, 34, 1)
	pMoveWindow.Call(a.btnOpen, 35, 300, 115, 34, 1)
	pMoveWindow.Call(a.btnCopy, 160, 300, 105, 34, 1)
	pMoveWindow.Call(a.list, 35, 345, uintptr(w-70), uintptr(h-405), 1)
}

func (a *App) paint() {
	var ps PAINTSTRUCT
	hdc, _, _ := pBeginPaint.Call(a.hwnd, uintptr(unsafe.Pointer(&ps)))
	defer pEndPaint.Call(a.hwnd, uintptr(unsafe.Pointer(&ps)))
	var r RECT
	pGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&r)))
	w := int(r.Right)
	fill(hdc, RECT{0, 0, r.Right, r.Bottom}, rgb(244, 247, 251))
	fill(hdc, RECT{0, 0, r.Right, 145}, rgb(15, 27, 49))
	fill(hdc, RECT{r.Right - 5, 0, r.Right, 145}, rgb(45, 112, 243))
	drawRoundCard(hdc, 25, 165, w-25, 245, 12, rgb(255, 255, 255))
	drawRoundCard(hdc, 25, 275, w-25, int(r.Bottom)-25, 12, rgb(255, 255, 255))
	selectFont(hdc, a.fontTitle)
	drawText(hdc, "كاشف عناوين أجهزة الشبكة", RECT{40, 35, r.Right - 35, 78}, rgb(255, 255, 255), DT_RIGHT|DT_VCENTER|DT_SINGLELINE|DT_RTLREADING)
	selectFont(hdc, a.font)
	drawText(hdc, "Ubiquiti • airMAX AC • Wave • Mimosa • Cambium", RECT{40, 82, r.Right - 35, 112}, rgb(148, 163, 184), DT_RIGHT|DT_VCENTER|DT_SINGLELINE)
	selectFont(hdc, a.fontBold)
	drawText(hdc, "كرت الشبكة", RECT{int32(w - 220), 158, int32(w - 35), 184}, rgb(35, 45, 62), DT_RIGHT|DT_VCENTER|DT_SINGLELINE|DT_RTLREADING)
	drawText(hdc, "الأجهزة المكتشفة", RECT{int32(w - 240), 270, int32(w - 35), 300}, rgb(35, 45, 62), DT_RIGHT|DT_VCENTER|DT_SINGLELINE|DT_RTLREADING)
	badgeX := w - 360
	drawRoundCard(hdc, badgeX, 282, badgeX+58, 313, 10, rgb(238, 245, 255))
	selectFont(hdc, a.fontBold)
	drawText(hdc, strconv.Itoa(a.count), RECT{int32(badgeX), 282, int32(badgeX + 58), 313}, rgb(45, 112, 243), DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	status := "متوقف"
	col := rgb(100, 116, 139)
	bg := rgb(236, 240, 245)
	if a.scanning {
		status = "جارٍ الفحص"
		col = rgb(6, 122, 85)
		bg = rgb(220, 252, 231)
	}
	drawRoundCard(hdc, 35, 150, 150, 178, 9, bg)
	drawText(hdc, status, RECT{35, 150, 150, 178}, col, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_RTLREADING)
}

func fill(hdc uintptr, r RECT, color uintptr) {
	b, _, _ := pCreateSolidBrush.Call(color)
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), b)
	pDeleteObject.Call(b)
}

func drawRoundCard(hdc uintptr, l, t, r, b, rad int, color uintptr) {
	brush, _, _ := pCreateSolidBrush.Call(color)
	pen, _, _ := pCreatePen.Call(0, 1, color)
	oldB, _, _ := pSelectObject.Call(hdc, brush)
	oldP, _, _ := pSelectObject.Call(hdc, pen)
	pRoundRect.Call(hdc, uintptr(l), uintptr(t), uintptr(r), uintptr(b), uintptr(rad), uintptr(rad))
	pSelectObject.Call(hdc, oldB)
	pSelectObject.Call(hdc, oldP)
	pDeleteObject.Call(brush)
	pDeleteObject.Call(pen)
}

func selectFont(hdc, font uintptr) { pSelectObject.Call(hdc, font) }

func drawText(hdc uintptr, text string, r RECT, color uintptr, flags uintptr) {
	pSetBkMode.Call(hdc, TRANSPARENT)
	pSetTextColor.Call(hdc, color)
	pDrawTextW.Call(hdc, uintptr(unsafe.Pointer(u16(text))), ^uintptr(0), uintptr(unsafe.Pointer(&r)), flags)
}

func (a *App) drawButton(di *DRAWITEMSTRUCT) uintptr {
	if di == nil {
		return 0
	}
	bg := rgb(71, 85, 105)
	switch di.CtlID {
	case ID_SCAN:
		bg = rgb(5, 150, 105)
	case ID_STOP:
		bg = rgb(220, 38, 38)
	case ID_REFRESH:
		bg = rgb(45, 112, 243)
	case ID_COPY:
		bg = rgb(13, 148, 136)
	case ID_OPEN:
		bg = rgb(124, 58, 237)
	case ID_CLEAR:
		bg = rgb(100, 116, 139)
	}
	if di.ItemState&0x0001 != 0 {
	}
	brush, _, _ := pCreateSolidBrush.Call(bg)
	pen, _, _ := pCreatePen.Call(0, 1, bg)
	ob, _, _ := pSelectObject.Call(di.HDC, brush)
	op, _, _ := pSelectObject.Call(di.HDC, pen)
	pRoundRect.Call(di.HDC, uintptr(di.RcItem.Left), uintptr(di.RcItem.Top), uintptr(di.RcItem.Right), uintptr(di.RcItem.Bottom), 10, 10)
	pSelectObject.Call(di.HDC, ob)
	pSelectObject.Call(di.HDC, op)
	pDeleteObject.Call(brush)
	pDeleteObject.Call(pen)
	n, _, _ := pGetWindowTextLengthW.Call(di.HwndItem)
	buf := make([]uint16, int(n)+1)
	pGetWindowTextW.Call(di.HwndItem, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	txt := syscall.UTF16ToString(buf)
	selectFont(di.HDC, a.fontBold)
	drawText(di.HDC, txt, di.RcItem, rgb(255, 255, 255), DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_RTLREADING)
	return 1
}
