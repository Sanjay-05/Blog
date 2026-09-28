# Blog Post

This project is a backend service for managing blog posts using Go.

<p align="center">
  <img src="https://github.com/user-attachments/assets/f9f2b3bc-8b81-43b5-ab0c-2c23bf865535" width="100" />
  <img src="https://github.com/user-attachments/assets/ccb5957b-bfb6-4e0e-8bfd-3e7a90859505" width="100" />
  <img width="100" height="70" alt="image" src="https://github.com/user-attachments/assets/ad488acd-4fdd-418a-9ac2-1b76a4c89bc1" />
  <img width="100" height="70" alt="image" src="https://github.com/user-attachments/assets/00606063-9d2f-479e-bcc5-0a2ba8b61267" />
</p>




## 🎯 Features
- Create, read, update, delete blog posts, add comments
- JWT authentication
- Built using Fiber framework
- For Database used PostgreSQL and GORM for ORM Tool

## 🚀 Installation
1. Clone the repo: `git clone https://github.com/NishanthSaravanan/Senthil-Nov2025-RTE.git`
2. Install dependencies: `go mod tidy`
3. Run the app: `go run main.go`

## Usage
- Register a user via `/api/v1/auth/register`
- Log in via `/api/v1/auth/login` to get a JWT token
- Use the token to access protected routes like `/api/v1/public/posts`
## 📁 Project Structure

```
RTE_blog/
├── main.go
├── config/        # Configuration management
├── db/            # Database repositories and data access
├── routes/        # Router definitions
├── handlers/      # HTTP request handlers
├── models/        # Application models
├── middleware/    # HTTP middleware components
└── go.mod
```

### ➡️ Control Flow

```mermaid
flowchart LR
   Request e1@==> Server
   e1@{ animate: true }
   Server e2@==> Router
   e2@{ animate: true }
   M@{ shape: processes, label: "Middlewares" }
   Router e3@==> M
   e3@{ animate: true }
   C@{ shape: processes, label: "Handlers" }
   M e4@==> C
   e4@{ animate: true }
   R@{ shape: processes, label: "Repository" }
   C e5@==> R
   e5@{ animate: true }
   id1[(Database)]
   R e6@==> id1
   e6@{ animate: true }
```

1. **Request**: Server receives the incoming request.
2. **Server**: Server processes the request and forwards it to the router.
3. **Router**: Router directs the request to the appropriate middleware(s).
4. **Middlewares**: The middlewares handle various tasks such as logging, authentication, security headers, tracing etc.,
5. **Handlers**: The request is passed to the appropriate handler, which validates the request and forwards it to the repository layer.
6. **Repository**: The repository layer communicates with the database to perform CRUD operations.


## 🛠 Tools & Stack

| Category | Technology |
|----------|-----------|
| **Web Framework** | [Fiber](https://github.com/gofiber/fiber) |
| **Logging** | [Logrus](https://pkg.go.dev/github.com/sirupsen/logrus@v1.9.3) |
| **Database** | [PostgreSQL](https://pkg.go.dev/gorm.io/driver/postgres@v1.6.0) |
| **ORM Tool** | [GORM](https://pkg.go.dev/gorm.io/gorm@v1.31.1) |


## Contributing
Feel free to fork the repo and submit pull requests.

## Author
Senthilkanagaraj

GitHub: https://github.com/Senthilkanagaraj