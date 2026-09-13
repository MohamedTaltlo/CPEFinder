# CPEFinder Android V2

السورس الكامل موجود في [`android-src/`](../android-src/)، ويتم بناؤه تلقائياً عبر GitHub Actions.

محركات الاكتشاف:

- Ubiquiti Discovery UDP/10001 لـ airMAX / airMAX AC وWave عندما يكون Discovery متاحاً.
- Cambium MNDP UDP/5678 عندما يكون MAC-Telnet/MNDP مفعلاً.
- Neighbor/ARP + OUI لـ Ubiquiti وMimosa وCambium.
- SSDP / WS-Discovery / mDNS.
- Active subnet/default-address probing.

Android بدون Root لا يسمح بـraw Ethernet ARP/LLDP، لذلك Mimosa بعنوان ثابت خارج subnet الهاتف لا يمكن ضمان اكتشافها إذا لم تعلن نفسها ببروتوكول IP محلي.
