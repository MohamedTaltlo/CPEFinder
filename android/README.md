# CPEFinder Android V2

حزمة Android Studio لتطبيق اكتشاف أجهزة CPE عبر شبكة Wi‑Fi المحلية.

يدعم محركات اكتشاف متعددة تعمل بالتوازي:

- Ubiquiti Discovery عبر UDP/10001 لـ airMAX / airMAX AC والأجهزة التي تدعم نفس بروتوكول Ubiquiti، بما فيها Wave عندما يكون Discovery متاحاً في الـFirmware.
- Cambium MNDP عبر UDP/5678 للأجهزة التي يكون MAC-Telnet/MNDP Discovery مفعلاً عليها.
- ARP/Neighbor table + OUI recognition لـ Ubiquiti وMimosa وCambium.
- SSDP وWS-Discovery وmDNS كطرق مساعدة.
- Active local-subnet probing وdefault-address probes لتغذية Neighbor table.

## Android limitation

Android العادي بدون Root لا يسمح للتطبيقات بإرسال/التقاط Ethernet ARP/LLDP frames خام. لذلك Ubiquiti/Cambium يمكن اكتشافهما خارج IP subnet عندما يرد بروتوكول الـbroadcast الخاص بهما، أما Mimosa بعنوان ثابت خارج subnet الهاتف فلا توجد طريقة عامة مضمونة 100% إذا لم يعلن الجهاز عن نفسه عبر IP discovery. داخل نفس subnet يحاول التطبيق اكتشافه عبر Neighbor/OUI والفحص النشط.

الحزمة الجاهزة: `CPEFinder_Android_Studio_V2_COMPLETE.zip`.
