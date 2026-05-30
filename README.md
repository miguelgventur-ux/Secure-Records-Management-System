# Secure Records Management System (SRMS)
**Organisation:** York City Medical Centre  
**Domain:** Patient Medical Records  
**Language:** Go 1.21 · SQLite · html/template

---

## Quick Start

```bash
# Requirements:
# 1. Go 1.21 or later: https://go.dev/dl/
# 2. TDM-GCC 64-bit (Windows only): https://jmeubank.github.io/tdm-gcc/

# Setup
go mod tidy
go mod vendor
set CGO_ENABLED=1 && go run .

# Visit http://localhost:8080
```

### Seed Credentials

| Role    | Username | Password    |
|---------|----------|-------------|
| Admin   | admin    | Admin@1234  |
| Patient | jsmith   | Pass@1234   |
| Patient | emilyr   | Pass@1234   |
| Patient | mbrown   | Pass@1234   |

---

## Security Features Implemented

### Required Features
| Feature | Implementation |
|---------|---------------|
| Secure session cookies | `HttpOnly + SameSite=Strict` flags on every `Set-Cookie` |
| CSRF protection | Per-session token in every POST form, validated server-side before processing |
| SQL injection prevention | All queries use `database/sql` prepared statements with `?` placeholders |
| XSS prevention | All output rendered via `html/template`, which context-sensitively auto-escapes values |
| Password hashing | `bcrypt` at default cost (10 rounds); constant-time comparison via `CompareHashAndPassword` |
| Input validation | Length and regex checks on all user-supplied fields before DB writes |
| Audit trail | `last_updated_by` / `last_updated_at` columns updated atomically on every record write |
| Role-based access | Regular users: own record + low-risk fields only. Admins: full read/write on all records |

### Additional Feature 1 – Security Response Headers
Every HTTP response carries:
- **Content-Security-Policy**: `default-src 'self'` blocks inline scripts, external resources, and object embeds
- **X-Frame-Options: DENY** prevents clickjacking via iframe embedding  
- **X-Content-Type-Options: nosniff** stops MIME-type sniffing attacks  
- **Referrer-Policy: strict-origin-when-cross-origin** limits referrer leakage to third parties

Applied via a `securityHeadersMiddleware` that wraps the entire router.

### Additional Feature 2 – Account Lockout (Brute-Force Protection)
- After **5 consecutive failed login attempts** the account is locked for **15 minutes**
- Lockout state (`failed_attempts`, `locked_until`) is stored in the database — survives server restarts and is enforced regardless of the attacker's IP address
- A **dummy bcrypt comparison** is performed on unknown usernames to equalise response time and defeat user-enumeration via timing side-channels
- Successful login resets the counter immediately

---

## File Structure

```
srms/
├── main.go
├── database.go
├── handlers.go
├── middleware.go
├── models.go
├── session.go
├── validation.go
├── go.mod
├── go.sum
├── vendor/ # Vendored dependencies (golang.org/x/crypto, mattn/go-sqlite3)
└── templates/
    ├── login.html # Login css
    ├── record.html # Patient css
    ├── admin_records.html # Admin: list all records css
    └── admin_record.html # Admin: full record view and edit css
```
