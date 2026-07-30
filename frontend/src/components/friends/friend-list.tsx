import { MessageCircle, UserMinus } from 'lucide-react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { getErrorMessage } from '@/lib/api'
import { fetchConversations } from '@/lib/chat-api'
import { removeFriend, type FriendItem } from '@/lib/friend-api'

interface Props {
  friends: FriendItem[]
  onUpdate: () => void
}

export function FriendList({ friends, onUpdate }: Props) {
  const navigate = useNavigate()
  const [removing, setRemoving] = useState<number | null>(null)

  async function handleRemove(friendshipId: number) {
    if (!confirm('确定要删除该好友吗？')) return
    setRemoving(friendshipId)
    try {
      await removeFriend(friendshipId)
      onUpdate()
    } catch (err) {
      alert(getErrorMessage(err))
    } finally {
      setRemoving(null)
    }
  }

  async function handleChat(userId: number) {
    try {
      const conversations = await fetchConversations()
      const conv = conversations.find((c) => c.peerUserId === userId)
      if (conv) {
        navigate(`/chat/${conv.id}`)
      } else {
        navigate('/chat')
      }
    } catch {
      navigate('/chat')
    }
  }

  if (friends.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-[#7e6d61]">
        还没有好友，快去添加吧
      </p>
    )
  }

  return (
    <div className="space-y-3">
      {friends.map((friend) => (
        <div
          className="flex items-center justify-between rounded-2xl border border-[#eee4db] bg-white/80 px-4 py-3"
          key={friend.friendshipId}
        >
          <div>
            <p className="font-medium text-[#2f221b]">
              {friend.user.nickname}
            </p>
            <p className="text-xs text-[#7e6d61]">
              {friend.user.department} · {friend.user.grade}
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              onClick={() => handleChat(friend.user.userId)}
              size="sm"
              variant="secondary"
            >
              <MessageCircle className="mr-1 h-3.5 w-3.5" />
              聊天
            </Button>
            <Button
              disabled={removing === friend.friendshipId}
              onClick={() => handleRemove(friend.friendshipId)}
              size="sm"
              variant="ghost"
            >
              <UserMinus className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      ))}
    </div>
  )
}
