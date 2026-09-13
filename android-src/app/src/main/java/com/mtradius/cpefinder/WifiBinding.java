package com.mtradius.cpefinder;

import android.content.Context;
import android.net.ConnectivityManager;
import android.net.LinkAddress;
import android.net.LinkProperties;
import android.net.Network;
import android.net.NetworkCapabilities;

import java.net.Inet4Address;
import java.net.InetAddress;
import java.util.ArrayList;
import java.util.List;

final class WifiBinding {
    final Network network;
    final Inet4Address localIp;
    final List<InetAddress> broadcasts;
    final String interfaceName;
    final int prefixLength;

    WifiBinding(Network network, Inet4Address localIp, List<InetAddress> broadcasts, String interfaceName, int prefixLength) {
        this.network = network;
        this.localIp = localIp;
        this.broadcasts = broadcasts;
        this.interfaceName = interfaceName == null ? "Wi‑Fi" : interfaceName;
        this.prefixLength = prefixLength;
    }

    static WifiBinding find(Context context) {
        ConnectivityManager cm = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);
        if (cm == null) return null;

        Network active = cm.getActiveNetwork();
        WifiBinding found = fromNetwork(cm, active);
        if (found != null) return found;

        for (Network n : cm.getAllNetworks()) {
            found = fromNetwork(cm, n);
            if (found != null) return found;
        }
        return null;
    }

    private static WifiBinding fromNetwork(ConnectivityManager cm, Network n) {
        if (n == null) return null;
        NetworkCapabilities caps = cm.getNetworkCapabilities(n);
        if (caps == null || !caps.hasTransport(NetworkCapabilities.TRANSPORT_WIFI)) return null;
        LinkProperties lp = cm.getLinkProperties(n);
        if (lp == null) return null;

        for (LinkAddress la : lp.getLinkAddresses()) {
            InetAddress addr = la.getAddress();
            if (!(addr instanceof Inet4Address) || addr.isLoopbackAddress()) continue;
            Inet4Address ip = (Inet4Address) addr;
            int prefix = la.getPrefixLength();
            List<InetAddress> bcasts = new ArrayList<>();
            try {
                bcasts.add(InetAddress.getByName("255.255.255.255"));
                InetAddress subnet = calculateBroadcast(ip, prefix);
                if (!containsAddress(bcasts, subnet)) bcasts.add(subnet);
            } catch (Exception ignored) {}
            return new WifiBinding(n, ip, bcasts, lp.getInterfaceName(), prefix);
        }
        return null;
    }

    private static boolean containsAddress(List<InetAddress> list, InetAddress wanted) {
        for (InetAddress a : list) if (a.equals(wanted)) return true;
        return false;
    }

    static InetAddress calculateBroadcast(Inet4Address ip, int prefix) throws Exception {
        byte[] raw = ip.getAddress();
        int value = ((raw[0] & 0xff) << 24) | ((raw[1] & 0xff) << 16) | ((raw[2] & 0xff) << 8) | (raw[3] & 0xff);
        int mask = prefix == 0 ? 0 : (int) (0xffffffffL << (32 - prefix));
        int b = value | ~mask;
        return intToAddress(b);
    }

    static int addressToInt(Inet4Address ip) {
        byte[] raw = ip.getAddress();
        return ((raw[0] & 0xff) << 24) | ((raw[1] & 0xff) << 16) | ((raw[2] & 0xff) << 8) | (raw[3] & 0xff);
    }

    static InetAddress intToAddress(int value) throws Exception {
        return InetAddress.getByAddress(new byte[]{(byte)(value >>> 24), (byte)(value >>> 16), (byte)(value >>> 8), (byte)value});
    }
}
