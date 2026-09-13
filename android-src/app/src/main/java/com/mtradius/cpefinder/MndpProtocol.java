package com.mtradius.cpefinder;

import java.nio.charset.StandardCharsets;
import java.util.Locale;

final class MndpProtocol {
    private MndpProtocol() {}
    static final byte[] PROBE = new byte[]{0,0,0,0};

    static Device parse(byte[] data, int len, String sourceIp) {
        if (data == null || len < 8) return null;
        Device d = new Device();
        int off = 4;
        while (off + 4 <= len) {
            int type = ((data[off]&0xff) << 8) | (data[off+1]&0xff);
            int l = ((data[off+2]&0xff) << 8) | (data[off+3]&0xff);
            off += 4;
            if (l < 0 || off + l > len) break;
            switch (type) {
                case 1:
                    if (l >= 6) d.mac = mac(data, off);
                    break;
                case 5:
                    d.hostname = text(data, off, l);
                    break;
                case 7:
                    d.firmware = text(data, off, l);
                    break;
                case 8:
                    d.description = text(data, off, l);
                    break;
                case 12:
                    d.model = text(data, off, l);
                    break;
                case 17:
                    if (l == 4) d.ip = String.format(Locale.US, "%d.%d.%d.%d", data[off]&0xff,data[off+1]&0xff,data[off+2]&0xff,data[off+3]&0xff);
                    break;
                default: break;
            }
            off += l;
        }
        d.mac = VendorDb.normalizeMac(d.mac);
        String vendor = VendorDb.fromMac(d.mac);
        if (vendor.isEmpty()) vendor = VendorDb.fromText(d.description + " " + d.model + " " + d.hostname);
        if (!vendor.toLowerCase(Locale.US).contains("cambium")) return null;
        d.vendor = "Cambium";
        if (!NetUtil.validIpv4(d.ip) && NetUtil.validIpv4(sourceIp)) d.ip = sourceIp;
        d.method = "Cambium MNDP";
        d.lastSeen = System.currentTimeMillis();
        d.key = !d.mac.isEmpty() ? "mac:" + d.mac : (!d.ip.isEmpty() ? "ip:" + d.ip : "");
        return d.key.isEmpty() ? null : d;
    }

    private static String mac(byte[] b, int o) {
        if (o + 6 > b.length) return "";
        return String.format(Locale.US, "%02X:%02X:%02X:%02X:%02X:%02X", b[o]&0xff,b[o+1]&0xff,b[o+2]&0xff,b[o+3]&0xff,b[o+4]&0xff,b[o+5]&0xff);
    }

    private static String text(byte[] b, int o, int l) {
        return new String(b, o, l, StandardCharsets.UTF_8).replace("\u0000", "").trim();
    }
}
