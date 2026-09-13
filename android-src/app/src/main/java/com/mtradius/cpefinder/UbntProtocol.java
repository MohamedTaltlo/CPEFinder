package com.mtradius.cpefinder;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Locale;

final class UbntProtocol {
    private UbntProtocol() {}

    static final byte[][] PROBES = new byte[][]{
            new byte[]{1,0,0,0},
            new byte[]{2,0,0,0}
    };

    static Device parse(byte[] data, int len) {
        if (data == null || len < 4) return null;
        int version = data[0] & 0xff;
        if (version != 1 && version != 2) return null;

        Device d = new Device();
        List<String> ips = new ArrayList<>();
        int off = 4;
        while (off + 3 <= len) {
            int type = data[off] & 0xff;
            int l = ((data[off + 1] & 0xff) << 8) | (data[off + 2] & 0xff);
            off += 3;
            if (l < 0 || off + l > len) break;
            byte[] v = Arrays.copyOfRange(data, off, off + l);
            off += l;
            switch (type) {
                case 0x01:
                    if (v.length >= 6) d.mac = mac(v, 0);
                    break;
                case 0x02:
                    for (int p = 0; p + 10 <= v.length; p += 10) {
                        if (d.mac.isEmpty()) d.mac = mac(v, p);
                        String ip = String.format(Locale.US, "%d.%d.%d.%d", v[p+6]&0xff, v[p+7]&0xff, v[p+8]&0xff, v[p+9]&0xff);
                        if (NetUtil.validIpv4(ip)) ips.add(ip);
                    }
                    break;
                case 0x03: d.firmware = clean(v); break;
                case 0x0B: d.hostname = clean(v); break;
                case 0x0C: if (d.model.isEmpty()) d.model = clean(v); break;
                case 0x0D: d.ssid = clean(v); break;
                case 0x14:
                case 0x15:
                    String model = clean(v);
                    if (!model.isEmpty()) d.model = model;
                    break;
                case 0x1B:
                    if (d.firmware.isEmpty()) d.firmware = clean(v);
                    break;
                default: break;
            }
        }
        String best = "";
        for (String ip : ips) best = NetUtil.chooseIp(best, ip);
        d.ip = best;
        d.mac = VendorDb.normalizeMac(d.mac);
        d.vendor = "Ubiquiti";
        d.method = "Ubiquiti Discovery";
        d.lastSeen = System.currentTimeMillis();
        d.key = !d.mac.isEmpty() ? "mac:" + d.mac : (!d.ip.isEmpty() ? "ip:" + d.ip : "");
        return d.key.isEmpty() ? null : d;
    }

    static String mac(byte[] b, int o) {
        if (o + 6 > b.length) return "";
        return String.format(Locale.US, "%02X:%02X:%02X:%02X:%02X:%02X", b[o]&0xff,b[o+1]&0xff,b[o+2]&0xff,b[o+3]&0xff,b[o+4]&0xff,b[o+5]&0xff);
    }

    private static String clean(byte[] b) {
        int start = 0, end = b.length;
        while (start < end && (b[start] == 0 || b[start] == ' ' || b[start] == '\r' || b[start] == '\n')) start++;
        while (end > start && (b[end-1] == 0 || b[end-1] == ' ' || b[end-1] == '\r' || b[end-1] == '\n')) end--;
        return new String(b, start, end - start, StandardCharsets.UTF_8).trim();
    }
}
