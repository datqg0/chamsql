import { useState } from 'react'
import toast from 'react-hot-toast'
import { useQuery } from '@tanstack/react-query'

import { Button } from '@/components/ui/button'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog'
import { Checkbox } from '@/components/ui/checkbox'
import { examsService } from '@/services/exams.service'
import { userService } from '@/services/user.service'
import type { User } from '@/types/auth.types'
import type { AddParticipantsRequest } from '@/types/exam.types'
import { Plus, Loader2, Search } from 'lucide-react'
import { Input } from '@/components/ui/input'

interface AddParticipantsDialogProps {
    examId: number
    onSuccess?: () => void
}

export function AddParticipantsDialog({ examId, onSuccess }: AddParticipantsDialogProps) {
    const [open, setOpen] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [selectedUserIds, setSelectedUserIds] = useState<number[]>([])
    const [searchQuery, setSearchQuery] = useState('')

    // Fetch students (users with role=student)
    const { data: students = [], isLoading } = useQuery<User[]>({
        queryKey: ['students'],
        queryFn: async () => {
            const response = await userService.getList()
            // Filter for students only
            return response.data.users.filter((user: User) => user.role === 'student')
        },
        enabled: open,
    })

    const filteredStudents = students.filter(
        (student: User) =>
            student.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
            student.email?.toLowerCase().includes(searchQuery.toLowerCase())
    )

    const handleToggleStudent = (userId: number) => {
        setSelectedUserIds((prev) =>
            prev.includes(userId)
                ? prev.filter((id) => id !== userId)
                : [...prev, userId]
        )
    }

    const handleSubmit = async () => {
        if (selectedUserIds.length === 0) {
            toast.error('Vui lòng chọn ít nhất 1 sinh viên!')
            return
        }

        setIsSubmitting(true)

        try {
            const request: AddParticipantsRequest = {
                userIds: selectedUserIds,
            }

            await examsService.addParticipants(examId, request)
            toast.success(`Đã thêm ${selectedUserIds.length} sinh viên vào kỳ thi!`)
            setOpen(false)
            setSelectedUserIds([])
            setSearchQuery('')
            onSuccess?.()
        } catch (error: any) {
            toast.error(error?.message || 'Thêm sinh viên thất bại!')
        } finally {
            setIsSubmitting(false)
        }
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button size="sm">
                    <Plus className="h-4 w-4 mr-1" />
                    Thêm
                </Button>
            </DialogTrigger>
            <DialogContent className="max-w-lg max-h-[600px] flex flex-col">
                <DialogHeader>
                    <DialogTitle>Thêm sinh viên vào kỳ thi</DialogTitle>
                    <DialogDescription>
                        Chọn sinh viên tham gia kỳ thi này
                    </DialogDescription>
                </DialogHeader>

                <div className="flex-1 overflow-hidden flex flex-col gap-4">
                    {/* Search */}
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <Input
                            placeholder="Tìm theo tên hoặc email..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="pl-9"
                        />
                    </div>

                    {/* Students List */}
                    <div className="flex-1 overflow-y-auto border rounded-lg">
                        {isLoading ? (
                            <div className="flex items-center justify-center py-8">
                                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                            </div>
                        ) : filteredStudents.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground">
                                {searchQuery ? 'Không tìm thấy sinh viên' : 'Không có sinh viên nào'}
                            </div>
                        ) : (
                            <div className="p-2 space-y-1">
                                {filteredStudents.map((student: User) => (
                                    <div
                                        key={student.id}
                                        className="flex items-center gap-3 p-3 rounded-lg hover:bg-muted/50 cursor-pointer transition-colors"
                                        onClick={() => handleToggleStudent(student.id)}
                                    >
                                        <Checkbox
                                            checked={selectedUserIds.includes(student.id)}
                                            onCheckedChange={() => handleToggleStudent(student.id)}
                                            onClick={(e) => e.stopPropagation()}
                                        />
                                        <div className="flex-1 min-w-0">
                                            <p className="font-medium truncate">{student.username}</p>
                                            {student.email && (
                                                <p className="text-sm text-muted-foreground truncate">
                                                    {student.email}
                                                </p>
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>

                    {/* Selected count */}
                    {selectedUserIds.length > 0 && (
                        <div className="text-sm text-muted-foreground">
                            Đã chọn: <span className="font-medium text-foreground">{selectedUserIds.length}</span> sinh viên
                        </div>
                    )}
                </div>

                <DialogFooter>
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            setOpen(false)
                            setSelectedUserIds([])
                            setSearchQuery('')
                        }}
                        disabled={isSubmitting}
                    >
                        Hủy
                    </Button>
                    <Button
                        onClick={handleSubmit}
                        disabled={isSubmitting || selectedUserIds.length === 0}
                    >
                        {isSubmitting ? (
                            <>
                                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                Đang thêm...
                            </>
                        ) : (
                            `Thêm ${selectedUserIds.length > 0 ? `(${selectedUserIds.length})` : ''}`
                        )}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
