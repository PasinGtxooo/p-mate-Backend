P2–P3 Detailed Spec (code-ready) — AI Project & Resource Planner

เอกสารนี้ลงรายละเอียดระดับโค้ดของ P2–P3 ให้ Claude Code วางตามได้เลย ต้องผ่าน P0 (build + ERP seed) และ P1 (chat assign + position) ก่อน pattern อ้างจากของเดิม: application/suggestion/{dto.go,service.go}, infrastructure/rest/suggestion_handler.go, interfaces/http/routes/private_routes.go, FE services/api.ts

P2.1 — Manpower Dashboard
Backend: เพิ่ม load รายคนใน ManpowerResponse

internal/application/suggestion/dto.go — เพิ่ม struct + field:

go
// MemberLoad is one member's workload for the manpower view.
type MemberLoad struct {
	AccountID  string  `json:"account_id"`
	Name       string  `json:"name"`
	AssignedMD float64 `json:"assigned_md"` // manday on open tasks
	FreeMD     float64 `json:"free_md"`     // deadlineWindowMD - assigned
	Over       bool    `json:"over"`        // assigned > capacity
}

// เพิ่มใน ManpowerResponse:
//   Members []MemberLoad `json:"members"`

internal/application/suggestion/service.go — ใน GetManpower(...) หลังรวม open tasks ต่อ assignee แล้ว เติม loop สร้าง members:

go
const deadlineWindowMD = 8.0 // มีอยู่แล้ว

members := make([]MemberLoad, 0, len(memberProfiles))
for _, p := range memberProfiles {
	assigned := assignedByAccount[p.AccountID] // map[string]float64 ที่รวม estimate_manday ของ task open ต่อคน (มี logic เดิมใน suggest อยู่แล้ว ใช้ซ้ำได้)
	free := deadlineWindowMD - assigned
	members = append(members, MemberLoad{
		AccountID:  p.AccountID.String(),
		Name:       p.DisplayName, // ใช้ field ชื่อจริงใน profile entity
		AssignedMD: assigned,
		FreeMD:     free,
		Over:       assigned > deadlineWindowMD,
	})
}
resp.Members = members

endpoint GET /api/projects/:projectId/manpower + handler มีอยู่แล้ว — แค่ response มี field เพิ่ม ไม่ต้องแก้ route

Frontend: service + component

fe/src/services/api.ts — เพิ่ม type ให้ตรง:

ts
export interface MemberLoad {
  account_id: string; name: string;
  assigned_md: number; free_md: number; over: boolean;
}
export interface ManpowerResponse {
  project_id: string; total_open_md: number; team_free_md: number;
  gap_md: number; need_more_people: boolean; suggest_headcount: number;
  members: MemberLoad[];
}
export const getManpower = (projectId: string) =>
  apiClient.get<{ success: boolean; data: ManpowerResponse }>(
    `/api/projects/${projectId}/manpower`,
  );

ไฟล์ใหม่ fe/src/components/ManpowerDashboard.tsx:

tsx
"use client";
import { useEffect, useState } from "react";
import { getManpower, ManpowerResponse, MemberLoad } from "@/services/api";
import { cn } from "@/lib/utils";
import { assigneeColor } from "@/lib/assignee-utils";

const CAP = 8;
function loadColor(m: MemberLoad) {
  if (m.assigned_md > CAP) return "bg-rose-500";
  if (m.assigned_md >= 6) return "bg-amber-500";
  return "bg-emerald-500";
}

export default function ManpowerDashboard({ projectId }: { projectId: string }) {
  const [data, setData] = useState<ManpowerResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getManpower(projectId)
      .then((r) => setData(r.data.data))
      .catch(() => setData(null))
      .finally(() => setLoading(false));
  }, [projectId]);

  if (loading) return <p className="text-sm text-gray-500">กำลังโหลด…</p>;
  if (!data) return <p className="text-sm text-rose-600">โหลดข้อมูลไม่สำเร็จ</p>;

  return (
    <div className="space-y-4">
      {/* team summary */}
      <div className="flex flex-wrap gap-4 rounded-xl border bg-white p-4 text-sm">
        <Stat label="งานค้างรวม" value={`${data.total_open_md} MD`} />
        <Stat label="คิวว่างทีม" value={`${data.team_free_md} MD`} />
        <Stat label="ขาด" value={`${data.gap_md} MD`}
          tone={data.gap_md > 0 ? "rose" : "emerald"} />
        {data.need_more_people && (
          <span className="rounded-full bg-rose-100 px-3 py-1 text-rose-700">
            ต้องเพิ่มคน ~{data.suggest_headcount} คน
          </span>
        )}
      </div>

      {/* per member */}
      <div className="space-y-2">
        {data.members.map((m) => (
          <div key={m.account_id} className="rounded-lg border bg-white px-3 py-2">
            <div className="flex items-center justify-between text-sm">
              <span className="flex items-center gap-2">
                <span className={cn("h-2 w-2 rounded-full", assigneeColor(m.account_id).dot)} />
                {m.name}
              </span>
              <span className={m.over ? "text-rose-600 font-medium" : "text-gray-500"}>
                {m.assigned_md}/{CAP} MD{m.over ? " ⚠️ ล้น" : ""}
              </span>
            </div>
            <div className="mt-1 h-2 overflow-hidden rounded-full bg-gray-100">
              <div className={cn("h-full rounded-full", loadColor(m))}
                style={{ width: `${Math.min(100, (m.assigned_md / CAP) * 100)}%` }} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function Stat({ label, value, tone = "gray" }:
  { label: string; value: string; tone?: "gray" | "rose" | "emerald" }) {
  const c = tone === "rose" ? "text-rose-600" : tone === "emerald" ? "text-emerald-600" : "text-gray-900";
  return (
    <div>
      <div className="text-xs text-gray-500">{label}</div>
      <div className={cn("text-lg font-semibold", c)}>{value}</div>
    </div>
  );
}

วางเป็น tab/section ในหน้าโปรเจกต์ (เช่นข้าง ๆ board) ให้เฉพาะ owner/pm เห็น

เสร็จเมื่อ: PM เปิดเห็นแถบ workload ทุกคน + วิชัยขึ้นแดง (10/8) + สรุปทีมพอ/ขาด

P2.2 — Go-live Conflict Alert
Backend: field วันที่

internal/domain/projects/entity/project.go:

go
GoLiveDate *time.Time `gorm:"column:go_live_date"`

map ต่อใน projects/domain.go ↔ application/project/dto.go (go_live_date *string ISO) และใส่ใน seed ทั้ง 2 โปรเจกต์ (ห่างกัน ≤7 วัน), ALTER TABLE projects ADD COLUMN IF NOT EXISTS go_live_date TIMESTAMP;

Frontend: ตรวจ client-side + banner

ไฟล์ใหม่ fe/src/lib/go-live-conflict.ts:

ts
import { Project } from "@/services/api";

export interface GoLiveConflict {
  a: Project; b: Project; daysApart: number; sharedMembers: string[];
}

// ต้องมีรายชื่อ member ต่อโปรเจกต์ (projectId -> accountId[])
export function findGoLiveConflicts(
  projects: Project[],
  membersByProject: Record<string, string[]>,
  windowDays = 7,
): GoLiveConflict[] {
  const out: GoLiveConflict[] = [];
  const withDate = projects.filter((p) => p.go_live_date);
  for (let i = 0; i < withDate.length; i++) {
    for (let j = i + 1; j < withDate.length; j++) {
      const a = withDate[i], b = withDate[j];
      const days = Math.abs(
        (+new Date(a.go_live_date!) - +new Date(b.go_live_date!)) / 86400000,
      );
      if (days > windowDays) continue;
      const setB = new Set(membersByProject[b.id] ?? []);
      const shared = (membersByProject[a.id] ?? []).filter((m) => setB.has(m));
      if (shared.length === 0) continue;
      out.push({ a, b, daysApart: Math.round(days), sharedMembers: shared });
    }
  }
  return out;
}

banner (หน้า dashboard/รายการโปรเจกต์) — เมื่อ conflicts.length > 0:

tsx
{conflicts.map((c, i) => (
  <div key={i} className="rounded-lg border border-rose-300 bg-rose-50 px-4 py-3 text-sm text-rose-800">
    ⚠️ Go-live ชนกัน: <b>{c.a.name}</b> และ <b>{c.b.name}</b> ห่างกัน {c.daysApart} วัน
    — คนที่ทำซ้ำ: {c.sharedMembers.map((id) => nameMap[id] ?? id).join(", ")}
  </div>
))}

ต้องแปลง sharedMembers (accountId) เป็นชื่อด้วย nameMap (buildNameMap จาก assignee-utils มีแล้ว)

เสร็จเมื่อ: 2 โปรเจกต์ go-live ใกล้กัน + คนซ้ำ → เด้ง banner แดงบอกชื่อคนที่โหลดชน

P2.3 — Customer role (read-only)

ใช้ position (P1.2) ค่า "Customer" — ไม่แตะ enum project_role

seed: เพิ่ม user ลูกค้า-ไทยเบฟ, role=member, position=Customer, เป็น member ของ Project 1
FE guard กลางที่เดียว — hook useMyProjectRole (หรือที่ที่รู้ว่าตัวเองเป็นใคร):
ts
const isCustomer = myPosition === "Customer";
const canManage = myRole === "owner" || myRole === "pm"; // ของเดิม
เมื่อ isCustomer:
หน้าโปรเจกต์ render เฉพาะ status board + วัน go-live (read-only)
ซ่อน: ปุ่ม create/edit task, assign, skill editor, ManpowerDashboard, SuggestTeammatePanel, ManageMembers
การ์ดงานแสดงแค่ชื่องาน + status + assignee (ไม่โชว์ manday/skill ภายใน)

อย่าพึ่ง guard ฝั่ง FE อย่างเดียวถ้ามีเวลา — backend ควรปฏิเสธ customer ที่เรียก write endpoint (เช็ก CanManageProject() ที่มีแล้วครอบ owner/pm; customer เป็น member ธรรมดาจึงถูกปฏิเสธ allocate/skill อยู่แล้ว ส่วน read ต้องกันเพิ่มถ้าต้องซ่อน manday)

เสร็จเมื่อ: login เป็นลูกค้า เห็นแค่ความคืบหน้า + go-live, ไม่เห็น/กดอะไรภายในไม่ได้

P3.1 — Auto Timesheet
Backend: endpoint ใหม่

internal/application/suggestion/dto.go (หรือแพ็กเกจ report แยก):

go
type TimesheetRow struct {
	AccountID string  `json:"account_id"`
	Name      string  `json:"name"`
	Week      string  `json:"week"`       // ISO "2026-W45"
	DoneMD    float64 `json:"done_md"`
	TaskCount int     `json:"task_count"`
}
type TimesheetResponse struct {
	ProjectID string         `json:"project_id"`
	Rows      []TimesheetRow `json:"rows"`
}

service: query tasks ที่ status = done ของโปรเจกต์ → group by (assignee_id, ISO week ของ end_date/updated_at) → sum estimate_manday, count:

go
func isoWeek(t time.Time) string {
	y, w := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}
// loop tasks done, key = accountID + "|" + isoWeek(task.EndDate or UpdatedAt)
// acc[key].DoneMD += manday; acc[key].TaskCount++

handler + route: GET /api/projects/:projectId/timesheet (pattern เดียวกับ GetManpower — gate ด้วย IsMember)

Frontend

fe/src/services/api.ts: getTimesheet(projectId) + types หน้า/tab Timesheet: ตาราง คน × สัปดาห์ = DoneMD, แถวรวมท้าย, ปุ่ม Export CSV:

ts
function toCSV(rows: TimesheetRow[]) {
  const head = "name,week,done_md,task_count\n";
  return head + rows.map(r => `${r.name},${r.week},${r.done_md},${r.task_count}`).join("\n");
}
// download: new Blob([csv],{type:"text/csv"}) -> URL.createObjectURL -> <a download>

เสร็จเมื่อ: ปิดงาน (done) แล้ว timesheet รวม MD ต่อคนต่อสัปดาห์ขึ้นเอง + export CSV ได้

P3.2 — AI Task Splitting (BC phase template)

ไฟล์: internal/domain/chats/service/instruction.txt

เพิ่มกติกา (ต่อจากของเดิม): เมื่อ user ระบุโมดูล/scope ของ BC ให้แตกงานตาม phase มาตรฐาน:
เมื่อผู้ใช้ขอ implement โมดูลหรือ scope ของ Business Central ให้แตกงานตามลำดับ phase มาตรฐาน:
Blueprint → Fit-Gap Analysis → Configuration → Development → Data Migration → UAT → Go-live
กำหนด required_skill และ estimate_manday ให้แต่ละงานตามประเภท:
- Blueprint / Fit-Gap / Configuration / UAT → สกิลสาย Consultant (เช่น BC Finance, BC Supply Chain) 2–5 MD
- Development / Integration / Report → สกิลสาย Developer (เช่น AL Development, Integration, RDLC Report) 2–6 MD
- Kick-off / Go-live Support → Project Management 1–2 MD
เพิ่ม field ในแต่ละ task ของ output schema (ถ้ายังไม่ได้ทำใน P1.3):
"required_skill": "<string, optional>",
"estimate_manday": <number, optional>
ใส่ few-shot 1 ชุดเต็ม เช่น input "implement BC Finance module" → output 6–7 งานเรียง phase พร้อม skill/MD ครบ (ให้โมเดลลอกรูปแบบ)

ปลายทางไหลเข้า ProposedTaskCard (P1.4) → เลือกคนอัตโนมัติทุกงานในชุด

เสร็จเมื่อ: พิมพ์ "ทำโมดูล BC Finance" → ได้ชุดงานตาม phase พร้อม skill/MD/คนแนะนำครบในคลิกเดียว

P3.3 — ขัดเงา demo / pitch
Export: ปุ่ม CSV ใน Manpower + Timesheet (ใช้ helper toCSV ด้านบน); PDF ถ้ามีเวลา (print view)
Demo script (README): login PM → แชต "ทำโมดูล BC Finance" แตกงาน → suggest เลือกคน → calendar แยกสี → manpower วิชัยแดง → go-live conflict alert → login customer read-only → timesheet export
i18n: เติม key ที่ขาดทั้ง th/en: manpower.*, timesheet.*, goLiveConflict.*, position.consultant/developer/customer
Empty/error states: manpower/timesheet/suggest ตอนไม่มีข้อมูล ขึ้นข้อความแทนหน้าว่าง
ชื่อแอป (มีคำว่า PM): เลือก 1 — PMate / PlanPM / PM Pilot / PMantis → ใส่ <title> + header + favicon

เสร็จเมื่อ: กด demo ตาม script จบลื่น ไม่มีหน้าพัง + มีไฟล์ export ให้ดูของจริง

ลำดับ & build gate
P2.1 → P2.2 → P2.3 → P3.1 → P3.2 → P3.3

ทุกขั้น: go build ./... + pnpm build ผ่านก่อนไปต่อ ถ้าเวลาน้อย เอาแค่ P2.1 (dashboard) + P2.2 (conflict) + P3.2 (AI splitting) = โชว์ได้ตรง pitch สุด