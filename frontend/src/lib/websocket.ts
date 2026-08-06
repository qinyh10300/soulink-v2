import { getAuthSession } from '@/lib/user-profile-storage'

export interface WSEnvelope {
  type: string
  data?: unknown
}

export type WSEventHandler = (envelope: WSEnvelope) => void
export type WSConnectionState = 'connecting' | 'connected' | 'reconnecting' | 'disconnected'
export type WSConnectionHandler = (state: WSConnectionState) => void

const RECONNECT_BASE_MS = 1000
const RECONNECT_MAX_MS = 30000

export class SoulinkWebSocket {
  private ws: WebSocket | null = null
  private handlers: WSEventHandler[] = []
  private connectionHandlers: WSConnectionHandler[] = []
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectDelay = RECONNECT_BASE_MS
  private intentionallyClosed = false
  private connectionState: WSConnectionState = 'disconnected'

  connect() {
    this.intentionallyClosed = false
    const session = getAuthSession()
    if (!session?.token) return
    if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
      return
    }

    this.setConnectionState(
      this.reconnectDelay > RECONNECT_BASE_MS ? 'reconnecting' : 'connecting',
    )

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/api/v1/ws?token=${encodeURIComponent(session.token)}`

    try {
      this.ws = new WebSocket(url)
    } catch {
      this.setConnectionState('reconnecting')
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      this.reconnectDelay = RECONNECT_BASE_MS
      this.setConnectionState('connected')
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
      this.ws = null
      if (!this.intentionallyClosed) {
        this.setConnectionState('reconnecting')
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
    this.setConnectionState('disconnected')
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

  sendGroupMessage(groupId: number, content: string) {
    this.send({
      type: 'group_message',
      data: { groupId, content, contentType: 'text' },
    })
  }

  sendRead(conversationId: number) {
    this.send({ type: 'read', data: { conversationId } })
  }

  sendGroupRead(groupId: number) {
    this.send({ type: 'group_read', data: { groupId } })
  }

  onEvent(handler: WSEventHandler) {
    this.handlers.push(handler)
    return () => {
      this.handlers = this.handlers.filter((h) => h !== handler)
    }
  }

  onConnectionState(handler: WSConnectionHandler) {
    this.connectionHandlers.push(handler)
    handler(this.connectionState)
    return () => {
      this.connectionHandlers = this.connectionHandlers.filter((item) => item !== handler)
    }
  }

  get isConnected() {
    return this.ws?.readyState === WebSocket.OPEN
  }

  private scheduleReconnect() {
    if (this.intentionallyClosed || this.reconnectTimer) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, this.reconnectDelay)
    this.reconnectDelay = Math.min(
      this.reconnectDelay * 2,
      RECONNECT_MAX_MS,
    )
  }

  private setConnectionState(state: WSConnectionState) {
    if (this.connectionState === state) return
    this.connectionState = state
    for (const handler of this.connectionHandlers) {
      handler(state)
    }
  }
}
