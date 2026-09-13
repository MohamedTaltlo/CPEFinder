package com.mtradius.cpefinder;

import java.util.Arrays;
import java.util.HashSet;
import java.util.Locale;
import java.util.Set;

final class VendorDb {
    private VendorDb() {}

    private static final Set<String> UBNT = new HashSet<>(Arrays.asList(
            "00:15:6D","00:27:22","04:18:D6","0C:EA:14","18:E8:29","1C:0B:8B","1C:6A:1B",
            "24:5A:4C","24:A4:3C","28:70:4E","2C:E5:BD","44:D9:E7","58:D6:1F","60:22:32",
            "68:2E:3C","68:72:51","68:D7:9A","6C:63:F8","70:A7:41","74:83:C2","74:AC:B9",
            "74:F9:2C","74:FA:29","78:45:58","78:8A:20","80:2A:A8","84:78:48","8C:30:66",
            "8C:ED:E1","90:41:B2","94:2A:6F","9C:05:D6","A4:F8:FF","A8:9C:6C","AC:8B:A9",
            "B4:FB:E4","CC:35:D9","D0:21:F9","D4:89:C1","D8:B3:70","D8:C2:62","DC:9F:DB",
            "E0:63:DA","E4:38:83","F0:9F:C2","F4:92:BF","F4:E2:C6","FC:EC:DA"
    ));

    private static final Set<String> MIMOSA = new HashSet<>(Arrays.asList(
            "20:B5:C6","84:9C:A4","8C:B6:C5","90:70:BF","CC:54:FE","D0:B0:DF",
            "00:A0:0A","00:01:AA"
    ));

    private static final Set<String> CAMBIUM = new HashSet<>(Arrays.asList(
            "00:04:56","30:CB:C7","58:C1:7A","90:14:AF","90:6D:62","B4:A2:5C","BC:A9:93",
            "BC:E6:7C","FC:11:65","0A:00:3E"
    ));

    static String normalizeMac(String mac) {
        if (mac == null) return "";
        String m = mac.trim().replace('-', ':').toUpperCase(Locale.US);
        if (!m.contains(":") && m.length() == 12) {
            StringBuilder b = new StringBuilder(17);
            for (int i = 0; i < 12; i += 2) {
                if (b.length() > 0) b.append(':');
                b.append(m, i, i + 2);
            }
            return b.toString();
        }
        return m;
    }

    static String fromMac(String mac) {
        String m = normalizeMac(mac);
        if (m.length() < 8) return "";
        String p = m.substring(0, 8);
        if (UBNT.contains(p)) return "Ubiquiti";
        if (MIMOSA.contains(p)) return (p.equals("00:A0:0A") || p.equals("00:01:AA")) ? "Mimosa / Airspan" : "Mimosa";
        if (CAMBIUM.contains(p)) return "Cambium";
        return "";
    }

    static String fromText(String text) {
        if (text == null) return "";
        String t = text.toLowerCase(Locale.US);
        if (t.contains("ubiquiti") || t.contains("airmax") || t.contains("uisp") || t.contains("wave")) return "Ubiquiti";
        if (t.contains("mimosa")) return "Mimosa";
        if (t.contains("cambium") || t.contains("epmp") || t.contains("ptp 550") || t.contains("ptp550") || t.contains("pmp 450") || t.contains("pmp450") || t.contains("cnpilot")) return "Cambium";
        return "";
    }

    static boolean targetVendor(String vendor) {
        return vendor != null && !vendor.isEmpty();
    }
}
