# CPEFinder Android V2

تطبيق Android لاكتشاف أجهزة CPE على نفس شبكة Wi‑Fi المحلية. يبدأ متوقفاً، يقرأ اتصال Wi‑Fi تلقائياً، ويبدأ الفحص فقط عند الضغط على **بدء الفحص**.

## محركات الاكتشاف

- Ubiquiti Discovery V1/V2 عبر UDP 10001: airMAX القديم، airMAX AC، والأجهزة التي تدعم بروتوكول Ubiquiti/Wave discovery.
- Cambium ePMP/PTP/PMP عندما يكون MAC‑Telnet/MNDP مفعلاً: UDP 5678 مع قراءة IP وMAC والاسم والطراز والإصدار.
- Mimosa: اكتشاف من جدول ARP/Neighbor حسب OUI، مع فحص نشط داخل subnet الحالي لتعبئة جدول الجيران، ودعم عناوين الإدارة الافتراضية الشائعة.
- SSDP + WS‑Discovery + mDNS كاكتشاف عام إضافي للأجهزة التي تعلن خدماتها محلياً.
- تعريف Offline لـOUI الخاصة بـUbiquiti وMimosa وCambium/Airspan.

## ملاحظة مهمة لميموسا على Android

Android العادي لا يمنح التطبيقات صلاحية إرسال/التقاط Ethernet ARP أو LLDP الخام. لذلك Mimosa ذات IP ثابت خارج subnet الذي ضبطته على الهاتف لا يمكن ضمان اكتشافها من تطبيق غير Root إذا لم تعلن عن نفسها ببروتوكول IP محلي. داخل subnet الحالي، أو عندما يظهر الجهاز في Neighbor/ARP، يتم اكتشافه تلقائياً.

## البناء

افتح المجلد في Android Studio وانتظر أول Gradle Sync. أول بناء فقط يحتاج إنترنت لتنزيل Android Gradle Plugin/Gradle إذا لم يكونا موجودين في الكاش. التطبيق نفسه بعد التثبيت لا يحتاج إنترنت.

- Min SDK: 24
- Target/Compile SDK: 35
- Java: 17
- Version: 2.0.0
