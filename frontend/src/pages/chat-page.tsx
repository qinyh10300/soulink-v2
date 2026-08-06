import {
  ArrowLeft,
  Check,
  MessageSquarePlus,
  Search,
  Trash2,
  UserPlus,
  Users,
  X,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { ChatWindow } from '@/components/chat/chat-window'
import {
  ConversationList,
  type ChatThreadItem,
} from '@/components/chat/conversation-list'
import { useChat } from '@/contexts/chat-context'
import { getErrorMessage } from '@/lib/api'
import {
  addGroupMembers,
  createGroup,
  ensureConversation,
  fetchConversations,
  fetchGroup,
  fetchGroups,
  removeGroupMember,
  renameGroup,
  type ConversationItem,
  type GroupItem,
} from '@/lib/chat-api'
import { fetchFriends, type FriendItem } from '@/lib/friend-api'
import { cn } from '@/lib/utils'

interface NewGroupDialogProps {
  friends: FriendItem[]
  onClose: () => void
  onCreated: (group: GroupItem) => void
}

function NewGroupDialog({ friends, onClose, onCreated }: NewGroupDialogProps) {
  const [name, setName] = useState('')
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [submitting, setSubmitting] = useState(false)
  const [notice, setNotice] = useState('')

  const handleCreate = async () => {
    if (!name.trim() || selected.size < 2) {
      setNotice('填写群名称并选择至少两位好友')
      return
    }
    try {
      setSubmitting(true)
      const group = await createGroup({ name: name.trim(), memberIds: [...selected] })
      onCreated(group)
    } catch (error) {
      setNotice(getErrorMessage(error, '群聊创建失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-60 flex items-center justify-center bg-black/45 p-4">
      <section className="w-full max-w-md rounded-lg bg-white shadow-2xl" role="dialog" aria-modal="true">
        <header className="flex h-14 items-center border-b border-[#e1e3e0] px-4">
          <h2 className="text-base font-semibold">发起群聊</h2>
          <button
            aria-label="关闭"
            className="ml-auto inline-flex h-8 w-8 items-center justify-center rounded hover:bg-[#eceeeb]"
            onClick={onClose}
            title="关闭"
            type="button"
          >
            <X className="h-4 w-4" />
          </button>
        </header>
        <div className="space-y-4 p-4">
          <label className="block">
            <span className="mb-1.5 block text-xs font-medium text-[#666a65]">群名称</span>
            <input
              className="h-10 w-full rounded border border-[#ced2cd] px-3 text-sm outline-none focus:border-[#4d8960]"
              maxLength={40}
              onChange={(event) => setName(event.target.value)}
              placeholder="例如：周末徒步搭子"
              value={name}
            />
          </label>
          <div>
            <div className="mb-2 flex items-center justify-between">
              <span className="text-xs font-medium text-[#666a65]">选择好友</span>
              <span className="text-xs text-[#8b8f89]">已选 {selected.size}</span>
            </div>
            <div className="max-h-64 overflow-y-auto rounded border border-[#e0e2df]">
              {friends.map((friend) => {
                const checked = selected.has(friend.user.userId)
                return (
                  <button
                    className="flex w-full items-center gap-3 border-b border-[#eceeeb] px-3 py-2.5 text-left last:border-0 hover:bg-[#f4f5f3]"
                    key={friend.friendshipId}
                    onClick={() =>
                      setSelected((current) => {
                        const next = new Set(current)
                        if (checked) next.delete(friend.user.userId)
                        else next.add(friend.user.userId)
                        return next
                      })
                    }
                    type="button"
                  >
                    <span className="flex h-9 w-9 items-center justify-center rounded bg-[#dce8e1] text-xs font-semibold text-[#315d4f]">
                      {friend.user.nickname.slice(0, 1)}
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-sm font-medium">{friend.user.nickname}</span>
                      <span className="block truncate text-xs text-[#858984]">
                        {friend.user.department} · {friend.user.grade}
                      </span>
                    </span>
                    <span
                      className={cn(
                        'inline-flex h-5 w-5 items-center justify-center rounded border',
                        checked ? 'border-[#3f8053] bg-[#3f8053] text-white' : 'border-[#bfc3be]',
                      )}
                    >
                      {checked ? <Check className="h-3.5 w-3.5" /> : null}
                    </span>
                  </button>
                )
              })}
              {friends.length === 0 ? (
                <p className="px-4 py-8 text-center text-sm text-[#858984]">通讯录里还没有好友</p>
              ) : null}
            </div>
          </div>
          {notice ? <p className="text-sm text-[#a04b4b]">{notice}</p> : null}
        </div>
        <footer className="flex justify-end gap-2 border-t border-[#e1e3e0] px-4 py-3">
          <button className="h-9 rounded px-4 text-sm hover:bg-[#eceeeb]" onClick={onClose} type="button">
            取消
          </button>
          <button
            className="h-9 rounded bg-[#337a49] px-4 text-sm font-medium text-white hover:bg-[#28663c] disabled:bg-[#a9b9ad]"
            disabled={submitting}
            onClick={() => void handleCreate()}
            type="button"
          >
            {submitting ? '创建中' : '创建'}
          </button>
        </footer>
      </section>
    </div>
  )
}

interface GroupInfoPanelProps {
  friends: FriendItem[]
  groupId: number
  onChanged: () => void
  onClose: () => void
  onLeft: () => void
}

function GroupInfoPanel({ friends, groupId, onChanged, onClose, onLeft }: GroupInfoPanelProps) {
  const [group, setGroup] = useState<GroupItem | null>(null)
  const [name, setName] = useState('')
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [notice, setNotice] = useState('')

  const load = useCallback(async () => {
    try {
      const data = await fetchGroup(groupId)
      setGroup(data)
      setName(data.name)
    } catch (error) {
      setNotice(getErrorMessage(error, '群信息加载失败'))
    }
  }, [groupId])

  useEffect(() => {
    void load()
  }, [load])

  const eligibleFriends = useMemo(() => {
    const current = new Set(group?.members.map((member) => member.userId) ?? [])
    return friends.filter((friend) => !current.has(friend.user.userId))
  }, [friends, group])

  const handleRename = async () => {
    if (!group || name.trim() === group.name) return
    try {
      await renameGroup(groupId, name.trim())
      await load()
      onChanged()
    } catch (error) {
      setNotice(getErrorMessage(error, '群名称修改失败'))
    }
  }

  const handleAdd = async () => {
    if (selected.size === 0) return
    try {
      await addGroupMembers(groupId, [...selected])
      setSelected(new Set())
      await load()
      onChanged()
    } catch (error) {
      setNotice(getErrorMessage(error, '添加成员失败'))
    }
  }

  const handleRemove = async (userId: number, isMe: boolean) => {
    try {
      await removeGroupMember(groupId, userId)
      if (isMe) {
        onLeft()
        return
      }
      await load()
      onChanged()
    } catch (error) {
      setNotice(getErrorMessage(error, '移除成员失败'))
    }
  }

  return (
    <aside className="absolute inset-y-0 right-0 z-30 flex w-full max-w-sm flex-col border-l border-[#d8dad7] bg-white shadow-[-10px_0_30px_rgba(30,35,31,0.12)]">
      <header className="flex h-15 shrink-0 items-center border-b border-[#e1e3e0] px-4">
        <h2 className="font-semibold">群聊信息</h2>
        <button
          aria-label="关闭群聊信息"
          className="ml-auto inline-flex h-8 w-8 items-center justify-center rounded hover:bg-[#eceeeb]"
          onClick={onClose}
          title="关闭"
          type="button"
        >
          <X className="h-4 w-4" />
        </button>
      </header>
      <div className="min-h-0 flex-1 space-y-6 overflow-y-auto p-4">
        {group ? (
          <>
            <section>
              <label className="text-xs font-medium text-[#696d68]">群名称</label>
              <div className="mt-2 flex gap-2">
                <input
                  className="h-9 min-w-0 flex-1 rounded border border-[#d1d4d0] px-3 text-sm outline-none focus:border-[#4d8960] disabled:bg-[#f2f3f1]"
                  disabled={!group.isOwner}
                  maxLength={40}
                  onChange={(event) => setName(event.target.value)}
                  value={name}
                />
                {group.isOwner ? (
                  <button
                    className="h-9 rounded bg-[#337a49] px-3 text-xs font-medium text-white hover:bg-[#28663c]"
                    onClick={() => void handleRename()}
                    type="button"
                  >
                    保存
                  </button>
                ) : null}
              </div>
            </section>

            <section>
              <h3 className="text-xs font-medium text-[#696d68]">成员 {group.memberCount}</h3>
              <div className="mt-2 divide-y divide-[#eceeeb] border-y border-[#e2e4e1]">
                {group.members.map((member) => (
                  <div className="flex items-center gap-3 py-2.5" key={member.userId}>
                    <span className="flex h-8 w-8 items-center justify-center rounded bg-[#dce8e1] text-xs font-semibold text-[#315d4f]">
                      {member.nickname.slice(0, 1)}
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-sm">
                        {member.nickname}{member.isMe ? '（我）' : ''}
                      </span>
                      <span className="block text-[11px] text-[#8a8e88]">
                        {member.role === 'owner' ? '群主' : member.department}
                      </span>
                    </span>
                    {member.role !== 'owner' && (group.isOwner || member.isMe) ? (
                      <button
                        aria-label={member.isMe ? '退出群聊' : `移除 ${member.nickname}`}
                        className="inline-flex h-8 w-8 items-center justify-center rounded text-[#9b5353] hover:bg-[#f5e8e8]"
                        onClick={() => void handleRemove(member.userId, member.isMe)}
                        title={member.isMe ? '退出群聊' : '移除成员'}
                        type="button"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    ) : null}
                  </div>
                ))}
              </div>
            </section>

            {group.isOwner && eligibleFriends.length > 0 ? (
              <section>
                <div className="flex items-center justify-between">
                  <h3 className="text-xs font-medium text-[#696d68]">添加好友</h3>
                  <button
                    className="inline-flex h-8 items-center gap-1 rounded px-2 text-xs font-medium text-[#337a49] hover:bg-[#e9f0eb] disabled:text-[#a2aaa4]"
                    disabled={selected.size === 0}
                    onClick={() => void handleAdd()}
                    type="button"
                  >
                    <UserPlus className="h-3.5 w-3.5" />
                    添加 {selected.size || ''}
                  </button>
                </div>
                <div className="mt-2 max-h-48 overflow-y-auto border-y border-[#e2e4e1]">
                  {eligibleFriends.map((friend) => {
                    const checked = selected.has(friend.user.userId)
                    return (
                      <button
                        className="flex w-full items-center gap-3 border-b border-[#eceeeb] py-2.5 text-left last:border-0"
                        key={friend.friendshipId}
                        onClick={() =>
                          setSelected((current) => {
                            const next = new Set(current)
                            if (checked) next.delete(friend.user.userId)
                            else next.add(friend.user.userId)
                            return next
                          })
                        }
                        type="button"
                      >
                        <span className="flex h-8 w-8 items-center justify-center rounded bg-[#e5e7e4] text-xs">
                          {friend.user.nickname.slice(0, 1)}
                        </span>
                        <span className="min-w-0 flex-1 truncate text-sm">{friend.user.nickname}</span>
                        <span className={cn('h-4 w-4 rounded border', checked && 'border-[#3f8053] bg-[#3f8053]')} />
                      </button>
                    )
                  })}
                </div>
              </section>
            ) : null}
          </>
        ) : null}
        {notice ? <p className="text-sm text-[#a04b4b]">{notice}</p> : null}
      </div>
    </aside>
  )
}

export function ChatPage() {
  const { conversationId, groupId } = useParams<{
    conversationId?: string
    groupId?: string
  }>()
  const navigate = useNavigate()
  const { lastMessage, lastGroupMessage, conversationVersion } = useChat()
  const [conversations, setConversations] = useState<ConversationItem[]>([])
  const [groups, setGroups] = useState<GroupItem[]>([])
  const [friends, setFriends] = useState<FriendItem[]>([])
  const [tab, setTab] = useState<'threads' | 'contacts'>('threads')
  const [search, setSearch] = useState('')
  const [newGroupOpen, setNewGroupOpen] = useState(false)
  const [groupInfoOpen, setGroupInfoOpen] = useState(false)
  const [notice, setNotice] = useState('')

  const activeConversationId = conversationId ? Number(conversationId) : undefined
  const activeGroupId = groupId ? Number(groupId) : undefined

  const loadAll = useCallback(async () => {
    try {
      const [nextConversations, nextGroups, nextFriends] = await Promise.all([
        fetchConversations(),
        fetchGroups(),
        fetchFriends(),
      ])
      setConversations(nextConversations)
      setGroups(nextGroups)
      setFriends(nextFriends)
      setNotice('')
    } catch (error) {
      setNotice(getErrorMessage(error, '消息列表加载失败'))
    }
  }, [])

  const handleThreadRead = useCallback(() => {
    void loadAll()
  }, [loadAll])

  useEffect(() => {
    void loadAll()
  }, [loadAll, conversationVersion, lastGroupMessage, lastMessage])

  const threads = useMemo<ChatThreadItem[]>(() => {
    const directThreads: ChatThreadItem[] = conversations.map((conversation) => ({
      id: conversation.id,
      kind: 'direct',
      title: conversation.peerNickname,
      subtitle: conversation.statusMessage || conversation.lastMessage,
      lastMessageAt: conversation.lastMessageAt || conversation.createdAt,
      unreadCount: conversation.unreadCount,
    }))
    const groupThreads: ChatThreadItem[] = groups.map((group) => ({
      id: group.id,
      kind: 'group',
      title: group.name,
      subtitle: group.lastMessage
        ? `${group.lastMessageSender ? `${group.lastMessageSender}: ` : ''}${group.lastMessage}`
        : `${group.memberCount} 位成员`,
      lastMessageAt: group.lastMessageAt || group.createdAt,
      unreadCount: group.unreadCount,
    }))
    return [...directThreads, ...groupThreads]
      .filter((thread) => thread.title.toLowerCase().includes(search.trim().toLowerCase()))
      .sort(
        (left, right) =>
          new Date(right.lastMessageAt || 0).getTime() - new Date(left.lastMessageAt || 0).getTime(),
      )
  }, [conversations, groups, search])

  const activeConversation = conversations.find((item) => item.id === activeConversationId)
  const activeGroup = groups.find((item) => item.id === activeGroupId)
  const activeKey = activeGroupId
    ? `group:${activeGroupId}`
    : activeConversationId
      ? `direct:${activeConversationId}`
      : undefined

  const handleContact = async (friend: FriendItem) => {
    try {
      const conversation = await ensureConversation(friend.user.userId)
      await loadAll()
      navigate(`/chat/direct/${conversation.id}`)
      setTab('threads')
    } catch (error) {
      setNotice(getErrorMessage(error, '暂时无法发起会话'))
    }
  }

  const showList = !activeKey

  return (
    <div className="relative flex h-[calc(100dvh-7.75rem)] min-h-[520px] overflow-hidden bg-[#f2f3f1] lg:h-screen">
      <aside
        className={cn(
          'h-full w-full shrink-0 border-r border-[#d5d8d4] bg-[#f7f8f6] md:w-[330px]',
          showList ? 'block' : 'hidden md:block',
        )}
      >
        <header className="flex h-15 items-center border-b border-[#dfe1de] px-3">
          <h1 className="text-lg font-semibold text-[#292b28]">消息</h1>
          <button
            aria-label="发起群聊"
            className="ml-auto inline-flex h-9 w-9 items-center justify-center rounded text-[#555a55] hover:bg-[#e3e6e2]"
            onClick={() => setNewGroupOpen(true)}
            title="发起群聊"
            type="button"
          >
            <MessageSquarePlus className="h-5 w-5" />
          </button>
        </header>
        <div className="p-3">
          <div className="relative">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#8b8f89]" />
            <input
              aria-label="搜索会话"
              className="h-9 w-full rounded bg-[#e9ebe8] pl-9 pr-3 text-sm outline-none focus:bg-white focus:ring-1 focus:ring-[#78a184]"
              onChange={(event) => setSearch(event.target.value)}
              placeholder="搜索"
              value={search}
            />
          </div>
          <div className="mt-3 grid grid-cols-2 border-b border-[#dfe1de]">
            {([
              ['threads', '聊天'],
              ['contacts', '通讯录'],
            ] as const).map(([value, label]) => (
              <button
                className={cn(
                  'h-9 border-b-2 text-xs font-medium',
                  tab === value ? 'border-[#337a49] text-[#337a49]' : 'border-transparent text-[#777b76]',
                )}
                key={value}
                onClick={() => setTab(value)}
                type="button"
              >
                {label}
              </button>
            ))}
          </div>
        </div>

        <div className="h-[calc(100%-7.5rem)] overflow-y-auto">
          {tab === 'threads' ? (
            <ConversationList
              activeKey={activeKey}
              onSelect={(thread) => navigate(`/chat/${thread.kind}/${thread.id}`)}
              threads={threads}
            />
          ) : (
            <div>
              {friends.map((friend) => (
                <button
                  className="flex w-full items-center gap-3 border-b border-[#e4e5e2] px-4 py-3 text-left hover:bg-[#e7e9e7]"
                  key={friend.friendshipId}
                  onClick={() => void handleContact(friend)}
                  type="button"
                >
                  <span className="flex h-10 w-10 items-center justify-center rounded bg-[#dce8e1] text-sm font-semibold text-[#315d4f]">
                    {friend.user.nickname.slice(0, 1)}
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-sm font-medium">{friend.user.nickname}</span>
                    <span className="block truncate text-xs text-[#858984]">
                      {friend.user.department} · {friend.user.grade}
                    </span>
                  </span>
                </button>
              ))}
              {friends.length === 0 ? (
                <p className="py-10 text-center text-sm text-[#858984]">暂无好友</p>
              ) : null}
            </div>
          )}
        </div>
      </aside>

      <section className={cn('min-w-0 flex-1', activeKey ? 'block' : 'hidden md:block')}>
        {activeConversation ? (
          <div className="flex h-full flex-col">
            <button
              className="flex h-10 shrink-0 items-center gap-1 border-b border-[#dfe1de] bg-[#f7f8f6] px-3 text-xs text-[#5f645f] md:hidden"
              onClick={() => navigate('/chat')}
              type="button"
            >
              <ArrowLeft className="h-4 w-4" />
              返回
            </button>
            <div className="min-h-0 flex-1">
              <ChatWindow
                onRead={handleThreadRead}
                thread={{
                  id: activeConversation.id,
                  kind: 'direct',
                  title: activeConversation.peerNickname,
                  subtitle: `${activeConversation.peerDepartment} · ${activeConversation.peerGrade}`,
                  canSend: activeConversation.canSend,
                  statusMessage: activeConversation.statusMessage,
                }}
              />
            </div>
          </div>
        ) : activeGroup ? (
          <div className="flex h-full flex-col">
            <button
              className="flex h-10 shrink-0 items-center gap-1 border-b border-[#dfe1de] bg-[#f7f8f6] px-3 text-xs text-[#5f645f] md:hidden"
              onClick={() => navigate('/chat')}
              type="button"
            >
              <ArrowLeft className="h-4 w-4" />
              返回
            </button>
            <div className="min-h-0 flex-1">
              <ChatWindow
                onOpenGroupInfo={() => setGroupInfoOpen(true)}
                onRead={handleThreadRead}
                thread={{
                  id: activeGroup.id,
                  kind: 'group',
                  title: activeGroup.name,
                  subtitle: `${activeGroup.memberCount} 位成员`,
                  canSend: true,
                }}
              />
            </div>
          </div>
        ) : (
          <div className="flex h-full items-center justify-center bg-[#f2f3f1]">
            <div className="text-center text-[#8a8e88]">
              <Users className="mx-auto h-10 w-10" strokeWidth={1.4} />
              <p className="mt-3 text-sm">选择一个会话</p>
            </div>
          </div>
        )}
      </section>

      {notice ? (
        <div className="absolute bottom-4 left-1/2 z-40 -translate-x-1/2 rounded bg-[#303330] px-4 py-2 text-sm text-white shadow-lg">
          {notice}
        </div>
      ) : null}

      {newGroupOpen ? (
        <NewGroupDialog
          friends={friends}
          onClose={() => setNewGroupOpen(false)}
          onCreated={(group) => {
            setNewGroupOpen(false)
            void loadAll()
            navigate(`/chat/group/${group.id}`)
          }}
        />
      ) : null}

      {groupInfoOpen && activeGroupId ? (
        <GroupInfoPanel
          friends={friends}
          groupId={activeGroupId}
          onChanged={() => void loadAll()}
          onClose={() => setGroupInfoOpen(false)}
          onLeft={() => {
            setGroupInfoOpen(false)
            void loadAll()
            navigate('/chat')
          }}
        />
      ) : null}
    </div>
  )
}
