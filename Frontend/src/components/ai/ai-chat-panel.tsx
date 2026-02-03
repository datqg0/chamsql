import { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import {
    MessageCircle,
    X,
    Send,
    Loader2,
    Bot,
    User,
    Sparkles,
    Minimize2,
    Maximize2,
} from 'lucide-react'
import { cn } from '@/lib/utils'

export interface ChatMessage {
    id: string
    role: 'user' | 'assistant'
    content: string
    timestamp: Date
}

interface AIChatPanelProps {
    exerciseContext?: {
        title: string
        description: string
    }
    className?: string
}

const SUGGESTED_QUESTIONS = [
    'Gợi ý cách tiếp cận bài này?',
    'Giải thích lỗi trong câu SQL của tôi',
    'So sánh JOIN và LEFT JOIN',
    'Khi nào cần dùng GROUP BY?',
]

export function AIChatPanel({ exerciseContext, className }: AIChatPanelProps) {
    const [isOpen, setIsOpen] = useState(false)
    const [isExpanded, setIsExpanded] = useState(false)
    const [messages, setMessages] = useState<ChatMessage[]>([
        {
            id: '1',
            role: 'assistant',
            content: 'Xin chào! 👋 Tôi là trợ lý AI của bạn. Tôi có thể giúp bạn:\n\n• Giải thích các khái niệm SQL\n• Gợi ý cách tiếp cận bài tập\n• Debug lỗi trong câu truy vấn\n• Trả lời các câu hỏi về database\n\nHãy hỏi tôi bất cứ điều gì!',
            timestamp: new Date(),
        },
    ])
    const [inputValue, setInputValue] = useState('')
    const [isLoading, setIsLoading] = useState(false)
    const messagesEndRef = useRef<HTMLDivElement>(null)
    const inputRef = useRef<HTMLInputElement>(null)

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }

    useEffect(() => {
        scrollToBottom()
    }, [messages])

    useEffect(() => {
        if (isOpen && inputRef.current) {
            inputRef.current.focus()
        }
    }, [isOpen])

    const handleSend = async () => {
        if (!inputValue.trim() || isLoading) return

        const userMessage: ChatMessage = {
            id: Date.now().toString(),
            role: 'user',
            content: inputValue,
            timestamp: new Date(),
        }

        setMessages((prev) => [...prev, userMessage])
        setInputValue('')
        setIsLoading(true)

        // Simulate AI response (replace with real API later)
        setTimeout(() => {
            const aiResponses = [
                'Đây là một câu hỏi tuyệt vời! Để giải quyết bài toán này, bạn nên bắt đầu bằng việc xác định các bảng cần sử dụng và mối quan hệ giữa chúng.',
                'Dựa trên câu hỏi của bạn, tôi nghĩ bạn nên xem xét việc sử dụng JOIN để kết hợp dữ liệu từ nhiều bảng.',
                'Lỗi phổ biến trong trường hợp này thường liên quan đến việc quên GROUP BY khi sử dụng aggregate functions. Hãy kiểm tra lại xem bạn đã bao gồm tất cả các cột không-aggregate trong GROUP BY chưa.',
                'Để tối ưu performance, bạn nên:\n\n1. Chỉ SELECT các cột cần thiết\n2. Sử dụng index phù hợp\n3. Tránh sử dụng SELECT * trong production',
            ]

            const aiMessage: ChatMessage = {
                id: (Date.now() + 1).toString(),
                role: 'assistant',
                content: aiResponses[Math.floor(Math.random() * aiResponses.length)],
                timestamp: new Date(),
            }

            setMessages((prev) => [...prev, aiMessage])
            setIsLoading(false)
        }, 1000 + Math.random() * 1000)
    }

    const handleSuggestedQuestion = (question: string) => {
        setInputValue(question)
        inputRef.current?.focus()
    }

    if (!isOpen) {
        return (
            <Button
                onClick={() => setIsOpen(true)}
                className="fixed bottom-6 right-6 h-14 w-14 rounded-full shadow-lg z-50 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
                size="icon"
            >
                <MessageCircle className="h-6 w-6" />
            </Button>
        )
    }

    return (
        <Card
            className={cn(
                'fixed z-50 flex flex-col shadow-2xl border-0 transition-all duration-300',
                isExpanded
                    ? 'bottom-0 right-0 w-full h-full md:w-[500px] md:h-[90vh] md:bottom-4 md:right-4 rounded-none md:rounded-2xl'
                    : 'bottom-4 right-4 w-[380px] h-[550px] rounded-2xl',
                className
            )}
        >
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-3 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-t-2xl">
                <div className="flex items-center gap-2">
                    <div className="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center">
                        <Sparkles className="h-4 w-4" />
                    </div>
                    <div>
                        <h3 className="font-semibold text-sm">Trợ lý AI SQL</h3>
                        <p className="text-xs text-white/70">Đang phát triển</p>
                    </div>
                </div>
                <div className="flex items-center gap-1">
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-white hover:bg-white/20"
                        onClick={() => setIsExpanded(!isExpanded)}
                    >
                        {isExpanded ? (
                            <Minimize2 className="h-4 w-4" />
                        ) : (
                            <Maximize2 className="h-4 w-4" />
                        )}
                    </Button>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-white hover:bg-white/20"
                        onClick={() => setIsOpen(false)}
                    >
                        <X className="h-4 w-4" />
                    </Button>
                </div>
            </div>

            {/* Exercise Context (if provided) */}
            {exerciseContext && (
                <div className="px-4 py-2 bg-muted/50 border-b text-xs">
                    <p className="text-muted-foreground">
                        Đang tham chiếu: <span className="font-medium text-foreground">{exerciseContext.title}</span>
                    </p>
                </div>
            )}

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {messages.map((message) => (
                    <div
                        key={message.id}
                        className={cn(
                            'flex gap-2',
                            message.role === 'user' ? 'justify-end' : 'justify-start'
                        )}
                    >
                        {message.role === 'assistant' && (
                            <div className="w-7 h-7 rounded-full bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center flex-shrink-0">
                                <Bot className="h-4 w-4 text-white" />
                            </div>
                        )}
                        <div
                            className={cn(
                                'max-w-[80%] rounded-2xl px-4 py-2 text-sm',
                                message.role === 'user'
                                    ? 'bg-primary text-primary-foreground rounded-br-md'
                                    : 'bg-muted rounded-bl-md'
                            )}
                        >
                            <p className="whitespace-pre-wrap">{message.content}</p>
                            <p className="text-[10px] mt-1 opacity-50">
                                {message.timestamp.toLocaleTimeString('vi-VN', {
                                    hour: '2-digit',
                                    minute: '2-digit',
                                })}
                            </p>
                        </div>
                        {message.role === 'user' && (
                            <div className="w-7 h-7 rounded-full bg-primary flex items-center justify-center flex-shrink-0">
                                <User className="h-4 w-4 text-primary-foreground" />
                            </div>
                        )}
                    </div>
                ))}

                {isLoading && (
                    <div className="flex gap-2 justify-start">
                        <div className="w-7 h-7 rounded-full bg-gradient-to-r from-blue-500 to-purple-500 flex items-center justify-center">
                            <Bot className="h-4 w-4 text-white" />
                        </div>
                        <div className="bg-muted rounded-2xl rounded-bl-md px-4 py-3">
                            <div className="flex gap-1">
                                <span className="w-2 h-2 bg-muted-foreground/50 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                                <span className="w-2 h-2 bg-muted-foreground/50 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                                <span className="w-2 h-2 bg-muted-foreground/50 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                            </div>
                        </div>
                    </div>
                )}

                <div ref={messagesEndRef} />
            </div>

            {/* Suggested Questions */}
            {messages.length <= 2 && (
                <div className="px-4 pb-2">
                    <p className="text-xs text-muted-foreground mb-2">Gợi ý câu hỏi:</p>
                    <div className="flex flex-wrap gap-2">
                        {SUGGESTED_QUESTIONS.map((question, index) => (
                            <button
                                key={index}
                                onClick={() => handleSuggestedQuestion(question)}
                                className="text-xs px-3 py-1.5 rounded-full bg-muted hover:bg-muted/80 transition-colors"
                            >
                                {question}
                            </button>
                        ))}
                    </div>
                </div>
            )}

            {/* Input */}
            <div className="p-4 border-t">
                <form
                    onSubmit={(e) => {
                        e.preventDefault()
                        handleSend()
                    }}
                    className="flex gap-2"
                >
                    <Input
                        ref={inputRef}
                        value={inputValue}
                        onChange={(e) => setInputValue(e.target.value)}
                        placeholder="Nhập câu hỏi của bạn..."
                        disabled={isLoading}
                        className="flex-1 rounded-full"
                    />
                    <Button
                        type="submit"
                        size="icon"
                        disabled={!inputValue.trim() || isLoading}
                        className="rounded-full h-10 w-10"
                    >
                        {isLoading ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                            <Send className="h-4 w-4" />
                        )}
                    </Button>
                </form>
                <p className="text-[10px] text-center text-muted-foreground mt-2">
                    ⚠️ Tính năng AI đang phát triển, câu trả lời có thể chưa chính xác
                </p>
            </div>
        </Card>
    )
}
