import { getAuthSession } from '@/lib/user-profile-storage'

export interface WSEnvelope {
  type: string
  data?: unknown
}

export type WSEventHandler = (envelope: WSEnvelope) => void

const RECONNECT_BASE_MS = 1000
const RECONNECT_MAX_MS = 30000

export class SoulinkWebSocket {
  private ws: WebSocket | null = null
  private handlers: WSEventHandler[] = []
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectDelay = RECONNECT_BASE_MS
  private intentionallyClosed = false

  connect() {
    this.intentionallyClosed = false
    const session = getAuthSession()
    if (!session?.token) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/api/v1/ws?token=${encodeURIComponent(session.token)}`

    try {
      this.ws = new WebSocket(url)
    } catch {
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      this.reconnectDelay = RECONNECT_BASE_MS
    }

    this.ws.onmessage = (event) => {
      try {
        const envelope = JSON.parse(event.data as string) as WSEnvelope
        for (const handler of this.handlers) {
          handler(envelope)
        }
      } catch {
        // ignore malformed messages
      }
    }

    this.ws.onclose = () => {
      if (!this.intentionallyClosed) {
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = () => {
      this.ws?.close()
    }
  }

  disconnect() {
    this.intentionallyClosed = true
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.ws?.close()
    this.ws = null
  }

  send(envelope: WSEnvelope) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(envelope))
    }
  }

  sendMessage(conversationId: number, content: string) {
    this.send({
      type: 'message',
      data: { conversationId, content, contentType: 'text' },
    })
  }

  sendRead(conversationId: number) {
    this.send({ type: 'read', data: { conversationId } })
  }

  onEvent(handler: WSEventHandler) {
    this.handlers.push(handler)
    return () => {
      this.handlers = this.handlers.filter((h) => h !== handler)
    }
  }

  get isConnected() {
    return this.ws?.readyState === WebSocket.OPEN
  }

  private scheduleReconnect() {
    if (this.intentionallyClosed) return
    this.reconnectTimer = setTimeout(() => {
      this.connect()
    }, this.reconnectDelay)
    this.reconnectDelay = Math.min(
      this.reconnectDelay * 2,
      RECONNECT_MAX_MS,
    )
  }
}
