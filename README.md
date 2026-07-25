# AuthD - Identity Management Service

**Version:** 1.0.0

## Overview

AuthD is a lightweight identity and access management service built with Go. It provides user authentication, authorization, and project-based access control capabilities. This is version 1.0.0 - a foundational release that will scale and expand in future versions.

## Features (v1.0.0)

### Authentication
- User registration with validation
- Login with JWT token generation
- Token refresh mechanism
- Password change/update
- Account deletion

### Project Management
- Create projects
- Add members to projects
- Remove members from projects
- List project members
- List user's projects
- Delete projects (cascading delete)

### Access Control
- Role-based access (owner, admin, member)
- Owner-only operations enforcement
- Admin-level permissions for specific actions

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/authd/register` | Register new user |
| POST | `/api/v1/authd/login` | Login and get JWT token |
| POST | `/api/v1/authd/refresh` | Refresh access token |
| POST | `/api/v1/authd/user/update` | Change password |
| DELETE | `/api/v1/authd/user/delete` | Delete account |

### Projects
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/authd/projects/create` | Create a new project |
| POST | `/api/v1/authd/projects/members/add` | Add members to project |
| DELETE | `/api/v1/authd/projects/members/remove` | Remove member from project |
| GET | `/api/v1/authd/projects/users/list` | List all users |
| GET | `/api/v1/authd/projects/members/list` | List project members |
| GET | `/api/v1/authd/projects/list` | List user's projects |
| DELETE | `/api/v1/authd/projects/delete` | Delete project |

## Database Schema

### Users Table
```sql
CREATE TABLE `users` (
  `id` char(36) NOT NULL DEFAULT (uuid()),
  `email` varchar(255) NOT NULL,
  `username` varchar(100) NOT NULL,
  `password_hash` text NOT NULL,
  `is_verified` tinyint(1) DEFAULT '0',
  `is_active` tinyint(1) DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

### Projects Table
```sql
CREATE TABLE `projects` (
  `id` char(36) NOT NULL DEFAULT (uuid()),
  `name` varchar(100) NOT NULL,
  `description` text,
  `owner_id` varchar(50) NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `projects_ibfk_1` (`owner_id`),
  CONSTRAINT `projects_ibfk_1` FOREIGN KEY (`owner_id`) REFERENCES `users` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

### Project Members Table
```sql
CREATE TABLE `project_members` (
  `project_id` char(36) NOT NULL,
  `user_id` varchar(20) NOT NULL,
  `role` enum('admin','user','owner') NOT NULL DEFAULT 'user',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`project_id`,`user_id`),
  KEY `fk_user` (`user_id`),
  CONSTRAINT `fk_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`username`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

## Requirements

- Go 1.21+
- MySQL 8.0+
- GCC (for CGO dependencies)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/grudge007/authd.git
cd authd
```

2. Install dependencies:
```bash
go mod download
```

3. Configure environment variables:
Create a `.env` file with the following variables:
```
JWT_KEY=your-secret-jwt-key
DSN=username:password@tcp(localhost:3306)/database_name
```

4. Run the server:
```bash
go run app/api/main.go
```

The server will start on `http://localhost:7070`

## Example Usage

### Register a User
```bash
curl -X POST http://localhost:7070/api/v1/authd/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret123","email":"john@example.com"}'
```

### Login
```bash
curl -X POST http://localhost:7070/api/v1/authd/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret123"}'
```

### Create a Project
```bash
curl -X POST http://localhost:7070/api/v1/authd/projects/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"project_name":"my-project","project_desc":"My awesome project"}'
```

## Future Roadmap

v1.0.0 is just the beginning. Future versions will include:

- Email verification
- Password reset flow
- Two-factor authentication (2FA)
- Session management
- Audit logging
- Permission system enhancements
- API rate limiting
- User profile management
- Pagination for list endpoints
- Search and filter capabilities

## License

See LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
