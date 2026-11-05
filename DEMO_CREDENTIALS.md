# PMate — Demo Credentials & Mock Data

> ข้อมูล seed สำหรับ demo ของ PMate
> รัน `go run ./scripts/seed` เพื่อสร้างข้อมูลพวกนี้ใหม่ (idempotent — รันซ้ำได้)

---

## 🔑 Login Credentials

**Password ทุกคน:** `password123`

| Username | ชื่อ-นามสกุล | ชื่อเล่น | Email | Role | Position |
|---|---|---|---|---|---|
| **pakorn** | ปกรณ์ ธนากุล | โอ | pakorn@thaibev.co.th | owner | **PM** |
| **manee** | มานี ศรีสุข | นี | manee@thaibev.co.th | member | **Consultant** |
| **somchai** | สมชาย พานิช | ชาย | somchai@thaibev.co.th | member | **Consultant** |
| **wichai** | วิชัย ชัยวงศ์ | ชัย | wichai@thaibev.co.th | member | **Developer** |
| **napha** | นภา วงศ์วาน | ภา | napha@thaibev.co.th | member | **Developer** |

> 💡 Backend login รับเฉพาะ **username** (ไม่ใช่ email) และ **case-sensitive** → พิมพ์ `pakorn` ไม่ใช่ `Pakorn`

---

## 👥 Team Skill Matrix

| Member | Position | Skills (Lv 1-5) |
|---|---|---|
| ปกรณ์ | PM | Project Management: **5**, BC Project: 4 |
| มานี | Consultant | BC Finance: **5**, Blueprint: **5**, Fit-Gap: 4 |
| สมชาย | Consultant | BC Supply Chain: **5**, BC Manufacturing: 4 |
| วิชัย | Developer | AL Development: **5**, Integration: 4, Extension: 4 |
| นภา | Developer | RDLC Report: **5**, Power Platform: 4 |

---

## 📁 Projects

### 1. BC Implementation – ไทยเบฟ โลจิสติกส์
- **Owner:** ปกรณ์
- **Members:** ปกรณ์, มานี, สมชาย, วิชัย, นภา
- **Go-live:** +21 วันจากวัน seed
- **Narrative:** โปรเจกต์หลัก implement Business Central ให้ลูกค้าไทยเบฟ

### 2. BC Phase 2 – ไทยเบฟ
- **Owner:** ปกรณ์
- **Members:** เดียวกับ Project 1
- **Go-live:** วันเดียวกับ Project 1 (**จงใจตั้งให้ตรงกัน** เพื่อ demo "Go-live Conflict")

---

## 📝 Tasks — Project 1 (11 รายการ)

| # | Task | Phase | Assignee | Required Skill | MD | Status |
|---|---|---|---|---|---|---|
| 1 | Kick-off meeting | Plan (วัน 0-1) | ปกรณ์ | Project Management | 1 | **done** |
| 2 | เก็บ Requirement / Blueprint | Analysis (2-6) | มานี | BC Finance | 5 | **in_progress** |
| 3 | Fit-Gap Analysis | Analysis (6-8) | มานี | Fit-Gap | 3 | todo |
| 4 | Config BC Finance | Develop (9-12) | มานี | BC Finance | 4 | todo |
| 5 | Config Supply Chain | Develop (9-12) | สมชาย | BC Supply Chain | 4 | todo |
| 6 | ⚠ พัฒนา AL Custom Invoice | Develop (9-14) | วิชัย | AL Development | **6** | todo |
| 7 | ⚠ สร้าง Integration API | Develop (12-15) | วิชัย | Integration | **4** | todo |
| 8 | พัฒนา RDLC Report | Develop (11-13) | นภา | RDLC Report | 2 | todo |
| 9 | Migrate Master Data | Deploy (16-18) | สมชาย | — | 3 | todo |
| 10 | UAT กับลูกค้า | Deploy (18-20) | มานี | — | 3 | todo |
| 11 | Go-live Support (standby) | Go-live (21) | ปกรณ์ | — | 2 | todo |

### ⚠ Workload Overload — วิชัย

| | MD |
|---|---|
| Task #6 AL Custom Invoice | 6 |
| Task #7 Integration API | 4 |
| **รวม** | **10** |
| Deadline window (1 สัปดาห์) | 8 |
| **Over-capacity** | **+2 MD** 🔴 |

→ หน้า **Manpower** จะโชว์วิชัยสีแดง  
→ หน้า **Chat Suggest** จะติด badge "งานล้น" ที่วิชัย

---

## 🎬 Demo Scenarios

### 1️⃣ AI Suggest ตาม Skill + ความว่าง

```
pakorn login → BC Implementation → Chat
"สร้าง task แก้ bug AL invoice schema"
```
**คาด:** AI เสนอ task มี `required_skill=AL Development`, `estimate_manday=1-2`  
→ dropdown default เลือก **วิชัย** (skill Lv5) แต่ติด "งานล้น"  
→ Suggest popup แสดงว่าถ้า assign วิชัยจะเกิน capacity

### 2️⃣ AI Suggest Consultant

```
Chat: "วางแผน workshop เก็บ requirement ฝั่ง supply chain"
```
**คาด:** Suggest แนะนำ **สมชาย** (BC Supply Chain Lv5)

### 3️⃣ Report Suggest

```
Chat: "สร้าง report ยอดขายรายวัน"
```
**คาด:** Suggest แนะนำ **นภา** (RDLC Report Lv5)

### 4️⃣ Admin Create User

```
pakorn login → เปิด URL /admin
สร้าง user ใหม่ + add เข้าโปรเจกต์ BC + role = pm
```

---

## 🔄 Reset / Re-seed

### ล้างข้อมูลเก่าก่อน re-seed
ใน Supabase SQL Editor:
```sql
TRUNCATE tasks, project_members, project_invitations, projects, profiles, accounts CASCADE;
```

### รัน seed
```powershell
cd C:\Users\PasinChearasirisin\OneDrive\HACKaton\smart-task-ai
go run ./scripts/seed
```

---

## 🌐 URLs

| Service | URL |
|---|---|
| Frontend (Vercel) | https://p-mate-eight.vercel.app |
| Backend (Railway) | https://p-mate-backend-production.up.railway.app |
| Supabase Dashboard | https://supabase.com/dashboard/project/kcqpdjmkyhpbrkcohcmx |

---

## 🧪 Local Development

```powershell
# Terminal 1 — backend
cd C:\Users\PasinChearasirisin\OneDrive\HACKaton\smart-task-ai
go run cmd/main.go
# → http://localhost:8080

# Terminal 2 — frontend
cd C:\Users\PasinChearasirisin\OneDrive\HACKaton\smart-task-ai-fe-prod
pnpm dev
# → http://localhost:3000
```

หรือ double-click `run-all.bat` จากโฟลเดอร์ `HACKaton\`
