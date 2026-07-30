import { ArrowLeft } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'

import { ChatWindow } from '@/components/chat/chat-window'
import { ConversationList } from '@/components/chat/conversation-list'
import { useChat } from '@/contexts/chat-context'
import { fetchConversations, type ConversationItem } from '@/lib/chat-api'

export function ChatPage() {
  const { conversationId } = useParams<{ conversationId?: string }>()
  const navigate = useNavigate()
  const { lastMessage, conversationVersion } = useChat()
  const [conversations, setConversations] = useState<ConversationItem[]>([])

  const activeId = conversationId ? parseInt(conversationId, 10) : undefined
  const activeConversation = conversations.find((c) => c.id === activeId)

  const loadConversations = useCallback(() => {
    fetchConversations().then(setConversations).catch(() => {})
  }, [])

  useEffect(() => {
    loadConversations()
  }, [loadConversations])

  useEffect(() => {
    if (lastMessage) {
      loadConversations()
    }
  }, [lastMessage, loadConversations])

  useEffect(() => {
    loadConversations()
  }, [conversationVersion, loadConversations])

  function handleSelect(id: number) {
    navigate(`/chat/${id}`)
  }

  const showList = !activeId
  const showWindow = !!activeConversation

  return (
    <div className="flex h-screen bg-[#f7f2ec]">
      <div
        className={`flex w-full flex-col border-r border-[#eee4db] bg-[rgba(255,251,247,0.96)] md:flex md:w-80 ${
          showList ? 'flex' : 'hidden'
        }`}
      >
        <div className="flex shrink-0 items-center gap-3 border-b border-[#eee4db] px-4 py-4">
          <Link
            className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#eee4db] bg-white/80 text-[#5f4d42] transition-colors hover:bg-white"
            to="/dashboard"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <h1 className="text-lg font-semibold text-[#2f221b]">消息</h1>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto p-2">
          <ConversationList
            activeId={activeId}
            conversations={conversations}
            onSelect={handleSelect}
          />
        </div>
      </div>

      <div
        className={`flex-1 ${showWindow ? 'flex flex-col' : 'hidden md:flex md:flex-col'}`}
      >
        {activeConversation ? (
          <>
            <div className="md:hidden">
              <button
                className="flex items-center gap-2 border-b border-[#eee4db] px-4 py-3 text-sm text-[#5f4d42]"
                onClick={() => navigate('/chat')}
                type="button"
              >
                <ArrowLeft className="h-4 w-4" />
                返回列表
              </button>
            </div>
            <div className="min-h-0 flex-1">
              <ChatWindow conversation={activeConversation} onRead={loadConversations} />
            </div>
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center">
            <p className="text-sm text-[#7e6d61]">选择一个会话开始聊天</p>
          </div>
        )}
      </div>
    </div>
  )
}
