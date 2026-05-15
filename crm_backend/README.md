# MDIS CRM Backend 

A robust, high-performance CRM backend designed specifically for educational institutions (MDIS). This system streamlines the recruitment process from initial lead generation to signed contracts, integrating multiple communication channels and providing powerful analytics.

---

## Key Features

*   **Lead & Application Management**: Automates the lifecycle of student recruitment with custom pipeline stages.
*   **Automatic Deduplication**: Intelligently prevents data duplication by syncing leads with existing contacts.
*   **Omnichannel Integration**: 
    *   **Telegram Bot Webhooks**: Real-time student inquiries from the app.
    *   **Telephony/VoIP Support**: Automatic call logging and lead creation from incoming calls.
*   **Automated Notifications**: Instant Welcome Emails (SMTP) and SMS notifications for new students.
*   **Powerful Exports**: 
    *   **Excel (XLSX)**: Complete lead lists for administrative needs.
    *   **PDF (Landscape)**: Professional, detailed reports ready for printing.
*   **Manager Dashboard & Analytics**:
    *   **KPI Tracking**: Monitor manager activity (calls, closed deals, revenue).
    *   **Conversion Reports**: Track movement from "Lead Request" to "Signed Contract".
*   **Secure RBAC**: Role-Based Access Control for Admins and Admissions teams.
*   **Interactive API**: Fully documented with **Swagger UI**.

---

## Tech Stack

*   **Language**: Go 1.21+ (Chi Router)
*   **Database**: PostgreSQL 15
*   **Authentication**: JWT (JSON Web Tokens)
*   **Containerization**: Docker & Docker Compose
*   **Exports**: `excelize/v2` and `maroto/v2` (PDF)
*   **Documentation**: Swagger / OpenAPI 2.0

---

## Quick Start

### 1. Configure Secrets
Create a `.env` file or update `docker-compose.yml` with your settings:
```yaml
JWT_SECRET=your_secret_key
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=your_user
SMTP_PASS=your_pass
```

### 2. Launch with Docker
```bash
docker-compose up -d --build
```

### 3. Access the API
*   **Base URL**: `http://localhost:8080/api/v1`
*   **Swagger Docs**: `http://localhost:8080/api/v1/swagger/index.html`

---

## Default Credentials (Admin)
*   **Email**: `admin@admin.com`
*   **Password**: `Admin123!`

---

## Notification Testing
This project integrates with **Mailtrap** for safe email testing. To see the automated "Welcome" emails, simply create a new lead via the `/leads` endpoint and check your Mailtrap sandbox inbox.

---

## License
Custom Project for MDIS Educational Institution.
