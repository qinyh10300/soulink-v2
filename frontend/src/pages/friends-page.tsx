import { ArrowLeft, UserPlus } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { AddFriendModal } from '@/components/friends/add-friend-modal'
import { FriendList } from '@/components/friends/friend-list'
import { FriendRequestList } from '@/components/friends/friend-request-list'
import { Button } from '@/components/ui/button'
import { useChat } from '@/contexts/chat-context'
import {
  fetchFriends,
  fetchReceivedRequests,
  type FriendItem,
  type FriendRequest,
} from '@/lib/friend-api'

type Tab = 'friends' | 'received' | 'sent'

export function FriendsPage() {
  const { refreshPendingRequests } = useChat()
  const [tab, setTab] = useState<Tab>('friends')
  const [friends, setFriends] = useState<FriendItem[]>([])
  const [received, setReceived] = useState<FriendRequest[]>([])
  const [addModalOpen, setAddModalOpen] = useState(false)

  const loadData = useCallback(() => {
    fetchFriends().then(setFriends).catch(() => {})
    fetchReceivedRequests()
      .then((reqs) => {
        setReceived(reqs)
        refreshPendingRequests()
      })
      .catch(() => {})
  }, [refreshPendingRequests])

  useEffect(() => {
    loadData()
  }, [loadData])

  const tabs: { key: Tab; label: string }[] = [
    { key: 'friends', label: '好友列表' },
    { key: 'received', label: `好友请求${received.length > 0 ? ` (${received.length})` : ''}` },
  ]

  return (
    <div className="min-h-screen bg-[#f7f2ec] px-4 pb-12 pt-6 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-2xl">
        <div className="mb-6 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link
              className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#eee4db] bg-white/80 text-[#5f4d42] transition-colors hover:bg-white"
              to="/dashboard"
            >
              <ArrowLeft className="h-4 w-4" />
            </Link>
            <h1 className="text-xl font-semibold text-[#2f221b]">好友</h1>
          </div>
          <Button onClick={() => setAddModalOpen(true)} size="sm">
            <UserPlus className="mr-1 h-4 w-4" />
            添加好友
          </Button>
        </div>

        <div className="mb-4 flex gap-2">
          {tabs.map((t) => (
            <button
              className={`rounded-full px-4 py-2 text-sm font-medium transition-colors ${
                tab === t.key
                  ? 'bg-[#2f221b] text-white'
                  : 'bg-white/80 text-[#5f4d42] hover:bg-white'
              }`}
              key={t.key}
              onClick={() => setTab(t.key)}
              type="button"
            >
              {t.label}
            </button>
          ))}
        </div>

        <div className="rounded-3xl border border-[#eee4db] bg-[rgba(255,251,247,0.96)] p-5 shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
          {tab === 'friends' ? (
            <FriendList friends={friends} onUpdate={loadData} />
          ) : (
            <FriendRequestList
              onUpdate={loadData}
              requests={received}
            />
          )}
        </div>
      </div>

      <AddFriendModal
        onClose={() => setAddModalOpen(false)}
        onSent={loadData}
        open={addModalOpen}
      />
    </div>
  )
}
