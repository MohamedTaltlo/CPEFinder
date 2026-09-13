package com.mtradius.cpefinder;

import android.app.Activity;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.widget.Button;
import android.widget.ListView;
import android.widget.TextView;
import android.widget.Toast;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicBoolean;

public final class MainActivity extends Activity implements DiscoveryEngine.Listener {
    private final Handler main = new Handler(Looper.getMainLooper());
    private final Map<String, Device> devices = new LinkedHashMap<>();
    private final AtomicBoolean renderQueued = new AtomicBoolean(false);

    private TextView txtStatus, txtNetwork, txtCount;
    private Button btnStart, btnStop, btnRefresh, btnClear;
    private ListView list;
    private DeviceAdapter adapter;
    private DiscoveryEngine discovery;

    @Override protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        getWindow().setStatusBarColor(Color.rgb(15,27,49));
        setContentView(R.layout.activity_main);

        txtStatus = findViewById(R.id.txtStatus);
        txtNetwork = findViewById(R.id.txtNetwork);
        txtCount = findViewById(R.id.txtCount);
        btnStart = findViewById(R.id.btnStart);
        btnStop = findViewById(R.id.btnStop);
        btnRefresh = findViewById(R.id.btnRefresh);
        btnClear = findViewById(R.id.btnClear);
        list = findViewById(R.id.listDevices);

        adapter = new DeviceAdapter(this);
        list.setAdapter(adapter);
        discovery = new DiscoveryEngine(this, this);

        btnStop.setEnabled(false);
        btnStart.setEnabled(false);
        btnStart.setOnClickListener(v -> startScan());
        btnStop.setOnClickListener(v -> stopScan());
        btnRefresh.setOnClickListener(v -> refreshNetwork());
        btnClear.setOnClickListener(v -> clearDevices());
        list.setOnItemClickListener((parent, view, position, id) -> openDevice(adapter.get(position)));
        list.setOnItemLongClickListener((parent, view, position, id) -> { copyIp(adapter.get(position)); return true; });
        refreshNetwork();
    }

    private void refreshNetwork() {
        btnRefresh.setEnabled(false);
        txtNetwork.setText("جارٍ قراءة اتصال Wi‑Fi...");
        new Thread(() -> {
            WifiBinding b = WifiBinding.find(getApplicationContext());
            main.post(() -> {
                btnRefresh.setEnabled(true);
                if (b == null) {
                    txtNetwork.setText("لا يوجد Wi‑Fi مع عنوان IPv4. ثبّت IP يدوياً ثم اضغط تحديث.");
                    if (!discovery.isRunning()) btnStart.setEnabled(false);
                } else {
                    txtNetwork.setText("Wi‑Fi: " + b.interfaceName + "   •   " + b.localIp.getHostAddress() + "/" + b.prefixLength + "   •   Ubiquiti + Cambium + Mimosa");
                    if (!discovery.isRunning()) btnStart.setEnabled(true);
                }
            });
        }, "CPEFinder-NetworkRefresh").start();
    }

    private void startScan() {
        if (discovery.isRunning()) return;
        btnStart.setEnabled(false);
        btnStop.setEnabled(true);
        setRunningUi(true, "جارٍ الفحص");
        discovery.start();
    }

    private void stopScan() {
        discovery.stop();
        btnStop.setEnabled(false);
        btnStart.setEnabled(true);
        setRunningUi(false, "متوقف");
    }

    private void clearDevices() {
        synchronized (devices) { devices.clear(); }
        queueRender();
    }

    private void setRunningUi(boolean running, String text) {
        txtStatus.setText(text);
        txtStatus.setTextColor(Color.rgb(running ? 6 : 100, running ? 122 : 116, running ? 85 : 139));
        txtStatus.setBackgroundResource(running ? R.drawable.bg_badge_scan : R.drawable.bg_badge_idle);
    }

    @Override public void onDevice(Device incoming) {
        if (incoming == null) return;
        synchronized (devices) {
            incoming.mac = VendorDb.normalizeMac(incoming.mac);
            String key = locateKey(incoming);
            Device old = devices.get(key);
            if (old == null) {
                incoming.key = key;
                devices.put(key, incoming.copy());
            } else mergeInto(old, incoming);
        }
        queueRender();
    }

    private String locateKey(Device d) {
        if (!d.mac.isEmpty()) {
            String target = "mac:" + d.mac;
            if (!devices.containsKey(target) && !d.ip.isEmpty()) {
                String oldIpKey = null;
                for (Map.Entry<String, Device> e : devices.entrySet()) if (d.ip.equals(e.getValue().ip)) { oldIpKey = e.getKey(); break; }
                if (oldIpKey != null && !oldIpKey.equals(target)) {
                    Device old = devices.remove(oldIpKey); if (old != null) devices.put(target, old);
                }
            }
            return target;
        }
        if (!d.ip.isEmpty()) {
            for (Map.Entry<String, Device> e : devices.entrySet()) if (d.ip.equals(e.getValue().ip)) return e.getKey();
            return "ip:" + d.ip;
        }
        return "unknown:" + System.nanoTime();
    }

    private void mergeInto(Device old, Device in) {
        old.ip = NetUtil.chooseIp(old.ip, in.ip);
        if (!in.mac.isEmpty()) old.mac = in.mac;
        if (!in.vendor.isEmpty()) old.vendor = in.vendor;
        if (!in.model.isEmpty()) old.model = in.model;
        if (!in.hostname.isEmpty()) old.hostname = in.hostname;
        if (!in.firmware.isEmpty()) old.firmware = in.firmware;
        if (!in.ssid.isEmpty()) old.ssid = in.ssid;
        if (!in.method.isEmpty()) old.method = in.method;
        if (!in.description.isEmpty()) old.description = in.description;
        old.lastSeen = Math.max(old.lastSeen, in.lastSeen);
    }

    private void queueRender() {
        if (!renderQueued.compareAndSet(false, true)) return;
        main.postDelayed(() -> { renderQueued.set(false); renderDevices(); }, 120);
    }

    @Override public void onState(boolean running, String message) {
        main.post(() -> {
            setRunningUi(running, message);
            btnStop.setEnabled(running);
            btnStart.setEnabled(!running);
        });
    }

    private void renderDevices() {
        List<Device> out = new ArrayList<>();
        synchronized (devices) { for (Device d : devices.values()) out.add(d.copy()); }
        Collections.sort(out, Comparator.comparing((Device d) -> d.vendor).thenComparing(d -> d.ip));
        adapter.replace(out); txtCount.setText(String.valueOf(out.size()));
    }

    private void openDevice(Device d) {
        if (d == null || d.ip.isEmpty()) return;
        try { startActivity(new Intent(Intent.ACTION_VIEW, Uri.parse("http://" + d.ip))); }
        catch (Exception e) { Toast.makeText(this, "تعذر فتح عنوان الجهاز", Toast.LENGTH_SHORT).show(); }
    }

    private void copyIp(Device d) {
        if (d == null || d.ip.isEmpty()) return;
        ClipboardManager cb = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        if (cb != null) cb.setPrimaryClip(ClipData.newPlainText("IP", d.ip));
        Toast.makeText(this, "تم نسخ " + d.ip, Toast.LENGTH_SHORT).show();
    }

    @Override protected void onDestroy() { discovery.shutdown(); super.onDestroy(); }
}
