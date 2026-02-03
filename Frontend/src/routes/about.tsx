import { createFileRoute } from '@tanstack/react-router'

import { ModeToggle } from '@/components/mode-toggle'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

function About() {
    return (
        <div className="min-h-screen flex flex-col items-center justify-center p-4">
            <div className="absolute top-4 right-4">
                <ModeToggle />
            </div>

            <Card className="w-full max-w-2xl">
                <CardHeader>
                    <CardTitle className="text-3xl font-bold">About This Stack</CardTitle>
                    <CardDescription>
                        Built with modern tools and best practices
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                    <div>
                        <h3 className="text-lg font-semibold mb-2">Tech Stack</h3>
                        <ul className="list-disc list-inside space-y-1 text-sm text-muted-foreground">
                            <li>React 18 + TypeScript</li>
                            <li>Vite for blazing fast development</li>
                            <li>TanStack Router for type-safe routing</li>
                            <li>React Query for server state management</li>
                            <li>Zustand for client state management</li>
                            <li>Tailwind CSS + shadcn/ui for styling</li>
                            <li>Axios for HTTP requests</li>
                        </ul>
                    </div>

                    <div>
                        <h3 className="text-lg font-semibold mb-2">Features</h3>
                        <ul className="list-disc list-inside space-y-1 text-sm text-muted-foreground">
                            <li>✅ File-based routing with auto-generated types</li>
                            <li>✅ Dark/Light mode with system preference support</li>
                            <li>✅ API integration with interceptors</li>
                            <li>✅ Form handling with validation</li>
                            <li>✅ Authentication flow ready</li>
                            <li>✅ Production-ready folder structure</li>
                        </ul>
                    </div>

                    <div className="flex gap-2 pt-4">
                        <Button asChild>
                            <a href="/">Back to Home</a>
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    )
}

export const Route = createFileRoute('/about')({
    component: About,
})
