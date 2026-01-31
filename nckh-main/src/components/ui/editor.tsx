import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import { Bold, Italic, Underline as UnderlineIcon, List, ListOrdered, Heading1, Heading2, Quote, Undo, Redo } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface EditorProps {
    value?: string
    onChange?: (content: string) => void
    placeholder?: string
    className?: string
}

export function Editor({ value, onChange, placeholder, className }: EditorProps) {
    const editor = useEditor({
        extensions: [
            StarterKit,
            Underline,
        ],
        content: value,
        onUpdate: ({ editor }) => {
            onChange?.(editor.getHTML())
        },
        editorProps: {
            attributes: {
                class: 'prose prose-sm dark:prose-invert max-w-none min-h-[150px] p-3 focus:outline-none',
            },
        },
    })

    if (!editor) {
        return null
    }

    const ToolbarButton = ({
        isActive,
        onClick,
        icon: Icon,
        label
    }: {
        isActive?: boolean
        onClick: () => void
        icon: any
        label: string
    }) => (
        <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onClick}
            className={cn("h-8 w-8 p-0", isActive && "bg-muted text-item-foreground")}
            title={label}
        >
            <Icon className="h-4 w-4" />
        </Button>
    )

    return (
        <div className={cn("border rounded-md overflow-hidden bg-background", className)}>
            <div className="border-b p-1 flex flex-wrap gap-1 bg-muted/50">
                <ToolbarButton
                    isActive={editor.isActive('bold')}
                    onClick={() => editor.chain().focus().toggleBold().run()}
                    icon={Bold}
                    label="Bold"
                />
                <ToolbarButton
                    isActive={editor.isActive('italic')}
                    onClick={() => editor.chain().focus().toggleItalic().run()}
                    icon={Italic}
                    label="Italic"
                />
                <ToolbarButton
                    isActive={editor.isActive('underline')}
                    onClick={() => editor.chain().focus().toggleUnderline().run()}
                    icon={UnderlineIcon}
                    label="Underline"
                />
                <div className="w-px h-6 bg-border mx-1 my-auto" />
                <ToolbarButton
                    isActive={editor.isActive('heading', { level: 1 })}
                    onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}
                    icon={Heading1}
                    label="Heading 1"
                />
                <ToolbarButton
                    isActive={editor.isActive('heading', { level: 2 })}
                    onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
                    icon={Heading2}
                    label="Heading 2"
                />
                <div className="w-px h-6 bg-border mx-1 my-auto" />
                <ToolbarButton
                    isActive={editor.isActive('bulletList')}
                    onClick={() => editor.chain().focus().toggleBulletList().run()}
                    icon={List}
                    label="Bullet List"
                />
                <ToolbarButton
                    isActive={editor.isActive('orderedList')}
                    onClick={() => editor.chain().focus().toggleOrderedList().run()}
                    icon={ListOrdered}
                    label="Ordered List"
                />
                <div className="w-px h-6 bg-border mx-1 my-auto" />
                <ToolbarButton
                    isActive={editor.isActive('blockquote')}
                    onClick={() => editor.chain().focus().toggleBlockquote().run()}
                    icon={Quote}
                    label="Blockquote"
                />
                <div className="flex-1" />
                <ToolbarButton
                    onClick={() => editor.chain().focus().undo().run()}
                    icon={Undo}
                    label="Undo"
                />
                <ToolbarButton
                    onClick={() => editor.chain().focus().redo().run()}
                    icon={Redo}
                    label="Redo"
                />
            </div>
            <EditorContent editor={editor} className="min-h-[150px]" />
        </div>
    )
}
