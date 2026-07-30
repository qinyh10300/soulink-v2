import { UserPlus } from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { getErrorMessage } from '@/lib/api'
import { sendFriendRequest } from '@/lib/friend-api'

interface Props {
  open: boolean
  onClose: () => void
  onSent: () => void
}

export function AddFriendModal({ open, onClose, onSent }: Props) {
  const [nickname, setNickname] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')

  if (!open) return null

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmedNickname = nickname.trim()
    if (!trimmedNickname) {
      setMessage('请输入对方昵称')
      return
    }
    setLoading(true)
    setMessage('')
    try {
      await sendFriendRequest(trimmedNickname)
      setMessage('好友请求已发送')
      setNickname('')
      onSent()
    } catch (err) {
      setMessage(getErrorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm">
      <div className="mx-4 w-full max-w-md rounded-3xl border border-[#eee4db] bg-[rgba(255,251,247,0.98)] p-6 shadow-[0_18px_54px_rgba(72,50,33,0.12)]">
        <h2 className="mb-4 text-lg font-semibold text-[#2f221b]">
          添加好友
        </h2>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div>
            <label
              className="mb-1 block text-sm text-[#7e6d61]"
              htmlFor="friend-nickname"
            >
              对方昵称
            </label>
            <Input
              id="friend-nickname"
              onChange={(e) => setNickname(e.target.value)}
              placeholder="输入对方昵称"
              value={nickname}
            />
          </div>
          {message ? (
            <p className="text-sm text-[#7e6d61]">{message}</p>
          ) : null}
          <div className="flex justify-end gap-3">
            <Button
              onClick={onClose}
              type="button"
              variant="ghost"
            >
              取消
            </Button>
            <Button disabled={loading} type="submit">
              <UserPlus className="mr-1 h-4 w-4" />
              发送请求
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
