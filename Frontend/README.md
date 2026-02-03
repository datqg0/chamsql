# Hệ Thống Chấm Thi và Gỡ Lỗi Truy Vấn SQL

Hệ thống web application cho phép sinh viên nộp bài tập SQL, giảng viên chấm bài và quản trị viên quản lý người dùng/phân quyền.

## 📋 Tech Stack

| Công nghệ | Mô tả |
|-----------|-------|
| **React 19** | UI Framework |
| **TypeScript** | Type-safe JavaScript |
| **Vite** | Build tool & Dev server |
| **TanStack Router** | File-based routing |
| **TanStack Query** | Server state management |
| **Zustand** | Client state management (Auth) |
| **TailwindCSS 4.0** | Utility-first CSS |
| **shadcn/ui** | UI Components |
| **React Hook Form + Zod** | Form handling & validation |
| **Monaco Editor** | SQL code editor |
| **WebSocket** | Real-time communication |
| **Axios** | HTTP client |

## 🏗️ Cấu trúc dự án

```
src/
├── components/          # React components
│   ├── ui/              # shadcn/ui components (Button, Card, Dialog, etc.)
│   ├── layouts/         # Layout components (MainLayout)
│   ├── forms/           # Form components (UserForm, RoleForm)
│   ├── editor/          # SQL Editor component
│   └── auth/            # Auth-related components
├── routes/              # TanStack Router pages (file-based routing)
│   ├── __root.tsx       # Root layout với providers
│   ├── index.tsx        # Trang đăng nhập (/)
│   ├── register.tsx     # Trang đăng ký
│   ├── dashboard.tsx    # Dashboard (protected)
│   ├── submissions.tsx  # Nộp bài SQL
│   ├── grading.tsx      # Chấm bài
│   ├── users.tsx        # Quản lý người dùng
│   ├── roles.tsx        # Quản lý vai trò
│   └── permissions.tsx  # Phân quyền
├── services/            # API services
│   ├── api/             # Axios client & endpoints
│   ├── auth.service.ts  # Authentication API
│   ├── user.service.ts  # User management API
│   ├── role.service.ts  # Role management API
│   ├── permission.service.ts # Permission API
│   └── websocket.service.ts  # WebSocket service
├── stores/              # Zustand stores
│   └── use-auth-store.ts # Auth state (token, user, permissions)
├── hooks/               # Custom React hooks
│   ├── use-auth.ts      # Auth mutations (login, logout, register)
│   └── use-websocket.ts # WebSocket hook
├── types/               # TypeScript types
│   └── auth.types.ts    # User, Role, Permission types
├── config/              # Configuration
│   └── env.ts           # Environment variables validation
└── lib/                 # Utilities
    └── utils.ts         # Helper functions (cn, etc.)
```

## 🔄 Luồng hoạt động

### Authentication Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                        AUTHENTICATION FLOW                       │
├─────────────────────────────────────────────────────────────────┤
│  Login (/) ──▶ authService.login() ──▶ JWT Token                │
│                    │                                             │
│                    ▼                                             │
│  Decode JWT ──▶ useAuthStore.setAuth() ──▶ localStorage persist │
│                    │                                             │
│                    ▼                                             │
│  Redirect to /dashboard (protected route)                       │
└─────────────────────────────────────────────────────────────────┘
```

### Luồng nộp bài SQL

```
User viết SQL (Monaco Editor)
       │
       ▼
┌──────────────────┐    WebSocket     ┌─────────────────┐
│  Chạy thử (Test) │ ──────────────▶  │   Backend API   │
│  Nộp bài (Submit)│ ◀──────────────  │   (SQL Engine)  │
└──────────────────┘    Kết quả/      └─────────────────┘
                        Feedback
```

## 📍 Routes & Phân quyền

| Route | Mô tả | Quyền truy cập |
|-------|-------|----------------|
| `/` | Trang đăng nhập | Public |
| `/register` | Trang đăng ký | Public |
| `/dashboard` | Bảng điều khiển | Đã đăng nhập |
| `/submissions` | Nộp bài SQL | Student, All |
| `/grading` | Chấm bài | Teacher, Admin, Operator |
| `/users` | Quản lý người dùng | Admin, Operator |
| `/roles` | Quản lý vai trò | Admin, Operator |
| `/permissions` | Phân quyền | Admin, Operator |

### Hệ thống Role

| Role ID | Tên | Mô tả |
|---------|-----|-------|
| 1 | Admin | Quản trị viên |
| 2 | Teacher | Giảng viên |
| 3 | Student | Sinh viên |
| isOperator = 1 | Operator | Full quyền hệ thống |

## 🔌 API Endpoints

### Authentication
- `POST /user/authenticate` - Đăng nhập
- `POST /user/create` - Đăng ký
- `POST /auth/logout` - Đăng xuất
- `GET /auth/me` - Lấy thông tin user hiện tại

### Users
- `GET /user/list` - Danh sách users
- `GET /user/getById` - Lấy user theo ID
- `PUT /user/update` - Cập nhật user
- `DELETE /user/delete/:id` - Xóa user

### Roles
- `GET /role/list` - Danh sách roles
- `PUT /role/update` - Cập nhật role
- `DELETE /role/delete/:id` - Xóa role

### Permissions
- `GET /permission/info` - Lấy thông tin permissions theo role
- `POST /permission/add` - Thêm permission
- `POST /permission/remove` - Xóa permission

## 🚀 Cài đặt & Chạy

### Yêu cầu
- Node.js >= 18
- pnpm (hoặc npm/yarn)

### Cài đặt

```bash
# Clone project
git clone <repository-url>
cd templateUi

# Cài đặt dependencies
pnpm install

# Tạo file .env từ .env.example
cp .env.example .env
```

### Cấu hình môi trường

```env
VITE_API_BASE_URL=http://your-api-url
VITE_WS_URL=ws://your-websocket-url
VITE_APP_NAME=SQL Exam System
```

### Chạy development

```bash
# Chạy dev server
pnpm dev

# Chạy route watcher (optional - auto generate routes)
pnpm dev:routes
```

### Build production

```bash
pnpm build
pnpm preview
```

## 🐳 Docker

```bash
# Tạo docker network
docker network create templateui_devnet
```

## 📁 State Management

### Auth Store (Zustand)

```typescript
interface AuthState {
    token: string | null
    user: User | null
    permissions: Permission[]
    isAuthenticated: boolean
    roleId: number | null
    
    setAuth: (token, user, permissions?) => void
    logout: () => void
    hasPermission: (resourceUri, action) => boolean
    isOperator: () => boolean
}
```

- State được persist vào `localStorage` với key `auth-storage`
- Token được gửi tự động qua Axios interceptor

## 🔐 Protected Routes

Routes được bảo vệ bằng `beforeLoad` hook:

```typescript
export const Route = createFileRoute('/dashboard')({
    component: DashboardPage,
    beforeLoad: () => {
        const authStore = JSON.parse(localStorage.getItem('auth-storage') || '{}')
        if (!authStore?.state?.isAuthenticated) {
            throw new Error('Unauthorized')
        }
    },
    errorComponent: ({ error }) => {
        if (error.message === 'Unauthorized') {
            window.location.href = '/'
            return null
        }
        return <div>Error: {error.message}</div>
    },
})
```

## 📝 License

MIT