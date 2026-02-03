export const API_ENDPOINTS = {
    auth: {
        login: '/auth/login',
        register: '/auth/register',
    },
    topics: {
        list: '/topics',
        bySlug: (slug: string) => `/topics/${slug}`,
        create: '/topics',
    },
    problems: {
        list: '/problems',
        bySlug: (slug: string) => `/problems/${slug}`,
        create: '/problems',
        run: (id: number) => `/problems/${id}/run`,
        submit: (id: number) => `/problems/${id}/submit`,
    },
    submissions: {
        list: '/submissions',
    },
    exams: {
        list: '/exams',
        create: '/exams',
        addProblem: (id: number) => `/exams/${id}/problems`,
        addParticipants: (id: number) => `/exams/${id}/participants`,
        start: (id: number) => `/exams/${id}/start`,
        submit: (id: number) => `/exams/${id}/submit`,
        finish: (id: number) => `/exams/${id}/finish`,
        myExams: '/my-exams',
    },
    admin: {
        stats: '/admin/stats',
        users: '/admin/users',
        importUsers: '/admin/users/import',
        updateRole: (id: number) => `/admin/users/${id}/role`,
        toggleActive: (id: number) => `/admin/users/${id}/toggle-active`,
    },
    health: '/health',
} as const

