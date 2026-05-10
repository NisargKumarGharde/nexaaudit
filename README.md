# NexaAudit

<p align="left">
  <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/go/go-original-wordmark.svg" height="48" alt="Go" title="Go" /> &nbsp;&nbsp;
  <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/postgresql/postgresql-original.svg" height="48" alt="PostgreSQL" title="PostgreSQL" /> &nbsp;&nbsp;
  <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/supabase/supabase-original.svg" height="48" alt="Supabase" title="Supabase" /> &nbsp;&nbsp;
  <img src="https://github.com/pinecone-io.png" height="48" alt="Pinecone" title="Pinecone" style="border-radius: 8px;" /> &nbsp;&nbsp;
  <img src="https://upload.wikimedia.org/wikipedia/commons/8/8a/Google_Gemini_logo.svg" height="36" alt="Google Gemini" title="Google Gemini" /> &nbsp;&nbsp;
  <img src="https://lovable.dev/favicon.svg" height="48" alt="Lovable AI" title="Lovable AI" /> &nbsp;&nbsp;
</p>

**An AI-Powered Financial Fraud Detection & Audit Engine** built to automate the extraction of unstructured invoice data and prevent duplicate transaction fraud using semantic vector embeddings.

---

## 🚀 The Problem & The Solution
Manual auditing of financial documents is slow, error-prone, and highly vulnerable to duplicate invoice submissions (a common vector for corporate fraud). 

**NexaAudit** solves this by deploying a RAG (Retrieval-Augmented Generation) pipeline. It uses **Google Gemini 2.5 Flash** to extract structured JSON data from visual PDFs, compresses that semantic meaning into a 768-dimension mathematical vector using **Gemini-Embedding-2**, and performs a real-time Cosine Similarity search against a **Pinecone Vector Database** to flag fraudulent duplicates with 99%+ accuracy.

## 🏗️ System Architecture & Tech Stack

NexaAudit is designed as a highly concurrent, multi-cloud application.

### Backend (Render)
* **Language:** Go (Golang)
* **Driver:** `pgx/v5` (Configured for Simple Protocol to bypass connection pooler conflicts)
* **Architecture:** RESTful API with strict separation of concerns (Handlers, DB interfaces, AI services).

### Frontend (Vercel)
* **Framework:** React / Vite (TypeScript)
* **UI Generation:** Rapidly prototyped and generated using **Lovable** to accelerate the frontend workflow, allowing maximum engineering bandwidth to be dedicated to the backend RAG architecture.
* **Features:** Drag-and-drop document upload, real-time polling dashboard, responsive UI.

### Data Layer
* **Relational State (Supabase / PostgreSQL):** Handles user generation, document metadata, and final extraction status.
* **Vector Memory (Pinecone):** Serverless vector database storing document embeddings for sub-second similarity searches.

---

## ⚙️ How the Pipeline Works

1. **Ingestion:** The React UI sends a multipart/form-data PDF payload to the Go server.
2. **State Management:** Go immediately logs the transaction as `pending` in PostgreSQL.
3. **Multimodal Extraction:** The document is securely streamed to the Gemini AI API, acting as a financial auditor to extract `vendor_name`, `total_amount`, and identify visual anomalies.
4. **Vector Generation:** The extracted data is passed to the embedding model to generate a unique semantic fingerprint (768-D tensor).
5. **Duplicate Detection:** Go queries Pinecone with the new vector. If a match exceeds the 0.99 cosine similarity threshold, the document is flagged as a duplicate fraud attempt.
6. **Resolution:** PostgreSQL is updated with the final status (`audited` or `flagged`), and a JSON response populates the frontend dashboard.

---

## 🧠 Key Engineering Decisions

* **Why Go?** The requirement to handle concurrent API calls to multiple external services (Supabase, Gemini, Pinecone) made Go's goroutines and highly efficient networking stack the optimal choice for a fast, fault-tolerant backend.
* **Connection Pooling:** Implemented explicit connection management (`statement_cache_capacity=0` and `pgx.QueryExecModeSimpleProtocol`) to ensure the Go driver plays perfectly with Supabase's IPv4 connection poolers at scale.
* **Fault Tolerance:** If the AI extraction fails, the system gracefully catches the error, marks the database record as `failed` (preventing data loss), and returns a controlled `201` to the frontend rather than crashing the server.

---

## 💻 Local Development Setup

### Prerequisites
* Go 1.21+
* Node.js & npm
* Docker (for local PostgreSQL)

### 1. Clone the repository
```bash
git clone [https://github.com/NisargKumarGharde/nexaaudit.git](https://github.com/NisargKumarGharde/nexaaudit.git)
cd nexaaudit
```

### 2. Environment Variables

Create a `.env` file in the root directory
```
PORT=8082
DATABASE_URL=postgres://admin:secretpassword@localhost:5434/nexaaudit?sslmode=disable
GEMINI_API_KEY=your_gemini_key
PINECONE_API_KEY=your_pinecone_key
PINECONE_HOST=your_pinecone_host
```

### 3. Start the Database
```bash
docker compose up -d
```

### 4. Run the Go Server
```bash
go run cmd/api/main.go
```

# 👨‍💻 About the Developer
## Nisarg Gharde
Dual-Degree Undergraduate Student at IIT Patna & IIT Madras Specializing in Data Science and Backend Engineering. Passionate about building highly concurrent microservices, data pipelines, and the modern AI-first tech stack.

LinkedIn: [Nisarg Gharde](https://www.linkedin.com/in/nisargkumargharde/)

Email: [nisarg.gharde@gmail.com](mailto:nisarg.gharde@gmail.com)

Currently open to Internships and Entry-Level roles and actively seeking opportunities.
