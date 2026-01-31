export type WebSocketMessage = {
    type: string
    data: any
    timestamp?: number
}

type WebSocketCallbacks = {
    onMessage?: (message: WebSocketMessage) => void
    onError?: (error: Event) => void
    onClose?: () => void
    onOpen?: () => void
}

class WebSocketService {
    private ws: WebSocket | null = null
    private url: string
    private reconnectAttempts = 0
    private maxReconnectAttempts = 5
    private reconnectDelay = 1000
    private callbacks: WebSocketCallbacks = {}
    private token: string | null = null

    constructor(url: string) {
        this.url = url
    }

    connect(token?: string) {
        if (token) {
            this.token = token
        }

        if (this.ws?.readyState === WebSocket.OPEN) {
            console.log('WebSocket already connected')
            return
        }

        try {
            const wsUrl = this.token
                ? `${this.url}?token=${this.token}`
                : this.url

            this.ws = new WebSocket(wsUrl)

            this.ws.onopen = () => {
                console.log('WebSocket connected')
                this.reconnectAttempts = 0
                this.callbacks.onOpen?.()
            }

            this.ws.onmessage = (event) => {
                try {
                    const message: WebSocketMessage = JSON.parse(event.data)
                    this.callbacks.onMessage?.(message)
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error)
                }
            }

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error)
                this.callbacks.onError?.(error)
            }

            this.ws.onclose = () => {
                console.log('WebSocket disconnected')
                this.callbacks.onClose?.()
                this.attemptReconnect()
            }
        } catch (error) {
            console.error('Error connecting WebSocket:', error)
        }
    }

    private attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++
            setTimeout(() => {
                console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`)
                this.connect()
            }, this.reconnectDelay * this.reconnectAttempts)
        } else {
            console.error('Max reconnection attempts reached')
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close()
            this.ws = null
        }
    }

    send(message: WebSocketMessage) {
        if (this.ws?.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message))
        } else {
            console.warn('WebSocket is not connected')
        }
    }

    setCallbacks(callbacks: WebSocketCallbacks) {
        this.callbacks = { ...this.callbacks, ...callbacks }
    }

    isConnected(): boolean {
        return this.ws?.readyState === WebSocket.OPEN
    }
}

// Singleton instance
const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
export const websocketService = new WebSocketService(wsUrl)

