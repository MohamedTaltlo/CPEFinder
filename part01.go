// SPDX-License-Identifier: MIT
// Copyright (c) 2026 CPEFinder contributors

//go:build windows

package main

var appVersion = "0.1.0"

const (
	appTitle = "كاشف عناوين أجهزة الشبكة"

	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_TABSTOP          = 0x00010000
	WS_CLIPCHILDREN     = 0x02000000
	WS_EX_LAYOUTRTL     = 0x00400000
	WS_EX_RTLREADING    = 0x00002000
	WS_EX_CLIENTEDGE    = 0x00000200

	BS_OWNERDRAW     = 0x0000000B
	CBS_DROPDOWNLIST = 0x0003
	CBS_HASSTRINGS   = 0x0200
	ES_AUTOHSCROLL   = 0x0080
	ES_RIGHT         = 0x0002

	LVS_REPORT           = 0x0001
	LVS_SINGLESEL        = 0x0004
	LVS_SHOWSELALWAYS    = 0x0008
	LVS_NOSORTHEADER     = 0x8000
	LVS_EX_FULLROWSELECT = 0x00000020
	LVS_EX_GRIDLINES     = 0x00000001
	LVS_EX_DOUBLEBUFFER  = 0x00010000

	WM_CREATE       = 0x0001
	WM_DESTROY      = 0x0002
	WM_PAINT        = 0x000F
	WM_CLOSE        = 0x0010
	WM_COMMAND      = 0x0111
	WM_TIMER        = 0x0113
	WM_DRAWITEM     = 0x002B
	WM_ERASEBKGND   = 0x0014
	WM_SETFONT      = 0x0030
	WM_NOTIFY       = 0x004E
	WM_SIZE         = 0x0005
	WM_SETICON      = 0x0080
	WM_APP_ADAPTERS = 0x8001
	WM_APP_EXIT     = 0x8002

	ICON_SMALL = 0
	ICON_BIG   = 1

	SW_SHOW       = 5
	SW_SHOWNORMAL = 1

	COLOR_WINDOW = 5
	IDC_ARROW    = 32512

	DT_CENTER       = 0x00000001
	DT_RIGHT        = 0x00000002
	DT_VCENTER      = 0x00000004
	DT_SINGLELINE   = 0x00000020
	DT_RTLREADING   = 0x00020000
	DT_END_ELLIPSIS = 0x00008000

	TRANSPARENT = 1

	GWLP_USERDATA = -21

	LVM_FIRST                    = 0x1000
	LVM_SETEXTENDEDLISTVIEWSTYLE = LVM_FIRST + 54
	LVM_INSERTCOLUMNW            = LVM_FIRST + 97
	LVM_INSERTITEMW              = LVM_FIRST + 77
	LVM_SETITEMW                 = LVM_FIRST + 76
	LVM_DELETEALLITEMS           = LVM_FIRST + 9
	LVM_GETNEXTITEM              = LVM_FIRST + 12
	LVNI_SELECTED                = 0x0002
	LVIF_TEXT                    = 0x0001
	LVCF_FMT                     = 0x0001
	LVCF_WIDTH                   = 0x0002
	LVCF_TEXT                    = 0x0004
	LVCF_SUBITEM                 = 0x0008
	LVCFMT_RIGHT                 = 0x0001

	CB_ADDSTRING    = 0x0143
	CB_RESETCONTENT = 0x014B
	CB_SETCURSEL    = 0x014E
	CB_GETCURSEL    = 0x0147

	EM_SETCUEBANNER = 0x1501

	NM_DBLCLK = -3

	ID_SCAN    = 1001
	ID_STOP    = 1002
	ID_REFRESH = 1003
	ID_CLEAR   = 1004
	ID_COPY    = 1005
	ID_OPEN    = 1006
	ID_ADAPTER = 1007
	ID_SEARCH  = 1008

	GMEM_MOVEABLE  = 0x0002
	CF_UNICODETEXT = 13
)

type POINT struct{ X, Y int32 }

type RECT struct{ Left, Top, Right, Bottom int32 }

type MSG struct {
	HWnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             POINT
	LPrivate       uint32
}

type WNDCLASSEX struct {
	CbSize                                   uint32
	Style                                    uint32
	LpfnWndProc                              uintptr
	CbClsExtra, CbWndExtra                   int32
	HInstance, HIcon, HCursor, HbrBackground uintptr
	LpszMenuName, LpszClassName              *uint16
	HIconSm                                  uintptr
}

type PAINTSTRUCT struct {
	Hdc                  uintptr
	FErase               int32
	RcPaint              RECT
	FRestore, FIncUpdate int32
	RgbReserved          [32]byte
}

type LVCOLUMN struct {
	Mask       uint32
	Fmt        int32
	Cx         int32
	PszText    *uint16
	CchTextMax int32
	ISubItem   int32
	IImage     int32
	IOrder     int32
	CxMin      int32
	CxDefault  int32
	CxIdeal    int32
}

type LVITEM struct {
	Mask             uint32
	IItem, ISubItem  int32
	State, StateMask uint32
	PszText          *uint16
	CchTextMax       int32
	IImage           int32
	LParam           uintptr
	IIndent          int32
	IGroupId         int32
	CColumns         uint32
	PuColumns        *uint32
	PiColFmt         *int32
	IGroup           int32
}

type NMHDR struct {
	HwndFrom uintptr
	IdFrom   uintptr
	Code     int32
}

type DRAWITEMSTRUCT struct {
	CtlType, CtlID, ItemID, ItemAction, ItemState uint32
	HwndItem, HDC                                 uintptr
	RcItem                                        RECT
	ItemData                                      uintptr
}

type INITCOMMONCONTROLSEX struct {
	DwSize uint32
	DwICC  uint32
}
