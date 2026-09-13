package com.mtradius.cpefinder;

import android.content.Context;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.BaseAdapter;
import android.widget.TextView;

import java.util.ArrayList;
import java.util.List;

final class DeviceAdapter extends BaseAdapter {
    private final LayoutInflater inflater;
    private final List<Device> items = new ArrayList<>();

    DeviceAdapter(Context context) { inflater = LayoutInflater.from(context); }

    void replace(List<Device> devices) {
        items.clear();
        items.addAll(devices);
        notifyDataSetChanged();
    }

    Device get(int position) { return items.get(position); }

    @Override public int getCount() { return items.size(); }
    @Override public Object getItem(int position) { return items.get(position); }
    @Override public long getItemId(int position) { return position; }

    @Override public View getView(int position, View convertView, ViewGroup parent) {
        Holder h;
        if (convertView == null) {
            convertView = inflater.inflate(R.layout.item_device, parent, false);
            h = new Holder();
            h.ip = convertView.findViewById(R.id.txtIp);
            h.vendor = convertView.findViewById(R.id.txtVendor);
            h.name = convertView.findViewById(R.id.txtName);
            h.details = convertView.findViewById(R.id.txtDetails);
            h.firmware = convertView.findViewById(R.id.txtFirmware);
            convertView.setTag(h);
        } else h = (Holder) convertView.getTag();

        Device d = items.get(position);
        h.ip.setText(d.ip.isEmpty() ? "—" : d.ip);
        h.vendor.setText(d.vendor.isEmpty() ? "جهاز" : d.vendor);
        String name = !d.hostname.isEmpty() ? d.hostname : (!d.model.isEmpty() ? d.model : "جهاز مكتشف");
        h.name.setText(name);

        StringBuilder det = new StringBuilder();
        if (!d.model.isEmpty() && !d.model.equals(name)) det.append("الطراز: ").append(d.model);
        if (!d.ssid.isEmpty()) { if (det.length() > 0) det.append("   •   "); det.append("SSID: ").append(d.ssid); }
        if (!d.mac.isEmpty()) { if (det.length() > 0) det.append("\n"); det.append("MAC: ").append(d.mac); }
        if (!d.description.isEmpty()) { if (det.length() > 0) det.append("\n"); det.append(d.description); }
        h.details.setText(det.length() == 0 ? "" : det.toString());

        StringBuilder fw = new StringBuilder();
        if (!d.firmware.isEmpty()) fw.append("الإصدار: ").append(d.firmware);
        if (!d.method.isEmpty()) { if (fw.length() > 0) fw.append("   •   "); fw.append(d.method); }
        h.firmware.setText(fw.toString());
        return convertView;
    }

    static final class Holder {
        TextView ip, vendor, name, details, firmware;
    }
}
