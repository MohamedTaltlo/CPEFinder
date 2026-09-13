# CPEFinder — كاشف عناوين أجهزة الشبكة

أداة مفتوحة المصدر لاكتشاف أجهزة CPE والراديو على الشبكة المحلية، مع نسختين Windows وAndroid.

## Windows v0.3.0

محركات الاكتشاف الحالية:

- Ubiquiti Discovery عبر UDP/10001: airMAX / airMAX AC والأجهزة التي تدعم نفس بروتوكول Ubiquiti، بما فيها Wave عندما يكون Discovery متاحاً في الـFirmware.
- Cambium MNDP عبر UDP/5678 عند تفعيل MAC-Telnet/MNDP Discovery.
- ARP table + OUI recognition لـ Ubiquiti وMimosa وCambium.
- SSDP كطريقة مساعدة.
- Active local-subnet probing وdefault-address probes لتغذية ARP table.

يعرض IPv4 وMAC والشركة والطراز واسم الجهاز وSSID وإصدار النظام عندما يرسل الجهاز هذه البيانات.

## Android V2

الحزمة موجودة في `android/CPEFinder_Android_Studio_V2_COMPLETE.zip` ويتم بناؤها تلقائياً بواسطة `.github/workflows/android.yml`.

Android يستخدم:

- Ubiquiti Discovery UDP/10001.
- Cambium MNDP UDP/5678.
- ARP/Neighbor + OUI recognition.
- SSDP / WS-Discovery / mDNS.
- Active subnet/default-address probing.

### ملاحظة Android المهمة

Android العادي بدون Root لا يسمح للتطبيقات بإرسال أو التقاط Ethernet ARP/LLDP frames خام. لذلك Ubiquiti/Cambium قد يُكتشفان حتى مع اختلاف IP subnet عندما يرد بروتوكول الـbroadcast الخاص بهما، أما Mimosa بعنوان ثابت خارج subnet الهاتف فلا توجد طريقة عامة مضمونة 100% إذا لم يعلن الجهاز عن نفسه ببروتوكول IP discovery. داخل نفس subnet يحاول التطبيق كشفه عبر Neighbor/OUI والفحص النشط.

## البناء من المصدر — Windows

يتطلب Go 1.23 أو أحدث:

```powershell
git clone https://github.com/MohamedTaltlo/CPEFinder.git
cd CPEFinder
$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -trimpath -ldflags="-H=windowsgui -s -w -X main.appVersion=0.3.0" -o CPEFinder.exe .
```

أو استخدم GitHub Actions.

## الخصوصية

لا توجد Telemetry أو Analytics. راجع [PRIVACY.md](PRIVACY.md).

## Code signing policy

Free code signing provided by SignPath.io, certificate by SignPath Foundation بعد قبول المشروع في برنامج SignPath Foundation.

التفاصيل: [CODE_SIGNING_POLICY.md](CODE_SIGNING_POLICY.md).

## الأمان

الأداة مخصصة لاكتشاف وإدارة الأجهزة التي تملكها أو مخولاً بإدارتها. لا تتضمن استغلال ثغرات أو تجاوز كلمات مرور أو آليات حماية.

## الترخيص

MIT License — راجع [LICENSE](LICENSE).
