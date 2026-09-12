# CPEFinder — كاشف عناوين أجهزة الشبكة

أداة Windows مفتوحة المصدر لاكتشاف أجهزة CPE المتصلة بالشبكة المحلية وعرض معلوماتها، مع واجهة عربية.

تهدف الأداة إلى مساعدة فنيي الشبكات في العثور على أجهزة مثل Ubiquiti airMAX / airMAX AC / Wave، مع دعم اكتشاف عام يفيد مع Mimosa وCambium والأجهزة الأخرى متى كانت تعلن عن نفسها عبر البروتوكولات المحلية المدعومة.

## الخصائص

- واجهة Windows عربية واتجاه RTL.
- اكتشاف Ubiquiti Discovery المحلي.
- قراءة معلومات الجهاز المتاحة مثل IPv4 وMAC والطراز واسم الجهاز وSSID وإصدار النظام متى أرسلها الجهاز.
- استخدام ARP وواجهات Windows المحلية للمساعدة في الاكتشاف.
- لا يرسل أي Telemetry ولا يحتاج خادماً سحابياً كي يعمل.
- لا يحتاج اتصال إنترنت أثناء تشغيل التطبيق.

> ملاحظة: إمكانات الاكتشاف تختلف باختلاف الشركة وإصدار Firmware وإعداد الجهاز. ليس كل جهاز يعلن عن كل الحقول.

## الأنظمة المدعومة

- Windows 10/11 x64.
- لا يحتاج Administrator في الإصدار الحالي.

## البناء من المصدر

يتطلب Go 1.23 أو أحدث:

```powershell
git clone https://github.com/MohamedTaltlo/CPEFinder.git
cd cpefinder
$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -trimpath -ldflags="-H=windowsgui -s -w -X main.appVersion=0.1.0" -o CPEFinder.exe .
```

أو استخدم GitHub Actions الموجود في `.github/workflows/build.yml`.

## الخصوصية

هذا البرنامج لا ينقل أي معلومات إلى أنظمة شبكية أخرى إلا عندما يطلب المستخدم صراحةً عملية مرتبطة بالاكتشاف المحلي أو فتح واجهة جهاز محدد. لا توجد Telemetry أو Analytics. راجع [PRIVACY.md](PRIVACY.md).

## Code signing policy

Free code signing provided by SignPath.io, certificate by SignPath Foundation **بعد قبول المشروع في برنامج SignPath Foundation**.

- جميع الملفات التنفيذية الرسمية يجب أن تُبنى آلياً من هذا المستودع العام.
- لا يتم توقيع ملفات تنفيذية مبنية من مصدر غير موجود في هذا المستودع.
- كل إصدار موقّع يتطلب موافقة يدوية وفق إعداد SignPath.
- Committers / reviewers / approvers ستُحدد روابطهم هنا بعد إنشاء مستودع GitHub والمنظمة/الحساب.

التفاصيل: [CODE_SIGNING_POLICY.md](CODE_SIGNING_POLICY.md).

## الأمان

الأداة مخصصة لاكتشاف الأجهزة التي تملكها أو مخولاً بإدارتها على الشبكة المحلية. لا تتضمن استغلال ثغرات أو تجاوز كلمات مرور أو آليات حماية.

للإبلاغ عن مشكلة أمنية راجع [SECURITY.md](SECURITY.md).

## الترخيص

MIT License — راجع [LICENSE](LICENSE).
