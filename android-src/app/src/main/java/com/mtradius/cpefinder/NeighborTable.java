package com.mtradius.cpefinder;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.InputStreamReader;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

final class NeighborTable {
    private NeighborTable() {}

    static final class Entry {
        final String ip;
        final String mac;
        Entry(String ip, String mac) { this.ip = ip; this.mac = VendorDb.normalizeMac(mac); }
    }

    static Map<String, Entry> read(String iface) {
        LinkedHashMap<String, Entry> out = new LinkedHashMap<>();
        readProc(out, iface);
        readIpNeigh(out, iface);
        return out;
    }

    static String macFor(String ip, String iface) {
        Entry e = read(iface).get(ip);
        return e == null ? "" : e.mac;
    }

    private static void readProc(Map<String, Entry> out, String iface) {
        File f = new File("/proc/net/arp");
        if (!f.canRead()) return;
        try (BufferedReader br = new BufferedReader(new FileReader(f))) {
            String line;
            while ((line = br.readLine()) != null) {
                String[] p = line.trim().split("\\s+");
                if (p.length < 6 || p[0].equalsIgnoreCase("IP")) continue;
                if (iface != null && !iface.isEmpty() && !iface.equals(p[5])) continue;
                String mac = VendorDb.normalizeMac(p[3]);
                if (NetUtil.validIpv4(p[0]) && mac.matches("[0-9A-F]{2}(:[0-9A-F]{2}){5}") && !mac.equals("00:00:00:00:00:00")) {
                    out.put(p[0], new Entry(p[0], mac));
                }
            }
        } catch (Throwable ignored) {}
    }

    private static void readIpNeigh(Map<String, Entry> out, String iface) {
        Process p = null;
        try {
            ProcessBuilder pb;
            if (iface != null && !iface.isEmpty()) pb = new ProcessBuilder("/system/bin/ip", "neigh", "show", "dev", iface);
            else pb = new ProcessBuilder("/system/bin/ip", "neigh", "show");
            pb.redirectErrorStream(true);
            p = pb.start();
            try (BufferedReader br = new BufferedReader(new InputStreamReader(p.getInputStream()))) {
                String line;
                while ((line = br.readLine()) != null) {
                    String[] a = line.trim().split("\\s+");
                    if (a.length < 3 || !NetUtil.validIpv4(a[0])) continue;
                    String mac = "";
                    for (int i = 1; i + 1 < a.length; i++) {
                        if (a[i].equalsIgnoreCase("lladdr")) { mac = VendorDb.normalizeMac(a[i+1]); break; }
                    }
                    if (mac.matches("[0-9A-F]{2}(:[0-9A-F]{2}){5}")) out.put(a[0], new Entry(a[0], mac));
                }
            }
            p.waitFor(250, TimeUnit.MILLISECONDS);
        } catch (Throwable ignored) {
        } finally {
            if (p != null) try { p.destroy(); } catch (Throwable ignored) {}
        }
    }
}
