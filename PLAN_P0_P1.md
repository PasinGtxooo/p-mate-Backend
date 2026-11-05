# แผนงาน P0–P1 — AI Project & Resource Planner

> เอกสารนี้สำหรับ **Claude Code** ทำต่อ (คนเขียน spec ไม่แก้ไฟล์เอง)
> Repo: `smart-task-ai` (backend Go) + `smart-task-ai-fe-prod` (frontend Next.js)

---

## บริบท (อ่านก่อนเริ่ม)

โฟลเดอร์ `C:\Users\PasinChearasirisin\OneDrive\HACKaton\`
- **`smart-task-ai`** — backend Go (clean architecture, Postgres/GORM, Fiber). **ไม่ใช่ git repo บนเครื่อง**
- **`smart-task-ai-fe-prod`** — frontend Next.js (git repo, remote `xssxx`, branch `feat/manday-skill-pm-matching` มี commit ค้าง)

### ทำไปแล้ว — อย่าทำซ้ำ
- Task: field `estimate_manday`, `required_skill`, `assignee_id`
- Profile: field `skills` (JSONB array `{name,level}`)
- Role: `pm` เพิ่มแล้ว (owner / pm / member) + helper `CanManageProject()`
- API ใหม่ที่มีแล้ว:
  - `GET  /api/tasks/:taskId/suggest-teammates`
  - `GET  /api/projects/:projectId/manpower`
  - `GET  /api/projects/:projectId/member-skills`
  - `PUT  /api/projects/:projectId/members/:accountId/skills`
  - `PUT  /api/projects/:projectId/members/:accountId/role`
- Frontend: ช่อง manday+skill ใน `CreateTaskModal`, `SuggestTeammatePanel` ใน `TaskDetailModal`, ปุ่มเลื่อน PM ใน `ManageMembersModal`, Calendar แยกสี/กรองตามคน (`assignee-utils.ts`, `AssigneeNameContext.tsx`)
- Migration `0001`, `0002`; `docker-compose.yml`; `scripts/seed/main.go`; `.env` ชี้ DB local

### ข้อควรรู้
ยังไม่เคย `go build` / `pnpm build` จริง (ตอนแก้ไม่มี Go/network) → **ต้อง build ก่อน แล้วไล่แก้ error**

---

## P0 — ทำให้รันได้จริง + verify + ERP demo data

### P0.1 Build & smoke test (ทำก่อนทุกอย่าง)

```powershell
# Postgres local: docker compose up -d  (มี docker-compose.yml)  หรือ Postgres native @ localhost:5432 user/pass = postgres
cd smart-task-ai
go build ./...          # ต้องผ่าน 0 error
go run ./scripts/seed   # สร้าง schema + seed
go vet ./...
cd ../smart-task-ai-fe-prod
pnpm install && pnpm build
```

จุดเสี่ยง compile ที่ต้องเช็ก (แก้ด้วย script มา ไม่เคย build):
- **Backend:** `internal/application/suggestion/{service.go,dto.go,usecase/*}`, `internal/infrastructure/rest/suggestion_handler.go`, `internal/interfaces/http/routes/private_routes.go`, `internal/domain/projects/entity/role.go`, `internal/mocks/{profile_repository.go,project_repository.go}` — เช็ก interface method ครบ, import ไม่ขาด/เกิน
- **Frontend:** `src/components/SuggestTeammatePanel.tsx`, `src/lib/assignee-utils.ts`, `src/contexts/AssigneeNameContext.tsx`, `src/app/app/calendar/page.tsx`, `src/components/ManageMembersModal.tsx`

**เสร็จเมื่อ:** build ทั้งคู่ผ่าน, login `alice`/`password123`, สร้าง task + suggest teammate + calendar แยกสี ทำงานไม่ error

### P0.2 เปลี่ยน seed เป็นธีม ERP/BC
ไฟล์: `smart-task-ai/scripts/seed/main.go` (enum มี `pm` แล้ว)

**Users + profiles** (skills = JSONB `{name,level}`):

| ชื่อ | role | position | skills |
|---|---|---|---|
| ปกรณ์ | owner | PM | Project Management:5, BC Project:4 |
| มานี | member | Consultant | BC Finance:5, Blueprint:5, Fit-Gap:4 |
| สมชาย | member | Consultant | BC Supply Chain:5, BC Manufacturing:4 |
| วิชัย | member | Developer | AL Development:5, Integration:4, Extension:4 |
| นภา | member | Developer | RDLC Report:5, Power Platform:4 |

password ทุกคน = `password123`

**Project 1:** `BC Implementation – ไทยเบฟ โลจิสติกส์` owner=ปกรณ์, go-live ≈ อีก 3 สัปดาห์
Members: ปกรณ์ owner, ที่เหลือ member

**Tasks** (assignee + required_skill + estimate_manday + start/end ตาม phase):

| Task | phase | assignee | required_skill | MD | status |
|---|---|---|---|---|---|
| Kick-off meeting | Plan | ปกรณ์ | Project Management | 1 | done |
| เก็บ Requirement / Blueprint | Analysis | มานี | BC Finance | 5 | in_progress |
| Fit-Gap Analysis | Analysis | มานี | Fit-Gap | 3 | todo |
| Config BC Finance | Develop | มานี | BC Finance | 4 | todo |
| Config Supply Chain | Develop | สมชาย | BC Supply Chain | 4 | todo |
| พัฒนา AL Custom Invoice | Develop | วิชัย | AL Development | 6 | todo |
| สร้าง Integration API | Develop | วิชัย | Integration | 4 | todo |
| พัฒนา RDLC Report | Develop | นภา | RDLC Report | 2 | todo |
| Migrate Master Data | Deploy | สมชาย | — | 3 | todo |
| UAT กับลูกค้า | Deploy | มานี | — | 3 | todo |
| Go-live Support (standby) | Go-live | ปกรณ์ | — | 2 | todo |

> วิชัยรวม 6+4 = **10 MD เกิน 8** → โชว์ Workload สีแดง (manpower)

**Project 2:** `BC Phase 2 – ไทยเบฟ` go-live **วันเดียวกับ Project 1** → ตั้ง narrative Go-live Conflict (ยังไม่ทำ alert)

> ถ้ายังไม่เพิ่มคอลัมน์ `position`: ใส่ใน `schemaSQL` ของ seed ด้วย
> `ALTER TABLE profiles ADD COLUMN IF NOT EXISTS position VARCHAR(50);`

**เสร็จเมื่อ:** เปิดแอปเห็นทีม ERP, งาน AL → suggest แนะนำวิชัย, manpower เห็นวิชัยแดง

---

## P1 — assign จากหน้า chat + แยก Consultant/Dev

### P1.1 Frontend API: `listMemberSkills` (ถ้ายังไม่มี)
ไฟล์: `fe/src/services/api.ts` (backend endpoint มีแล้ว)

```ts
export interface MemberSkill { name: string; level: number; }
export interface MemberWithSkills {
  account_id: string; name: string; role: string; skills: MemberSkill[];
}
export const listMemberSkills = (projectId: string) =>
  apiClient.get<{ success: boolean; message: string;
    data: { members: MemberWithSkills[] }; error: unknown;
  }>(`/api/projects/${projectId}/member-skills`);
```

### P1.2 แยก Consultant vs Developer (`position`)
- **Backend:** `profiles/entity/profile.go` เพิ่ม `Position *string gorm:"column:position"` → map ใน `profiles/domain.go` + `application/profile/dto.go` (+ schema seed ถ้ายังไม่ได้ทำ)
- **Frontend:** โชว์ `position` เป็น badge บน member row (ManageMembers) + ชื่อใน calendar / suggest panel

**เสร็จเมื่อ:** เห็นป้าย PM / Consultant / Developer แยกชัด

### P1.3 AI เสนอ skill + manday (backend)
- `internal/domain/chats/service/instruction.txt` — เพิ่มใน schema ของแต่ละ task:
  ```
  "required_skill": "<string, optional, สกิลที่งานต้องใช้ เช่น BC Finance, AL Development>",
  "estimate_manday": <number, optional, man-day>
  ```
  + อัปเดต example ให้มี 2 field
- ให้ไหลครบสาย: `chats/service` struct `TaskFromAI` → `send_message.go` `mapTasksToDTO` → `chat/dto.go` `TaskDTO` → FE `types/chat.ts` `BackendTask`

### P1.4 หน้า chat — โชว์ทุกคน + มอบงานได้เลย
ไฟล์: `fe/src/app/app/[projectId]/chat/page.tsx`

1. โหลดสมาชิกตอนเปิดหน้า:
   ```ts
   const [members, setMembers] = useState<MemberWithSkills[]>([]);
   useEffect(() => {
     listMemberSkills(projectId).then(r => setMembers(r.data.data.members)).catch(()=>{});
   }, [projectId]);
   ```
2. ฟังก์ชันหาคนเหมาะสุด (logic เดียวกับ backend: skill level ÷ 5, คน skill ตรงขึ้นก่อน):
   ```ts
   const skillLv = (m, s) => m.skills.find(k => k.name.toLowerCase() === s?.toLowerCase())?.level ?? 0;
   const bestMatch = (ms, s) => (!s ? ms[0]?.account_id
     : [...ms].sort((a,b)=>skillLv(b,s)-skillLv(a,s))[0]?.account_id);
   ```
3. ใน `ProposedTaskCard` (~line 1240–1270) เพิ่ม dropdown assignee:
   - แสดงสมาชิก **ทุกคน**: `ชื่อ + position + skill badges` (ไม่ต้องพิมพ์หา)
   - default = `bestMatch(members, task.required_skill)` (เลือกอัตโนมัติ)
   - skill ตรง → ไฮไลต์เขียว
   - เก็บค่าที่เลือก state ต่อ task เช่น `assigneeByTask[task.id]`
4. `proposedTaskToRequest` (~line 884) + `handleAcceptTask` (~line 902) ใส่:
   ```ts
   ...(assigneeByTask[task.id] && { assignee_id: assigneeByTask[task.id] }),
   ...(task.estimate_manday && { estimate_manday: task.estimate_manday }),
   ...(task.required_skill && { required_skill: task.required_skill }),
   ```
   (backend `CreateTaskRequest` รองรับแล้ว)
5. `ProposedTask` type (`types/chat.ts`) เพิ่ม `assignee_id?`, `required_skill?`, `estimate_manday?`

**เสร็จเมื่อ:** พิมพ์ในแชต → AI เสนอ task (มี skill/manday) → การ์ดโชว์รายชื่อทุกคน + เลือกคนเหมาะสุดอัตโนมัติ → กด accept สร้าง task พร้อม assignee ทันที ไม่ต้องไปหน้าอื่น

---

## ลำดับทำ

```
P0.1 build/verify → P0.2 ERP seed
→ P1.1 listMemberSkills → P1.2 position → P1.3 AI skill/manday → P1.4 chat assign
```

ทำทีละขั้น → build + ทดสอบก่อนไปขั้นถัดไป
