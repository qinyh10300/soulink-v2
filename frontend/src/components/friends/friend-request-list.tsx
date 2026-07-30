import { Check, X } from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'
import { getErrorMessage } from '@/lib/api'
import {
  acceptFriendRequest,
  rejectFriendRequest,
  type FriendRequest,
} from '@/lib/friend-api'

interface Props {
  requests: FriendRequest[]
  onUpdate: () => void
}

export function FriendRequestList({ requests, onUpdate }: Props) {
  const [loading, setLoading] = useState<number | null>(null)

  async function handleAccept(friendshipId: number) {
    setLoading(friendshipId)
    try {
      await acceptFriendRequest(friendshipId)
      onUpdate()
    } catch (err) {
      alert(getErrorMessage(err))
    } finally {
      setLoading(null)
    }
  }

  async function handleReject(friendshipId: number) {
    setLoading(friendshipId)
    try {
      await rejectFriendRequest(friendshipId)
      onUpdate()
    } catch (err) {
      alert(getErrorMessage(err))
    } finally {
      setLoading(null)
    }
  }

  if (requests.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-[#7e6d61]">
        暂无好友请求
      </p>
    )
  }

  return (
    <div className="space-y-3">
      {requests.map((req) => (
        <div
          className="flex items-center justify-between rounded-2xl border border-[#eee4db] bg-white/80 px-4 py-3"
          key={req.friendshipId}
        >
          <div>
            <p className="font-medium text-[#2f221b]">
              {req.user.nickname}
            </p>
            <p className="text-xs text-[#7e6d61]">
              {req.user.department} · {req.user.grade}
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              disabled={loading === req.friendshipId}
              onClick={() => handleAccept(req.friendshipId)}
              size="sm"
              variant="default"
            >
              <Check className="mr-1 h-3.5 w-3.5" />
              接受
            </Button>
            <Button
              disabled={loading === req.friendshipId}
              onClick={() => handleReject(req.friendshipId)}
              size="sm"
              variant="ghost"
            >
              <X className="mr-1 h-3.5 w-3.5" />
              拒绝
            </Button>
          </div>
        </div>
      ))}
    </div>
  )
}
