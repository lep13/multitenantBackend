# Multitenant Application

A multitenant web application designed to manage cloud resources on AWS and GCP, with role-based access control for admins, managers, and users. The application is divided into two repositories: **Frontend (UI)** and **Backend (API)**.

## Features

- **Role-Based Access Control (RBAC):**
  - **Admin**: Create/Delete managers.
  - **Manager**: Create/Delete users, manage groups, monitor budgets and costs, and oversee cloud resources created or deleted by users.
  - **User**: Create/Delete resources on AWS and GCP, send notifications to managers for budget approvals, and generate alerts.
- **Cloud Resource Management**: AWS and GCP resources are manipulated using their respective CLI tools.
- **Notifications**: Automatic notifications sent to managers for resource creation, deletion, alerts, and budget requests.
- **Database**: MongoDB, with the URI securely fetched from AWS Secrets Manager.

---

## Repositories

1. **Frontend (UI)**

   - [Frontend Repository](https://github.com/lep13/multitenantUI)
   - Built with Angular and Node.js.
   - Runs on port `4200`, fetching data from the Node.js backend on port `5000` and the Golang API on port `8080`.

2. **Backend (API)**

   - [Backend Repository](https://github.com/lep13/multitenantBackend)
   - Built with Golang, exposing REST APIs on port `8080`.
   - MongoDB is used for data storage, with credentials managed via AWS Secrets Manager.

---

## Setup Instructions

### Frontend

1. Clone the frontend repository:
   ```bash
   git clone https://github.com/lep13/multitenantUI
   cd multitenantUI
   npm install
   ng serve

   cd backend
   node server.js
   ```
2. Ensure the Node.js backend runs on port `5000` and the Golang API on port `8080`.

### Backend

1. Clone the backend repository:
   ```bash
   git clone https://github.com/lep13/multitenantBackend
   cd multitenantBackend
   go run main.go
   ```
2. Ensure MongoDB URI is set in AWS Secrets Manager and accessible.

---

## Cloud Resource Management

- AWS and GCP resources are manipulated on the respective accounts using the AWS CLI and GCP CLI tools. 
- Ensure AWS and Google CLI are configured
- Users perform operations like creation, deletion, and updates, while managers monitor and approve budget requests or handle alerts.

---

## Technologies Used

- **Frontend**: Angular, Node.js
- **Backend**: Golang
- **Database**: MongoDB
- **Cloud**: AWS, GCP
- **Others**: Docker, REST APIs, AWS Secrets Manager

---

For more details, refer to the [Frontend Repo](https://github.com/lep13/multitenantUI) and [Backend Repo](https://github.com/lep13/multitenantBackend).

