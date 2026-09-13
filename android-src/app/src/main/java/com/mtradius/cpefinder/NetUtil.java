package com.mtradius.cpefinder;

import java.net.InetAddress;

final class NetUtil {
    private NetUtil() {}

    static boolean validIpv4(String ip) {
        if (ip == null || ip.isEmpty()) return false;
        try {
            InetAddress a = InetAddress.getByName(ip);
            byte[] b = a.getAddress();
            return b.length == 4 && !ip.equals("0.0.0.0") && !ip.equals("255.255.255.255");
        } catch (Exception e) { return false; }
    }

    static int score(String ip) {
        if (!validIpv4(ip)) return -100;
        String[] p = ip.split("\\.");
        int a = Integer.parseInt(p[0]), b = Integer.parseInt(p[1]);
        if (a == 169 && b == 254) return 10;
        if (a == 0 || a >= 224) return -50;
        return 100;
    }

    static String chooseIp(String oldIp, String newIp) {
        if (!validIpv4(newIp)) return oldIp == null ? "" : oldIp;
        if (!validIpv4(oldIp) || score(newIp) > score(oldIp)) return newIp;
        return oldIp;
    }
}
