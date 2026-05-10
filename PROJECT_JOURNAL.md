# 📖 Engineering Journal & Lessons Learned

**Project:** NexaAudit 

**Timeline:** May 3, 2026 – May 10, 2026

Building a multi-cloud, AI-native application from scratch taught me that the code is often the easy part; orchestrating the infrastructure is where the real engineering happens. Below is a log of the critical technical hurdles I faced, the mistakes I made, and the architectural solutions I implemented.

---

## 1. The Supabase Connection Pooling Trap

**The Problem**: I could connect to my local Docker PostgreSQL database perfectly, but my deployed Render backend kept failing to connect to Supabase.

**The Mistake**: I treated a cloud database exactly like a local database. I didn't realize that Supabase utilizes an IPv4 connection pooler (PgBouncer/Supavisor) to handle traffic. The standard Go `pgx/v5` driver attempts to use prepared statements by default, which conflicts with Supabase's transaction-mode connection poolers.

**The Fix**: I had to explicitly bypass prepared statements. I appended `?statement_cache_capacity=0` to my Database URL and forced the Go driver into `pgx.QueryExecModeSimpleProtocol`.

**The Lesson**: Always read the cloud provider's pooling documentation before writing the database driver logic.

---

## 2. The Silent AI Failure & "Loud Logging"

**The Problem**: My React frontend was successfully uploading PDFs, and the backend was returning a `201 Created` response, but the dashboard table was completely blank.

**The Mistake**: I had built my backend to be too fault-tolerant. When the Gemini AI extraction failed, my Go code caught the error, set the document status to `failed` to prevent data loss, and sent a success response to the frontend without any data. It was failing silently.

**The Fix**: I implemented a "Loud Logging" pattern, adding distinct console logs at every exact millisecond of the pipeline (Form parsing -> File reading -> DB Insert -> AI Extraction -> Pinecone check). This immediately exposed the hidden `gemini generation failed` error.

**The Lesson**: Fault tolerance is crucial, but failing silently is a nightmare for debugging. Errors must be caught, but they must also be loudly broadcasted to the server logs.

---

## 3. The GitHub Auto-Revoke Assassin (Secret Leaks)

**The Problem**: Even after fixing my AI code, Google kept throwing an `API_KEY_INVALID` error. I generated new keys, but they were expiring instantly.

**The Mistake**: During local testing, I hardcoded my Gemini API key into `test_models.go` and `internal/ai/gemini.go`. When I pushed my code to GitHub, Google’s automated security bots scanned the public repository, found the plaintext keys, and instantly revoked them to protect my account. Even though I set up Render environment variables, my Go code was still pointing to the dead, hardcoded strings.

**The Fix**: I permanently deleted the test files, updated `gemini.go` to strictly use `os.Getenv("GEMINI_API_KEY")`, pushed the clean code to GitHub to prove the leak was plugged, and finally generated a secure key.

**The Lesson**: Never, ever hardcode secrets, even for "quick local tests." Always use a `.env` file and immediately add it to `.gitignore`.

---

## 4. Cloud Region Geo-Blocking

**The Problem**: The pipeline was flawless, but Gemini suddenly threw a `User location is not supported` error.

**The Mistake**: When I deployed my Go server on Render, I selected the **Singapore** region because it was geographically closer to my Supabase database in Mumbai. However, Google's API blocked requests coming from that specific Asian data center IP pool.

**The Fix**: I destroyed the Render instance and re-architected the deployment, moving the backend server to **Oregon (US West)**. While this added a tiny fraction of network latency communicating with Mumbai, US-based servers are universally trusted by US-based AI models. The API request passed immediately.

**The Lesson**: When dealing with bleeding-edge AI APIs, backend infrastructure should generally be hosted in US regions to avoid unexpected geo-blocking or IP blacklisting.

---

## Final Thoughts

This week of debugging transformed this from a simple coding project into an exercise in System Design. I learned how to manage cross-cloud communication (Vercel -> Render -> Supabase -> Pinecone -> Google AI) and how to secure a production pipeline.
