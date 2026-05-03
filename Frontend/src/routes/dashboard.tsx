import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect, useState, useMemo } from 'react'
import { 
    LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, 
    BarChart, Bar, Cell, PieChart, Pie, Legend 
} from 'recharts'
import { 
    Users, BookOpen, Send, CheckCircle, TrendingUp, Clock, 
    Award, BarChart3, ArrowUpRight, ArrowDownRight, Activity
} from 'lucide-react'

import { MainLayout } from '@/components/layouts/main-layout'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuthStore } from '@/stores/use-auth-store'
import { adminService } from '@/services/admin.service'
import { DashboardResponse, SystemStats } from '@/types/admin.types'
import { Skeleton } from '@/components/ui/skeleton'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

function DashboardPage() {
    const navigate = useNavigate()
    const { user, isOperator, isAuthenticated, userRole } = useAuthStore()
    const [isHydrated, setIsHydrated] = useState(false)
    const [data, setData] = useState<DashboardResponse | null>(null)
    const [stats, setStats] = useState<SystemStats | null>(null)
    const [isLoading, setIsLoading] = useState(true)

    useEffect(() => {
        setIsHydrated(true)
    }, [])

    useEffect(() => {
        if (isHydrated && !isAuthenticated) {
            navigate({ to: '/' })
        }
    }, [isHydrated, isAuthenticated, navigate])

    useEffect(() => {
        const fetchStats = async () => {
            if (isOperator()) {
                try {
                    const [dashboardData, systemStats] = await Promise.all([
                        adminService.getDashboard(),
                        adminService.getSystemStats()
                    ])
                    setData(dashboardData)
                    setStats(systemStats)
                } catch (error) {
                    console.error('Failed to fetch dashboard data:', error)
                } finally {
                    setIsLoading(false)
                }
            } else {
                setIsLoading(false)
            }
        }

        if (isHydrated && isAuthenticated) {
            fetchStats()
        }
    }, [isHydrated, isAuthenticated, isOperator])

    const roleDistribution = useMemo(() => {
        if (!data?.overview?.usersByRole) return []
        return Object.entries(data.overview.usersByRole).map(([name, value]) => ({
            name: name.charAt(0).toUpperCase() + name.slice(1),
            value
        }))
    }, [data])

    if (!isHydrated || !isAuthenticated) {
        return (
            <MainLayout>
                <div className="flex items-center justify-center py-24">
                    <Activity className="h-8 w-8 text-primary animate-pulse mr-2" />
                    <p className="text-muted-foreground animate-pulse">Khởi tạo hệ thống...</p>
                </div>
            </MainLayout>
        )
    }

    if (!isOperator()) {
        return (
            <MainLayout title="Dashboard">
                <div className="space-y-6">
                    <div className="flex flex-col gap-2">
                        <h1 className="text-4xl font-extrabold tracking-tight">Chào {user?.name || 'bạn'}! 👋</h1>
                        <p className="text-muted-foreground text-lg">
                            Chào mừng bạn quay trở lại với hệ thống chấm SQL tự động.
                        </p>
                    </div>
                    
                    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
                        <Card className="border-l-4 border-l-blue-500 shadow-md hover:shadow-lg transition-shadow">
                            <CardHeader className="pb-2">
                                <CardDescription className="font-medium">Vai trò</CardDescription>
                                <CardTitle className="text-2xl font-bold flex items-center gap-2">
                                    <Award className="h-5 w-5 text-blue-500" />
                                    {userRole}
                                </CardTitle>
                            </CardHeader>
                        </Card>
                        <Card className="border-l-4 border-l-purple-500 shadow-md hover:shadow-lg transition-shadow">
                            <CardHeader className="pb-2">
                                <CardDescription className="font-medium">Tài khoản</CardDescription>
                                <CardTitle className="text-xl font-bold truncate">{user?.email}</CardTitle>
                            </CardHeader>
                        </Card>
                    </div>

                    <Card className="overflow-hidden border-none shadow-xl bg-gradient-to-br from-primary/5 to-primary/10">
                        <CardHeader>
                            <CardTitle>Sẵn sàng luyện tập?</CardTitle>
                            <CardDescription>Bắt đầu giải quyết các thử thách SQL ngay hôm nay.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <button 
                                onClick={() => navigate({ to: '/practice' })}
                                className="px-6 py-3 bg-primary text-primary-foreground rounded-full font-bold hover:opacity-90 transition-all transform hover:scale-105 shadow-lg flex items-center gap-2"
                            >
                                <Send className="h-4 w-4" />
                                Vào luyện tập ngay
                            </button>
                        </CardContent>
                    </Card>
                </div>
            </MainLayout>
        )
    }

    return (
        <MainLayout title="Hệ thống Báo cáo & Phân tích">
            <div className="space-y-8 max-w-[1600px] mx-auto pb-12">
                {/* Header section with glassmorphism feel */}
                <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
                    <div className="space-y-1">
                        <h1 className="text-4xl font-black tracking-tighter bg-clip-text text-transparent bg-gradient-to-r from-blue-600 to-indigo-600">
                            DASHBOARD ANALYTICS
                        </h1>
                        <p className="text-muted-foreground font-medium flex items-center gap-2">
                            <Clock className="h-4 w-4" />
                            Dữ liệu được cập nhật thời gian thực từ hệ thống.
                        </p>
                    </div>
                    <div className="flex gap-2">
                        <div className="px-4 py-2 bg-background border rounded-lg shadow-sm flex items-center gap-2 text-sm font-bold">
                            <div className="h-2 w-2 rounded-full bg-green-500 animate-ping" />
                            Hệ thống ổn định
                        </div>
                    </div>
                </div>

                {/* Quick Stats Grid */}
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                    <Card className="relative overflow-hidden group">
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-bold text-muted-foreground uppercase tracking-wider">Tổng Sinh viên</CardTitle>
                            <Users className="h-4 w-4 text-blue-500" />
                        </CardHeader>
                        <CardContent>
                            {isLoading ? <Skeleton className="h-9 w-24" /> : (
                                <>
                                    <div className="text-3xl font-black">{(data?.overview?.totalUsers || 0).toLocaleString()}</div>
                                    <div className="text-xs text-green-600 font-bold flex items-center mt-1">
                                        <ArrowUpRight className="h-3 w-3 mr-1" />
                                        +12% <span className="text-muted-foreground font-medium ml-1">so với tháng trước</span>
                                    </div>
                                </>
                            )}
                        </CardContent>
                        <div className="absolute bottom-0 left-0 h-1 w-full bg-blue-500 transform scale-x-0 group-hover:scale-x-100 transition-transform origin-left" />
                    </Card>

                    <Card className="relative overflow-hidden group">
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-bold text-muted-foreground uppercase tracking-wider">Lượt Nộp bài</CardTitle>
                            <Send className="h-4 w-4 text-indigo-500" />
                        </CardHeader>
                        <CardContent>
                            {isLoading ? <Skeleton className="h-9 w-24" /> : (
                                <>
                                    <div className="text-3xl font-black">{(data?.overview?.totalSubmissions || 0).toLocaleString()}</div>
                                    <div className="text-xs text-green-600 font-bold flex items-center mt-1">
                                        <ArrowUpRight className="h-3 w-3 mr-1" />
                                        +24% <span className="text-muted-foreground font-medium ml-1">lượt mới hôm nay</span>
                                    </div>
                                </>
                            )}
                        </CardContent>
                        <div className="absolute bottom-0 left-0 h-1 w-full bg-indigo-500 transform scale-x-0 group-hover:scale-x-100 transition-transform origin-left" />
                    </Card>

                    <Card className="relative overflow-hidden group">
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-bold text-muted-foreground uppercase tracking-wider">Tỉ lệ Pass</CardTitle>
                            <CheckCircle className="h-4 w-4 text-emerald-500" />
                        </CardHeader>
                        <CardContent>
                            {isLoading ? <Skeleton className="h-9 w-24" /> : (
                                <>
                                    <div className="text-3xl font-black">{(data?.gradingStats?.passRate || 0).toFixed(1)}%</div>
                                    <div className="text-xs text-red-600 font-bold flex items-center mt-1">
                                        <ArrowDownRight className="h-3 w-3 mr-1" />
                                        -2.4% <span className="text-muted-foreground font-medium ml-1">giảm nhẹ</span>
                                    </div>
                                </>
                            )}
                        </CardContent>
                        <div className="absolute bottom-0 left-0 h-1 w-full bg-emerald-500 transform scale-x-0 group-hover:scale-x-100 transition-transform origin-left" />
                    </Card>

                    <Card className="relative overflow-hidden group">
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-bold text-muted-foreground uppercase tracking-wider">Active tuần</CardTitle>
                            <TrendingUp className="h-4 w-4 text-orange-500" />
                        </CardHeader>
                        <CardContent>
                            {isLoading ? <Skeleton className="h-9 w-24" /> : (
                                <>
                                    <div className="text-3xl font-black">{(data?.overview?.activeUsersWeek || 0).toLocaleString()}</div>
                                    <div className="text-xs text-green-600 font-bold flex items-center mt-1">
                                        <ArrowUpRight className="h-3 w-3 mr-1" />
                                        +5.2% <span className="text-muted-foreground font-medium ml-1">đang tăng trưởng</span>
                                    </div>
                                </>
                            )}
                        </CardContent>
                        <div className="absolute bottom-0 left-0 h-1 w-full bg-orange-500 transform scale-x-0 group-hover:scale-x-100 transition-transform origin-left" />
                    </Card>
                </div>

                <div className="grid gap-6 grid-cols-1 lg:grid-cols-7">
                    {/* Submission Trend - Large Main Chart */}
                    <Card className="lg:col-span-4 shadow-xl border-none bg-background/50 backdrop-blur-sm">
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <div>
                                    <CardTitle className="text-xl font-black">Xu hướng Nộp bài</CardTitle>
                                    <CardDescription>Thống kê số lượng bài nộp và kết quả đúng 7 ngày gần nhất.</CardDescription>
                                </div>
                                <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                                    <BarChart3 className="h-5 w-5 text-primary" />
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent className="pl-2">
                            <div className="h-[350px] w-full">
                                {isLoading ? <Skeleton className="h-full w-full" /> : (
                                    <ResponsiveContainer width="100%" height="100%">
                                        <LineChart data={data?.dailySubmissions || []}>
                                            <defs>
                                                <linearGradient id="colorTotal" x1="0" y1="0" x2="0" y2="1">
                                                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.8}/>
                                                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                                                </linearGradient>
                                            </defs>
                                            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e2e8f0" />
                                            <XAxis 
                                                dataKey="date" 
                                                axisLine={false} 
                                                tickLine={false} 
                                                tick={{ fill: '#64748b', fontSize: 12 }} 
                                                dy={10}
                                            />
                                            <YAxis 
                                                axisLine={false} 
                                                tickLine={false} 
                                                tick={{ fill: '#64748b', fontSize: 12 }} 
                                            />
                                            <Tooltip 
                                                contentStyle={{ backgroundColor: 'white', borderRadius: '12px', border: 'none', boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.1)' }}
                                            />
                                            <Legend verticalAlign="top" height={36}/>
                                            <Line 
                                                name="Tổng lượt nộp"
                                                type="monotone" 
                                                dataKey="totalSubmissions" 
                                                stroke="#3b82f6" 
                                                strokeWidth={4} 
                                                dot={{ r: 4, strokeWidth: 2, fill: 'white' }}
                                                activeDot={{ r: 8 }}
                                            />
                                            <Line 
                                                name="Đúng (Correct)"
                                                type="monotone" 
                                                dataKey="correctCount" 
                                                stroke="#10b981" 
                                                strokeWidth={3} 
                                                strokeDasharray="5 5"
                                                dot={{ r: 3 }}
                                            />
                                        </LineChart>
                                    </ResponsiveContainer>
                                )}
                            </div>
                        </CardContent>
                    </Card>

                    {/* Role Distribution - Side Chart */}
                    <Card className="lg:col-span-3 shadow-xl border-none bg-background/50 backdrop-blur-sm">
                        <CardHeader>
                            <CardTitle className="text-xl font-black">Cơ cấu Người dùng</CardTitle>
                            <CardDescription>Phân bố theo vai trò trong hệ thống.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="h-[300px] w-full">
                                {isLoading ? <Skeleton className="h-full w-full" /> : (
                                    <ResponsiveContainer width="100%" height="100%">
                                        <PieChart>
                                            <Pie
                                                data={roleDistribution}
                                                cx="50%"
                                                cy="50%"
                                                innerRadius={60}
                                                outerRadius={100}
                                                paddingAngle={5}
                                                dataKey="value"
                                            >
                                                {roleDistribution.map((_, index) => (
                                                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                                ))}
                                            </Pie>
                                            <Tooltip />
                                            <Legend layout="vertical" verticalAlign="middle" align="right" />
                                        </PieChart>
                                    </ResponsiveContainer>
                                )}
                            </div>
                            <div className="mt-4 space-y-2">
                                {roleDistribution.map((role, index) => (
                                    <div key={role.name} className="flex items-center justify-between text-sm">
                                        <div className="flex items-center gap-2 font-medium">
                                            <div className="h-2 w-2 rounded-full" style={{ backgroundColor: COLORS[index % COLORS.length] }} />
                                            {role.name}
                                        </div>
                                        <span className="font-bold">{role.value}</span>
                                    </div>
                                ))}
                            </div>
                        </CardContent>
                    </Card>
                </div>

                <div className="grid gap-6 grid-cols-1 lg:grid-cols-2">
                    {/* Top Problems */}
                    <Card className="shadow-xl border-none">
                        <CardHeader>
                            <CardTitle className="text-xl font-black">Top Bài tập Phổ biến</CardTitle>
                            <CardDescription>Các bài tập có lượt tham gia nhiều nhất.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-6">
                                {isLoading ? [1, 2, 3].map(i => <Skeleton key={i} className="h-16 w-full" />) : (
                                    (data?.topProblems || []).length > 0 ? (
                                        data?.topProblems.map((prob, index) => (
                                            <div key={prob.id} className="flex items-center gap-4 group cursor-pointer" onClick={() => navigate({ to: `/problems/${prob.slug}` as any })}>
                                                <div className="h-12 w-12 rounded-xl bg-muted flex items-center justify-center text-xl font-black text-muted-foreground group-hover:bg-primary group-hover:text-primary-foreground transition-colors">
                                                    {index + 1}
                                                </div>
                                                <div className="flex-1 space-y-1">
                                                    <div className="font-bold group-hover:text-primary transition-colors">{prob.title}</div>
                                                    <div className="flex items-center gap-3 text-xs font-medium">
                                                        <span className={`px-2 py-0.5 rounded uppercase ${
                                                            prob.difficulty === 'easy' ? 'bg-green-100 text-green-700' :
                                                            prob.difficulty === 'medium' ? 'bg-orange-100 text-orange-700' :
                                                            'bg-red-100 text-red-700'
                                                        }`}>
                                                            {prob.difficulty}
                                                        </span>
                                                        <span className="text-muted-foreground flex items-center gap-1">
                                                            <Users className="h-3 w-3" /> {prob.uniqueUsers} người làm
                                                        </span>
                                                    </div>
                                                </div>
                                                <div className="text-right">
                                                    <div className="font-black text-lg">{prob.submissionCount}</div>
                                                    <div className="text-[10px] uppercase font-bold text-muted-foreground">Lượt nộp</div>
                                                </div>
                                            </div>
                                        ))
                                    ) : (
                                        <div className="text-center py-12 text-muted-foreground italic">Chưa có bài tập nào</div>
                                    )
                                )}
                            </div>
                        </CardContent>
                    </Card>

                    {/* Recent Activity */}
                    <Card className="shadow-xl border-none">
                        <CardHeader>
                            <CardTitle className="text-xl font-black">Hoạt động Hệ thống</CardTitle>
                            <CardDescription>Các sự kiện mới nhất vừa diễn ra.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-6">
                                {isLoading ? [1, 2, 3, 4].map(i => <Skeleton key={i} className="h-12 w-full" />) : (
                                    stats?.recentActivity && stats.recentActivity.length > 0 ? (
                                        stats.recentActivity.slice(0, 6).map((activity, idx) => (
                                            <div key={idx} className="flex items-start gap-4">
                                                <div className={`mt-1 h-8 w-8 rounded-full flex items-center justify-center shrink-0 ${
                                                    activity.type === 'submission' ? 'bg-blue-100 text-blue-600' :
                                                    activity.type === 'user_registered' ? 'bg-green-100 text-green-600' :
                                                    'bg-purple-100 text-purple-600'
                                                }`}>
                                                    {activity.type === 'submission' ? <Send className="h-4 w-4" /> :
                                                     activity.type === 'user_registered' ? <Users className="h-4 w-4" /> :
                                                     <Award className="h-4 w-4" />}
                                                </div>
                                                <div className="flex-1 space-y-1">
                                                    <p className="text-sm font-medium leading-tight">{activity.message}</p>
                                                    <p className="text-[10px] text-muted-foreground font-bold flex items-center gap-1 uppercase">
                                                        <Clock className="h-3 w-3" />
                                                        {new Date(activity.timestamp).toLocaleTimeString()} • {new Date(activity.timestamp).toLocaleDateString()}
                                                    </p>
                                                </div>
                                            </div>
                                        ))
                                    ) : (
                                        <div className="text-center py-12 text-muted-foreground italic">Chưa có hoạt động nào</div>
                                    )
                                )}
                            </div>
                        </CardContent>
                    </Card>
                </div>

                <div className="grid gap-6 grid-cols-1 lg:grid-cols-3">
                    {/* Pass Rates Bar Chart - Now in a wider section */}
                    <Card className="lg:col-span-2 shadow-xl border-none">
                        <CardHeader>
                            <CardTitle className="text-xl font-black">Tỉ lệ Hoàn thành theo Bài tập</CardTitle>
                            <CardDescription>So sánh tỉ lệ vượt qua giữa các bài tập tiêu biểu.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="h-[300px] w-full">
                                {isLoading ? <Skeleton className="h-full w-full" /> : (
                                    <ResponsiveContainer width="100%" height="100%">
                                        <BarChart data={data?.passRates || []} layout="vertical" margin={{ left: 40, right: 20 }}>
                                            <CartesianGrid strokeDasharray="3 3" horizontal={true} vertical={false} />
                                            <XAxis type="number" hide domain={[0, 100]} />
                                            <YAxis 
                                                dataKey="title" 
                                                type="category" 
                                                axisLine={false} 
                                                tickLine={false}
                                                width={120}
                                                tick={{ fill: '#64748b', fontSize: 10, fontWeight: 700 }}
                                            />
                                            <Tooltip 
                                                cursor={{ fill: 'rgba(0,0,0,0.02)' }}
                                                content={({ active, payload }) => {
                                                    if (active && payload && payload.length) {
                                                        const p = payload[0].payload
                                                        return (
                                                            <div className="bg-background border p-3 rounded-xl shadow-2xl">
                                                                <p className="font-bold text-sm mb-1">{p.title}</p>
                                                                <div className="flex items-center gap-2">
                                                                    <div className="h-2 w-2 rounded-full bg-emerald-500" />
                                                                    <p className="text-xs text-emerald-600 font-bold">Pass Rate: {p.passRate.toFixed(1)}%</p>
                                                                </div>
                                                                <p className="text-[10px] text-muted-foreground mt-1 font-medium">
                                                                    Thành công: {p.correctCount} / {p.totalSubmissions} bài nộp
                                                                </p>
                                                            </div>
                                                        )
                                                    }
                                                    return null
                                                }}
                                            />
                                            <Bar dataKey="passRate" radius={[0, 8, 8, 0]} barSize={20}>
                                                {(data?.passRates || []).map((_, index) => (
                                                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                                ))}
                                            </Bar>
                                        </BarChart>
                                    </ResponsiveContainer>
                                )}
                            </div>
                        </CardContent>
                    </Card>

                    {/* Additional Metrics Card */}
                    <Card className="shadow-xl border-none bg-gradient-to-br from-indigo-50 to-blue-50 dark:from-indigo-950/20 dark:to-blue-950/20">
                        <CardHeader>
                            <CardTitle className="text-lg font-black uppercase tracking-tight">Thông tin thêm</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="p-4 bg-background rounded-2xl shadow-sm border border-indigo-100 flex items-center justify-between">
                                <div className="space-y-1">
                                    <p className="text-[10px] font-black text-muted-foreground uppercase">Thời gian giải TB</p>
                                    <p className="text-xl font-black text-indigo-600">{data?.overview?.avgSolveTimeMs || 0}ms</p>
                                </div>
                                <Clock className="h-8 w-8 text-indigo-200" />
                            </div>
                            <div className="p-4 bg-background rounded-2xl shadow-sm border border-emerald-100 flex items-center justify-between">
                                <div className="space-y-1">
                                    <p className="text-[10px] font-black text-muted-foreground uppercase">Tỉ lệ đạt hệ thống</p>
                                    <p className="text-xl font-black text-emerald-600">{(data?.gradingStats?.passRate || 0).toFixed(1)}%</p>
                                </div>
                                <CheckCircle className="h-8 w-8 text-emerald-200" />
                            </div>
                            <div className="p-4 bg-background rounded-2xl shadow-sm border border-orange-100 flex items-center justify-between">
                                <div className="space-y-1">
                                    <p className="text-[10px] font-black text-muted-foreground uppercase">Tổng bài tập</p>
                                    <p className="text-xl font-black text-orange-600">{data?.overview?.totalProblems || 0}</p>
                                </div>
                                <BookOpen className="h-8 w-8 text-orange-200" />
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* Footer Section - Metrics Detail */}
                <div className="grid gap-6 md:grid-cols-3">
                    <Card className="bg-blue-600 text-white border-none shadow-lg">
                        <CardHeader className="pb-2">
                            <CardTitle className="text-lg font-bold flex items-center gap-2">
                                <Clock className="h-5 w-5" />
                                Hiệu năng chấm
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-3xl font-black">{data?.gradingStats.avgGradingTimeMs || 0}ms</div>
                            <p className="text-blue-100 text-xs mt-1 font-medium">Thời gian chấm trung bình cho mỗi bài nộp.</p>
                        </CardContent>
                    </Card>
                    <Card className="bg-indigo-600 text-white border-none shadow-lg">
                        <CardHeader className="pb-2">
                            <CardTitle className="text-lg font-bold flex items-center gap-2">
                                <BookOpen className="h-5 w-5" />
                                Bài tập thử thách
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-3xl font-black">{data?.gradingStats.totalProblemsAttempted || 0}</div>
                            <p className="text-indigo-100 text-xs mt-1 font-medium">Số lượng bài tập đã có sinh viên tham gia giải quyết.</p>
                        </CardContent>
                    </Card>
                    <Card className="bg-emerald-600 text-white border-none shadow-lg">
                        <CardHeader className="pb-2">
                            <CardTitle className="text-lg font-bold flex items-center gap-2">
                                <Award className="h-5 w-5" />
                                Tỉ lệ Thành công
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-3xl font-black">{data?.gradingStats.totalCorrect.toLocaleString()}</div>
                            <p className="text-emerald-100 text-xs mt-1 font-medium">Tổng số lượt nộp bài đạt trạng thái Accepted.</p>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </MainLayout>
    )
}

export const Route = createFileRoute('/dashboard')({
    component: DashboardPage,
})
