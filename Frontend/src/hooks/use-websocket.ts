import { useEffect, useRef } from 'react'

import { websocketService } from '@/services/websocket.service'
import { useAuthStore } from '@/stores/use-auth-store'
import type { WebSocketMessage } from '@/services/websocket.service'

interface UseWebSocketOptions {
    onMessage?: (message: WebSocketMessage) => void
    onError?: (error: Event) => void
    onClose?: () => void
    onOpen?: () => void
    autoConnect?: boolean
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
    const { onMessage, onError, onClose, onOpen, autoConnect = true } = options
    const token = useAuthStore((state) => state.token)
    const callbacksRef = useRef({ onMessage, onError, onClose, onOpen })

    useEffect(() => {
        callbacksRef.current = { onMessage, onError, onClose, onOpen }
    }, [onMessage, onError, onClose, onOpen])

    useEffect(() => {
        if (!autoConnect) return

        websocketService.setCallbacks({
            onMessage: (message) => callbacksRef.current.onMessage?.(message),
            onError: (error) => callbacksRef.current.onError?.(error),
            onClose: () => callbacksRef.current.onClose?.(),
            onOpen: () => callbacksRef.current.onOpen?.(),
        })

        if (token) {
            websocketService.connect(token)
        }

        return () => {
            websocketService.disconnect()
        }
    }, [token, autoConnect])

    const send = (message: WebSocketMessage) => {
        websocketService.send(message)
    }

    const connect = () => {
        if (token) {
            websocketService.connect(token)
        }
    }

    const disconnect = () => {
        websocketService.disconnect()
    }

    return {
        send,
        connect,
        disconnect,
        isConnected: websocketService.isConnected(),
    }
}

