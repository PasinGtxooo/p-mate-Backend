# 🚀 คู่มือ Deploy Backend ไป Railway

## ข้อดีของ Railway
- ✅ ฟรี $5/เดือน credit
- ✅ มี PostgreSQL ในตัว
- ✅ Auto deploy จาก Git
- ✅ ไม่ต้อง Docker บนเครื่อง
- ✅ Setup ง่าย

---

## 📋 ขั้นตอนการ Deploy

### 1. สร้าง Account Railway
1. ไปที่ [railway.app](https://railway.app)
2. คลิก **"Start a New Project"**
3. Login ด้วย GitHub

### 2. สร้าง Project
1. คลิก **"New Project"**
2. เลือก **"Deploy from GitHub repo"**
3. เลือก repository ที่มี Backend (smart-task-ai)
4. เลือก branch **main** หรือ **Main**

### 3. เพิ่ม PostgreSQL Database
1. ใน Project dashboard คลิก **"+ New"**
2. เลือก **"Database"** → **"Add PostgreSQL"**
3. Railway จะสร้าง database พร้อม connection string ให้อัตโนมัติ

### 4. ตั้งค่า Environment Variables

คลิกที่ **Backend service** → **Variables** → เพิ่มตัวแปรเหล่านี้:

#### Database (Railway จะ inject อัตโนมัติ)
```bash
# Railway จะตั้งค่าให้เอง ไม่ต้องใส่
DATABASE_URL=${{Postgres.DATABASE_URL}}
```

#### ตั้งค่าเองทีละตัว:
```bash
# Frontend URL (จะได้จาก Vercel หลัง deploy frontend)
FRONTEND_URL=https://your-app.vercel.app

# Database Connection (ใช้ตัวแปรจาก Railway Postgres)
DB_HOST=${{Postgres.PGHOST}}
DB_DATABASE=${{Postgres.PGDATABASE}}
DB_USERNAME=${{Postgres.PGUSER}}
DB_PASSWORD=${{Postgres.PGPASSWORD}}
DB_PORT=${{Postgres.PGPORT}}

# JWT Secret (สร้างแบบสุ่ม)
JWT_SECRET=your_random_secret_key_here_at_least_32_characters

# Supabase (ถ้าใช้ Supabase Storage)
SUPABASE_PUBLIC_URL=https://kcqpdjmkyhpbrkcohcmx.supabase.co
S3_ENDPOINT=https://kcqpdjmkyhpbrkcohcmx.storage.supabase.co/storage/v1/s3
S3_ACCESS_KEY=your_supabase_s3_access_key
S3_SECRET_KEY=your_supabase_s3_secret_key
S3_BUCKET=images
S3_REGION=ap-northeast-1

# Groq AI
GROQ_API_KEY=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
GROQ_API_URL=https://api.groq.com/openai/v1/chat/completions

# CORS (จะได้ URL จาก Vercel)
CORS_ALLOW_ORIGINS=https://your-app.vercel.app
```

### 5. ตั้งค่า Root Directory (ถ้าใช้ monorepo)
1. ไปที่ **Settings** → **Service**
2. หา **"Root Directory"**
3. ใส่: `smart-task-ai`
4. **Save**

### 6. Deploy!
1. Railway จะ auto deploy
2. รอประมาณ 3-5 นาที
3. จะได้ URL แบบ: `https://your-app.up.railway.app`

---

## 🔗 เชื่อมต่อกับ Frontend (Vercel)

หลังจาก Backend deploy เสร็จแล้ว:

1. Copy URL จาก Railway (เช่น `https://smart-task-api.up.railway.app`)
2. ไปที่ Vercel Project Settings → Environment Variables
3. เพิ่มหรือแก้ไข:
   ```
   NEXT_PUBLIC_API_BASE_URL=https://smart-task-api.up.railway.app
   ```
4. Redeploy Frontend

---

## ⚙️ ตั้งค่า CORS

ใน Railway Backend Environment Variables เพิ่ม:
```bash
CORS_ALLOW_ORIGINS=https://your-frontend.vercel.app,https://your-frontend-branch.vercel.app
```

**หมายเหตุ:** ใส่ทั้ง production และ preview URLs

---

## 🔍 Monitoring & Logs

### ดู Logs:
1. คลิกที่ Backend service
2. ไปที่ tab **"Deployments"**
3. คลิก deployment ล่าสุด
4. ดู **"View Logs"**

### ดู Metrics:
1. Tab **"Metrics"**
2. เห็น CPU, Memory, Network usage

---

## 💡 Tips & Troubleshooting

### Backend ไม่ขึ้น?
```bash
# ตรวจสอบ logs ว่ามี error อะไร
# มักจะเป็นเรื่อง:
- Database connection failed → เช็ค DB_* variables
- Port binding → Railway จะ inject PORT variable ให้ ต้องใช้ port นั้น
```

### ตั้งค่า Health Check
Railway จะเช็คที่ `/health` endpoint (ตาม railway.toml)
ถ้า API ไม่มี endpoint นี้ ต้องเพิ่มใน Go code:

```go
// ใน cmd/main.go หรือ router
app.Get("/health", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ok"})
})
```

### Database Migration
หลัง deploy ครั้งแรก อาจต้องรัน migration:

1. ไปที่ Railway Postgres service
2. คลิก **"Connect"** → copy connection string
3. รัน migration จากเครื่องของคุณ:
   ```bash
   # ใช้ psql หรือเครื่องมือ migration ที่ใช้อยู่
   psql <DATABASE_URL> -f migrations/schema.sql
   ```

หรือเพิ่ม migration ใน startup command

---

## 🎯 Custom Domain (Optional)

1. ไปที่ Service Settings → **Networking**
2. คลิก **"Generate Domain"** หรือ **"Custom Domain"**
3. เพิ่ม custom domain ของคุณ
4. ตั้งค่า DNS ตามที่ Railway บอก:
   - Type: `CNAME`
   - Name: `api` (หรืออะไรก็ได้)
   - Value: `<your-service>.up.railway.app`

---

## 💰 Cost & Usage

- **Free tier:** $5 credit/เดือน
- **Usage:**
  - ~$0.10/day สำหรับ small app
  - Postgres: รวมอยู่ใน credit
- **ดู usage:** Dashboard → **Usage**

---

## 📚 เอกสารเพิ่มเติม

- [Railway Docs](https://docs.railway.app)
- [Railway PostgreSQL Guide](https://docs.railway.app/databases/postgresql)
- [Environment Variables](https://docs.railway.app/develop/variables)

---

## ✅ Checklist

- [ ] Deploy Backend to Railway
- [ ] เพิ่ม PostgreSQL database
- [ ] ตั้งค่า Environment Variables ครบ
- [ ] Backend ขึ้นสำเร็จ (เช็คจาก logs)
- [ ] เพิ่ม Backend URL ไปที่ Vercel Frontend
- [ ] ตั้งค่า CORS ให้รองรับ Vercel domain
- [ ] Test API ด้วย Postman/Thunder Client
- [ ] เช็ค Database connection
- [ ] Run migration (ถ้ามี)

---

**เสร็จแล้ว! Backend และ Frontend จะทำงานร่วมกันได้** 🎉
