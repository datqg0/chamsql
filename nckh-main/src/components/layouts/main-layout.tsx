import { Link, useNavigate, useRouterState } from '@tanstack/react-router'
import {
    LogOut,
    User,
    LayoutDashboard,
    Users,
    Shield,
    BookOpen,
    FileText,
    CheckSquare,
    ChevronLeft,
    Database,
} from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ModeToggle } from '@/components/mode-toggle'
import { useLogout } from '@/hooks/use-auth'
import { useAuthStore } from '@/stores/use-auth-store'
import { cn } from '@/lib/utils'
import logoImg from '@/assets/logo.png'

interface MainLayoutProps {
    children: React.ReactNode
}

export function MainLayout({ children }: MainLayoutProps) {
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
    const { user, userRole, isOperator } = useAuthStore()
    const { mutate: logout } = useLogout()
    const navigate = useNavigate()
    const routerState = useRouterState()
    const currentPath = routerState.location.pathname

    const roleName = userRole || (isOperator() ? 'Admin' : 'User')

    const menuItems = [
        {
            title: 'Dashboard',
            href: '/dashboard',
            icon: LayoutDashboard,
            roles: ['admin', 'lecturer'],
        },
        {
            title: 'Quản lý User',
            href: '/users',
            icon: Users,
            roles: ['admin'],
        },
        {
            title: 'Quản lý Role',
            href: '/roles',
            icon: Shield,
            roles: ['admin'],
        },
        {
            title: 'Luyện tập',
            href: '/practice',
            icon: BookOpen,
            roles: ['student', 'all'],
        },
        {
            title: 'Thi trực tuyến',
            href: '/submissions',
            icon: FileText,
            roles: ['student'],
        },
        {
            title: 'Chấm bài',
            href: '/grading',
            icon: CheckSquare,
            roles: ['lecturer', 'admin'],
        },
        {
            title: 'Quản lý kỳ thi',
            href: '/exams',
            icon: FileText,
            roles: ['lecturer', 'admin'],
        },
    ]



    const filteredMenuItems = menuItems.filter((item) => {
        // Admin sees all pages
        if (userRole === 'admin' || isOperator()) return true
        // Other roles see only their allowed pages
        if (item.roles.includes('all')) return true
        if (userRole && item.roles.includes(userRole)) return true
        return false
    })

    const isActive = (href: string) => {
        if (href === '/dashboard') {
            return currentPath === '/dashboard' || currentPath === '/'
        }
        return currentPath.startsWith(href)
    }

    return (
        <div className="min-h-screen bg-background">
            {/* Header */}
            <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
                <div className="flex h-16 items-center justify-between px-4">
                    <div className="flex items-center gap-2 sm:gap-4 flex-1 min-w-0">
                        <Link to="/dashboard" className="flex items-center gap-2 sm:gap-3 min-w-0">
                            <img
                                src={logoImg}
                                alt="Logo"
                                className="h-8 w-8 sm:h-9 sm:w-9 rounded-lg object-cover flex-shrink-0"
                            />
                            <div className="hidden md:block space-y-0.5 min-w-0">
                                <h1 className="text-sm lg:text-lg font-bold leading-none truncate">TRƯỜNG ĐẠI HỌC KIẾN TRÚC HÀ NỘI</h1>
                                <p className="text-xs text-muted-foreground mt-1">KHOA CÔNG NGHỆ THÔNG TIN</p>
                            </div>
                        </Link>
                    </div>

                    <div className="flex items-center gap-2">
                        <ModeToggle />
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button variant="ghost" className="h-9 gap-2 px-2">
                                    <div className="flex h-7 w-7 items-center justify-center rounded-full bg-gradient-to-br from-primary/80 to-accent text-primary-foreground text-sm font-medium">
                                        {user?.name?.charAt(0)?.toUpperCase() || 'U'}
                                    </div>
                                    <div className="hidden md:block text-left">
                                        <p className="text-sm font-medium leading-none">{user?.name}</p>
                                        <p className="text-xs text-muted-foreground">{roleName}</p>
                                    </div>
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" className="w-56">
                                <DropdownMenuLabel>
                                    <div className="flex flex-col space-y-1">
                                        <p className="text-sm font-medium">{user?.name}</p>
                                        <p className="text-xs text-muted-foreground">{user?.email || roleName}</p>
                                    </div>
                                </DropdownMenuLabel>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem onClick={() => navigate({ to: '/dashboard' })}>
                                    <User className="mr-2 h-4 w-4" />
                                    <span>Hồ sơ cá nhân</span>
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    onClick={() => logout()}
                                    className="text-destructive focus:text-destructive"
                                >
                                    <LogOut className="mr-2 h-4 w-4" />
                                    <span>Đăng xuất</span>
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>
                </div>
            </header>

            <div className="flex">
                {/* Desktop Sidebar */}
                <aside
                    className={cn(
                        "hidden lg:flex flex-col h-[calc(100vh-64px)] sticky top-16 transition-all duration-300 border-r bg-card",
                        sidebarCollapsed ? "w-[68px]" : "w-64"
                    )}
                >
                    {/* Sidebar Header with Logo and Toggle */}
                    <div className="flex items-center justify-between p-3 border-b">
                        {/* Logo */}
                        <Link to="/dashboard" className={cn(
                            "flex items-center gap-2 transition-opacity",
                            sidebarCollapsed && "opacity-0 w-0 overflow-hidden"
                        )}>
                            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary to-primary/60 text-primary-foreground">
                                <Database className="h-4 w-4" />
                            </div>
                            <span className="font-bold text-sm">SQL Exam</span>
                        </Link>

                        {/* Toggle Button */}
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
                            className="h-8 w-8 rounded-full hover:bg-accent flex-shrink-0"
                            title={sidebarCollapsed ? "Mở rộng" : "Thu gọn"}
                        >
                            <ChevronLeft className={cn(
                                "h-4 w-4 transition-transform duration-300",
                                sidebarCollapsed && "rotate-180"
                            )} />
                        </Button>
                    </div>

                    {/* Navigation */}
                    <nav className="flex-1 p-3 space-y-1 overflow-y-auto">
                        {filteredMenuItems.map((item) => {
                            const Icon = item.icon
                            const active = isActive(item.href)
                            return (
                                <Link
                                    key={item.href}
                                    to={item.href}
                                    className={cn(
                                        "flex items-center gap-3 px-3 py-2.5 rounded-lg font-medium transition-all duration-200",
                                        "hover:bg-accent/80 hover:text-accent-foreground",
                                        active && "bg-primary/10 text-primary border-l-2 border-primary",
                                        !active && "text-muted-foreground hover:text-foreground",
                                        sidebarCollapsed && "justify-center px-2"
                                    )}
                                    title={sidebarCollapsed ? item.title : undefined}
                                >
                                    <Icon className={cn("h-5 w-5 flex-shrink-0", active && "text-primary")} />
                                    {!sidebarCollapsed && (
                                        <span className="text-sm">{item.title}</span>
                                    )}
                                </Link>
                            )
                        })}
                    </nav>
                </aside>

                {/* Main Content */}
                <main className="flex-1 p-4 md:p-6 min-h-[calc(100vh-64px)] pb-20 lg:pb-6">{children}</main>
            </div>

            {/* Mobile Bottom Navigation */}
            <nav className="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-card border-t shadow-lg">
                <div className="flex items-center justify-around h-16 px-2">
                    {filteredMenuItems.slice(0, 5).map((item) => {
                        const Icon = item.icon
                        const active = isActive(item.href)
                        return (
                            <Link
                                key={item.href}
                                to={item.href}
                                className={cn(
                                    "flex flex-col items-center justify-center gap-1 px-3 py-2 rounded-lg transition-colors min-w-[60px]",
                                    active
                                        ? "text-primary bg-primary/10"
                                        : "text-muted-foreground hover:text-foreground hover:bg-accent/50"
                                )}
                            >
                                <Icon className={cn("h-5 w-5", active && "text-primary")} />
                                <span className="text-[10px] font-medium truncate max-w-[60px]">
                                    {item.title}
                                </span>
                            </Link>
                        )
                    })}
                </div>
            </nav>
        </div >
    )
}
