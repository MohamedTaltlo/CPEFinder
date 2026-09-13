package com.mtradius.cpefinder;

final class Device {
    String key = "";
    String ip = "";
    String mac = "";
    String vendor = "";
    String model = "";
    String hostname = "";
    String firmware = "";
    String ssid = "";
    String method = "";
    String description = "";
    long lastSeen = System.currentTimeMillis();

    Device copy() {
        Device d = new Device();
        d.key = key;
        d.ip = ip;
        d.mac = mac;
        d.vendor = vendor;
        d.model = model;
        d.hostname = hostname;
        d.firmware = firmware;
        d.ssid = ssid;
        d.method = method;
        d.description = description;
        d.lastSeen = lastSeen;
        return d;
    }
}
