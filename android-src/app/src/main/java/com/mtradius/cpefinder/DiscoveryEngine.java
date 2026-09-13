package com.mtradius.cpefinder;

import android.content.Context;
import android.net.wifi.WifiManager;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.HttpURLConnection;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicBoolean;

final class DiscoveryEngine {
    interface Listener {
        void onDevice(Device d);
        void onState(boolean running, String message);
    }

    private final Context context;
    private final Listener listener;
    private final AtomicBoolean running = new AtomicBoolean(false);
    private final CopyOnWriteArrayList<DatagramSocket> sockets = new CopyOnWriteArrayList<>();
    private final Set<String> httpProbed = ConcurrentHashMap.newKeySet();
    private volatile ExecutorService workers;
    private volatile WifiManager.MulticastLock multicastLock;
    private volatile WifiBinding binding;

    DiscoveryEngine(Context context, Listener listener) {
        this.context = context.getApplicationContext();
        this.listener = listener;
    }

    boolean isRunning() { return running.get(); }

    void start() {
        if (!running.compareAndSet(false, true)) return;
        WifiBinding b = WifiBinding.find(context);
        if (b == null) {
            running.set(false);
            listener.onState(false, "لا يوجد Wi‑Fi مع IPv4");
            return;
        }
        binding = b;
        acquireMulticast();
        listener.onState(true, "جارٍ الفحص");
        httpProbed.clear();
        ExecutorService ex = Executors.newFixedThreadPool(10);
        workers = ex;
        ex.execute(this::ubntLoop);
        ex.execute(this::mndpLoop);
        ex.execute(this::neighborLoop);
        ex.execute(this::genericDiscoveryLoop);
        ex.execute(this::subnetProbeLoop);
        ex.execute(this::defaultProbeLoop);
    }

    void stop() {
        if (!running.getAndSet(false)) return;
        for (DatagramSocket s : sockets) try { s.close(); } catch (Throwable ignored) {}
        sockets.clear();
        ExecutorService ex = workers;
        workers = null;
        if (ex != null) ex.shutdownNow();
        releaseMulticast();
        listener.onState(false, "متوقف");
    }

    void shutdown() { stop(); }

    private void acquireMulticast() {
        try {
            WifiManager wm = (WifiManager) context.getSystemService(Context.WIFI_SERVICE);
            if (wm != null) {
                WifiManager.MulticastLock l = wm.createMulticastLock("CPEFinderV2");
                l.setReferenceCounted(false);
                l.acquire();
                multicastLock = l;
            }
        } catch (Throwable ignored) {}
    }

    private void releaseMulticast() {
        try {
            WifiManager.MulticastLock l = multicastLock;
            multicastLock = null;
            if (l != null && l.isHeld()) l.release();
        } catch (Throwable ignored) {}
    }

    private DatagramSocket socket(int timeoutMs) throws Exception {
        WifiBinding b = binding;
        DatagramSocket s = new DatagramSocket(null);
        s.setReuseAddress(true);
        s.setBroadcast(true);
        b.network.bindSocket(s);
        s.bind(new InetSocketAddress(b.localIp, 0));
        s.setSoTimeout(timeoutMs);
        sockets.add(s);
        return s;
    }

    private void close(DatagramSocket s) {
        sockets.remove(s);
        if (s != null) try { s.close(); } catch (Throwable ignored) {}
    }

    private void ubntLoop() {
        while (running.get()) {
            DatagramSocket s = null;
            try {
                s = socket(220);
                List<InetAddress> targets = new ArrayList<>(binding.broadcasts);
                try { targets.add(InetAddress.getByName("233.89.188.1")); } catch (Exception ignored) {}
                for (InetAddress t : targets) for (byte[] q : UbntProtocol.PROBES) {
                    if (!running.get()) break;
                    try { s.send(new DatagramPacket(q, q.length, t, 10001)); } catch (Throwable ignored) {}
                }
                long until = System.currentTimeMillis() + 1800;
                byte[] buf = new byte[8192];
                while (running.get() && System.currentTimeMillis() < until) {
                    DatagramPacket p = new DatagramPacket(buf, buf.length);
                    try { s.receive(p); } catch (Throwable e) { continue; }
                    Device d = UbntProtocol.parse(p.getData(), p.getLength());
                    if (d != null) {
                        if (d.ip.isEmpty() && p.getAddress() != null) d.ip = p.getAddress().getHostAddress();
                        emit(d);
                    }
                }
            } catch (Throwable ignored) {
            } finally { close(s); }
            sleep(2200);
        }
    }

    private void mndpLoop() {
        while (running.get()) {
            DatagramSocket s = null;
            try {
                s = socket(220);
                for (InetAddress t : binding.broadcasts) {
                    try { s.send(new DatagramPacket(MndpProtocol.PROBE, MndpProtocol.PROBE.length, t, 5678)); } catch (Throwable ignored) {}
                }
                long until = System.currentTimeMillis() + 1700;
                byte[] buf = new byte[4096];
                while (running.get() && System.currentTimeMillis() < until) {
                    DatagramPacket p = new DatagramPacket(buf, buf.length);
                    try { s.receive(p); } catch (Throwable e) { continue; }
                    Device d = MndpProtocol.parse(p.getData(), p.getLength(), p.getAddress() == null ? "" : p.getAddress().getHostAddress());
                    if (d != null) emit(d);
                }
            } catch (Throwable ignored) {
            } finally { close(s); }
            sleep(2200);
        }
    }

    private void neighborLoop() {
        while (running.get()) {
            try {
                Map<String, NeighborTable.Entry> table = NeighborTable.read(binding.interfaceName);
                for (NeighborTable.Entry e : table.values()) {
                    String vendor = VendorDb.fromMac(e.mac);
                    if (VendorDb.targetVendor(vendor)) {
                        Device d = new Device();
                        d.ip = e.ip; d.mac = e.mac; d.vendor = vendor; d.method = "ARP / Neighbor";
                        d.lastSeen = System.currentTimeMillis(); d.key = "mac:" + e.mac;
                        emit(d); scheduleHttpProbe(d);
                    }
                }
            } catch (Throwable ignored) {}
            sleep(700);
        }
    }

    private void genericDiscoveryLoop() {
        while (running.get()) {
            ssdpRound();
            wsDiscoveryRound();
            mdnsRound();
            sleep(3500);
        }
    }

    private void ssdpRound() {
        DatagramSocket s = null;
        try {
            s = socket(250);
            String q = "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nMAN: \"ssdp:discover\"\r\nMX: 1\r\nST: ssdp:all\r\n\r\n";
            InetAddress dst = InetAddress.getByName("239.255.255.250");
            byte[] data = q.getBytes(StandardCharsets.US_ASCII);
            s.send(new DatagramPacket(data, data.length, dst, 1900));
            receiveTextDiscovery(s, 900, "SSDP");
        } catch (Throwable ignored) { } finally { close(s); }
    }

    private void wsDiscoveryRound() {
        DatagramSocket s = null;
        try {
            s = socket(250);
            String id = Long.toHexString(System.nanoTime());
            String xml = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><e:Envelope xmlns:e=\"http://www.w3.org/2003/05/soap-envelope\" xmlns:w=\"http://schemas.xmlsoap.org/ws/2004/08/addressing\" xmlns:d=\"http://schemas.xmlsoap.org/ws/2005/04/discovery\"><e:Header><w:MessageID>uuid:"+id+"</w:MessageID><w:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</w:To><w:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</w:Action></e:Header><e:Body><d:Probe/></e:Body></e:Envelope>";
            byte[] data = xml.getBytes(StandardCharsets.UTF_8);
            s.send(new DatagramPacket(data, data.length, InetAddress.getByName("239.255.255.250"), 3702));
            receiveTextDiscovery(s, 800, "WS-Discovery");
        } catch (Throwable ignored) { } finally { close(s); }
    }

    private void mdnsRound() {
        DatagramSocket s = null;
        try {
            s = socket(250);
            byte[] q = mdnsQuery("_services._dns-sd._udp.local");
            s.send(new DatagramPacket(q, q.length, InetAddress.getByName("224.0.0.251"), 5353));
            receiveTextDiscovery(s, 700, "mDNS");
        } catch (Throwable ignored) { } finally { close(s); }
    }

    private void receiveTextDiscovery(DatagramSocket s, long duration, String method) {
        long until = System.currentTimeMillis() + duration;
        byte[] buf = new byte[8192];
        while (running.get() && System.currentTimeMillis() < until) {
            DatagramPacket p = new DatagramPacket(buf, buf.length);
            try { s.receive(p); } catch (Throwable e) { continue; }
            String text = new String(p.getData(), 0, p.getLength(), StandardCharsets.ISO_8859_1);
            String vendor = VendorDb.fromText(text);
            String ip = p.getAddress() == null ? "" : p.getAddress().getHostAddress();
            String mac = NetUtil.validIpv4(ip) ? NeighborTable.macFor(ip, binding.interfaceName) : "";
            if (vendor.isEmpty()) vendor = VendorDb.fromMac(mac);
            if (!VendorDb.targetVendor(vendor)) continue;
            Device d = new Device();
            d.ip = ip; d.mac = mac; d.vendor = vendor; d.method = method; d.description = compactText(text);
            d.lastSeen = System.currentTimeMillis(); d.key = !mac.isEmpty() ? "mac:" + VendorDb.normalizeMac(mac) : "ip:" + ip;
            emit(d); scheduleHttpProbe(d);
        }
    }

    private void subnetProbeLoop() {
        int cursor = 1;
        while (running.get()) {
            try { cursor = probeSubnetChunk(cursor, 4096); } catch (Throwable ignored) {}
            sleep(900);
        }
    }

    private int probeSubnetChunk(int cursor, int maxCount) throws Exception {
        WifiBinding b = binding;
        if (b == null || b.prefixLength < 16 || b.prefixLength > 30) return cursor;
        int hostBits = 32 - b.prefixLength;
        int total = 1 << hostBits;
        int ip = WifiBinding.addressToInt(b.localIp);
        int mask = b.prefixLength == 0 ? 0 : (int)(0xffffffffL << hostBits);
        int network = ip & mask;
        if (cursor <= 0 || cursor >= total - 1) cursor = 1;
        DatagramSocket s = socket(80);
        try {
            byte[] payload = new byte[]{0};
            int done = 0;
            while (running.get() && cursor < total - 1 && done < maxCount) {
                int target = network + cursor;
                if (target != ip) {
                    InetAddress a = WifiBinding.intToAddress(target);
                    try { s.send(new DatagramPacket(payload, payload.length, a, 9)); } catch (Throwable ignored) {}
                }
                cursor++; done++;
            }
        } finally { close(s); }
        if (cursor >= total - 1) cursor = 1;
        return cursor;
    }

    private void defaultProbeLoop() {
        List<String> ips = Arrays.asList("192.168.1.20", "192.168.1.10", "169.254.1.1", "192.168.0.20", "192.168.0.1");
        while (running.get()) {
            DatagramSocket s = null;
            try {
                s = socket(100);
                byte[] one = new byte[]{0};
                for (String ip : ips) {
                    if (!running.get()) break;
                    try { s.send(new DatagramPacket(one, one.length, InetAddress.getByName(ip), 9)); } catch (Throwable ignored) {}
                }
            } catch (Throwable ignored) { } finally { close(s); }
            sleep(5000);
        }
    }

    private void scheduleHttpProbe(Device d) {
        if (!running.get() || d == null || !NetUtil.validIpv4(d.ip)) return;
        String k = d.ip;
        if (!httpProbed.add(k)) return;
        ExecutorService ex = workers;
        if (ex == null) return;
        ex.execute(() -> {
            try { httpProbe(d.ip, false); } catch (Throwable ignored) {}
            try { httpProbe(d.ip, true); } catch (Throwable ignored) {}
        });
    }

    private void httpProbe(String ip, boolean https) {
        if (!running.get()) return;
        HttpURLConnection c = null;
        try {
            URL u = new URL((https ? "https" : "http") + "://" + ip + "/");
            c = (HttpURLConnection) binding.network.openConnection(u);
            c.setConnectTimeout(550); c.setReadTimeout(550); c.setInstanceFollowRedirects(false); c.setRequestMethod("GET");
            c.connect();
            String server = c.getHeaderField("Server");
            StringBuilder body = new StringBuilder();
            try (BufferedReader br = new BufferedReader(new InputStreamReader(c.getInputStream()))) {
                char[] buf = new char[1024]; int n; int total = 0;
                while ((n = br.read(buf)) > 0 && total < 8192) { body.append(buf, 0, n); total += n; }
            } catch (Throwable ignored) {}
            String text = (server == null ? "" : server) + " " + body;
            String vendor = VendorDb.fromText(text);
            String mac = NeighborTable.macFor(ip, binding.interfaceName);
            if (vendor.isEmpty()) vendor = VendorDb.fromMac(mac);
            if (!VendorDb.targetVendor(vendor)) return;
            Device d = new Device(); d.ip = ip; d.mac = mac; d.vendor = vendor; d.method = https ? "HTTPS" : "HTTP";
            d.description = pageTitle(body.toString()); d.lastSeen = System.currentTimeMillis();
            d.key = !mac.isEmpty() ? "mac:" + VendorDb.normalizeMac(mac) : "ip:" + ip;
            emit(d);
        } catch (Throwable ignored) { } finally { if (c != null) c.disconnect(); }
    }

    private void emit(Device d) {
        if (d == null) return;
        d.mac = VendorDb.normalizeMac(d.mac);
        if (d.vendor.isEmpty()) d.vendor = VendorDb.fromMac(d.mac);
        if (d.vendor.isEmpty()) d.vendor = VendorDb.fromText(d.model + " " + d.hostname + " " + d.description);
        if (!VendorDb.targetVendor(d.vendor)) return;
        if (d.key.isEmpty()) d.key = !d.mac.isEmpty() ? "mac:" + d.mac : (!d.ip.isEmpty() ? "ip:" + d.ip : "");
        if (d.key.isEmpty()) return;
        d.lastSeen = System.currentTimeMillis();
        listener.onDevice(d);
    }

    private void sleep(long ms) {
        long end = System.currentTimeMillis() + ms;
        while (running.get() && System.currentTimeMillis() < end) {
            try { Thread.sleep(Math.min(150, end - System.currentTimeMillis())); }
            catch (InterruptedException e) { Thread.currentThread().interrupt(); return; }
        }
    }

    private static byte[] mdnsQuery(String name) {
        byte[] out = new byte[512];
        int p = 0;
        out[p++]=0; out[p++]=0; out[p++]=0; out[p++]=0; out[p++]=0; out[p++]=1; out[p++]=0; out[p++]=0; out[p++]=0; out[p++]=0; out[p++]=0; out[p++]=0;
        for (String label : name.split("\\.")) {
            byte[] b = label.getBytes(StandardCharsets.UTF_8); out[p++] = (byte)b.length; System.arraycopy(b,0,out,p,b.length); p += b.length;
        }
        out[p++] = 0; out[p++] = 0; out[p++] = 12; out[p++] = 0; out[p++] = 1;
        return Arrays.copyOf(out, p);
    }

    private static String compactText(String s) {
        if (s == null) return "";
        String x = s.replace('\r',' ').replace('\n',' ').replace('\u0000',' ').replaceAll("\\s+", " ").trim();
        return x.length() > 180 ? x.substring(0,180) : x;
    }

    private static String pageTitle(String html) {
        if (html == null) return "";
        String l = html.toLowerCase(Locale.US);
        int a = l.indexOf("<title"); if (a < 0) return "";
        a = l.indexOf('>', a); if (a < 0) return "";
        int b = l.indexOf("</title>", a); if (b < 0) return "";
        return html.substring(a+1,b).replaceAll("<[^>]+>", "").trim();
    }
}
