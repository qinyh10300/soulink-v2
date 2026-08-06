import {
  BrainCircuit,
  CalendarClock,
  FileText,
  ListTodo,
  LoaderCircle,
  MapPin,
  MessageSquareText,
  Power,
  RotateCcw,
  Sparkles,
  Trash2,
  X,
} from 'lucide-react'
import { useEffect, useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import {
  deleteAssistantArtifact,
  fetchAssistantWorkspace,
  generateAssistantArtifact,
  resetAssistantMemory,
  updateAssistantPreferences,
  type AssistantActionCard,
  type AssistantActionsData,
  type AssistantArtifact,
  type AssistantGroupSummaryData,
  type AssistantStartersData,
  type AssistantSummaryData,
} from '@/lib/chat-api'
import { cn } from '@/lib/utils'

interface Props {
  contextId: number
  contextType: 'direct' | 'group'
  onClose: () => void
  onOpenMeetup: (card?: AssistantActionCard) => void
  onUseDraft: (text: string) => void
}

function ArtifactMeta({ artifact }: { artifact: AssistantArtifact }) {
  return (
    <span
      className="text-[11px] text-[#858b85]"
      title={`${artifact.inferenceSource} · ${artifact.promptVersion}`}
    >
      {artifact.inferenceSource === 'model' ? 'AI' : '本地规则'} · {Math.round(artifact.confidence * 100)}%
    </span>
  )
}

function SummaryList({ label, values }: { label: string; values: string[] }) {
  if (values.length === 0) return null
  return (
    <div className="mt-3">
      <h4 className="text-xs font-semibold text-[#5c635d]">{label}</h4>
      <ul className="mt-1.5 space-y-1.5 text-xs leading-5 text-[#555c56]">
        {values.map((value) => <li key={value}>• {value}</li>)}
      </ul>
    </div>
  )
}

function ActionCardEditor({
  card,
  onOpenMeetup,
  onUseDraft,
}: {
  card: AssistantActionCard
  onOpenMeetup: (card: AssistantActionCard) => void
  onUseDraft: (text: string) => void
}) {
  const [detail, setDetail] = useState(card.detail)
  const [timeText, setTimeText] = useState(card.timeText ?? '')
  const [location, setLocation] = useState(card.location ?? '')
  const [todoText, setTodoText] = useState(card.todoText ?? '')
  const Icon = card.type === 'location' ? MapPin : card.type === 'todo' ? ListTodo : CalendarClock

  const draft = card.type === 'time'
    ? timeText || detail
    : card.type === 'location'
      ? location || detail
      : todoText || detail
  return (
    <div className="rounded-md border border-[#d9ddd9] bg-white p-3">
      <div className="flex items-center gap-2">
        <Icon className="h-4 w-4 text-[#477060]" />
        <strong className="text-xs text-[#363c38]">{card.title}</strong>
        <span className="ml-auto text-[11px] text-[#8a8f8a]">{Math.round(card.confidence * 100)}%</span>
      </div>
      {card.type === 'meetup' ? (
        <textarea
          aria-label={`${card.title}内容`}
          className="mt-2 min-h-16 w-full resize-y rounded border border-[#d5d9d5] px-2.5 py-2 text-xs leading-5 outline-none focus:border-[#608273]"
          onChange={(event) => setDetail(event.target.value)}
          value={detail}
        />
      ) : null}
      {card.type === 'time' || card.type === 'meetup' ? (
        <input
          aria-label="时间草稿"
          className="mt-2 h-8 w-full rounded border border-[#d5d9d5] px-2.5 text-xs outline-none focus:border-[#608273]"
          onChange={(event) => setTimeText(event.target.value)}
          placeholder="时间"
          value={timeText}
        />
      ) : null}
      {card.type === 'location' || card.type === 'meetup' ? (
        <input
          aria-label="地点草稿"
          className="mt-2 h-8 w-full rounded border border-[#d5d9d5] px-2.5 text-xs outline-none focus:border-[#608273]"
          onChange={(event) => setLocation(event.target.value)}
          placeholder="地点"
          value={location}
        />
      ) : null}
      {card.type === 'todo' ? (
        <textarea
          aria-label="待办草稿"
          className="mt-2 min-h-16 w-full resize-y rounded border border-[#d5d9d5] px-2.5 py-2 text-xs leading-5 outline-none focus:border-[#608273]"
          onChange={(event) => setTodoText(event.target.value)}
          placeholder="待办"
          value={todoText}
        />
      ) : null}
      <div className="mt-2 flex justify-end">
        {card.type === 'meetup' ? (
          <button
            className="inline-flex h-8 items-center gap-1.5 rounded bg-[#315c4d] px-3 text-xs font-semibold text-white hover:bg-[#264b3f]"
            onClick={() => onOpenMeetup({ ...card, detail, timeText, location, todoText })}
            type="button"
          >
            <CalendarClock className="h-3.5 w-3.5" />打开约见
          </button>
        ) : (
          <button
            className="inline-flex h-8 items-center gap-1.5 rounded border border-[#ccd2cd] bg-white px-3 text-xs font-semibold text-[#49665a] hover:bg-[#edf1ee]"
            onClick={() => onUseDraft(draft)}
            type="button"
          >
            <MessageSquareText className="h-3.5 w-3.5" />填入输入框
          </button>
        )}
      </div>
    </div>
  )
}

export function ChatAssistantPanel({
  contextId,
  contextType,
  onClose,
  onOpenMeetup,
  onUseDraft,
}: Props) {
  const [artifacts, setArtifacts] = useState<AssistantArtifact[]>([])
  const [enabled, setEnabled] = useState(true)
  const [loading, setLoading] = useState(true)
  const [action, setAction] = useState('')
  const [notice, setNotice] = useState('')
  const [confirmReset, setConfirmReset] = useState(false)

  useEffect(() => {
    let cancelled = false
    void fetchAssistantWorkspace(contextType, contextId)
      .then((workspace) => {
        if (cancelled) return
        setArtifacts(workspace.artifacts)
        setEnabled(workspace.preferences.enabled)
      })
      .catch((error: unknown) => {
        if (!cancelled) setNotice(getErrorMessage(error, '助手加载失败'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [contextId, contextType])

  const generate = async (artifactType: 'starters' | 'summary' | 'actions') => {
    try {
      setAction(artifactType)
      setNotice('')
      const artifact = await generateAssistantArtifact(contextType, contextId, artifactType)
      setArtifacts((current) => [
        artifact,
        ...current.filter((item) => item.artifactType !== artifact.artifactType),
      ])
    } catch (error) {
      setNotice(getErrorMessage(error, '草稿生成失败'))
    } finally {
      setAction('')
    }
  }

  const remove = async (artifact: AssistantArtifact) => {
    try {
      setAction(`delete-${artifact.id}`)
      await deleteAssistantArtifact(contextType, contextId, artifact.id)
      setArtifacts((current) => current.filter((item) => item.id !== artifact.id))
    } catch (error) {
      setNotice(getErrorMessage(error, '删除失败'))
    } finally {
      setAction('')
    }
  }

  const toggle = async () => {
    try {
      setAction('toggle')
      const preferences = await updateAssistantPreferences(!enabled)
      setEnabled(preferences.enabled)
      setNotice('')
    } catch (error) {
      setNotice(getErrorMessage(error, '设置保存失败'))
    } finally {
      setAction('')
    }
  }

  const reset = async () => {
    if (!confirmReset) {
      setConfirmReset(true)
      return
    }
    try {
      setAction('reset')
      await resetAssistantMemory()
      setArtifacts([])
      setConfirmReset(false)
      setNotice('助手记忆已重置')
    } catch (error) {
      setNotice(getErrorMessage(error, '重置失败'))
    } finally {
      setAction('')
    }
  }

  return (
    <aside className="absolute inset-y-0 right-0 z-30 flex w-full flex-col border-l border-[#d7dad6] bg-[#f8f9f7] shadow-[-12px_0_30px_rgba(43,49,44,0.12)] sm:w-[380px]">
      <header className="flex h-15 shrink-0 items-center border-b border-[#dfe1de] px-4">
        <BrainCircuit className="h-5 w-5 text-[#3f6858]" />
        <h2 className="ml-2 text-sm font-semibold text-[#303531]">聊天助手</h2>
        <button
          aria-label={enabled ? '关闭聊天助手' : '开启聊天助手'}
          className={cn(
            'ml-auto inline-flex h-7 w-12 items-center rounded-full p-0.5 transition-colors',
            enabled ? 'justify-end bg-[#3c755f]' : 'justify-start bg-[#b9bdb9]',
          )}
          disabled={action === 'toggle'}
          onClick={() => void toggle()}
          title={enabled ? '关闭助手' : '开启助手'}
          type="button"
        >
          <span className="h-6 w-6 rounded-full bg-white shadow-sm" />
        </button>
        <button
          aria-label="关闭助手面板"
          className="ml-1 inline-flex h-9 w-9 items-center justify-center rounded text-[#717771] hover:bg-[#e7e9e6]"
          onClick={onClose}
          title="关闭"
          type="button"
        >
          <X className="h-4 w-4" />
        </button>
      </header>

      <div className="flex-1 overflow-y-auto">
        <div className="border-b border-[#dfe1de] p-4">
          <div className="grid grid-cols-2 gap-2">
            {contextType === 'direct' ? (
              <button
                className="inline-flex h-9 items-center justify-center gap-2 rounded border border-[#ccd2cd] bg-white text-xs font-semibold text-[#49665a] hover:bg-[#edf1ee] disabled:opacity-50"
                disabled={!enabled || Boolean(action)}
                onClick={() => void generate('starters')}
                type="button"
              >
                {action === 'starters' ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                开场建议
              </button>
            ) : null}
            <button
              className="inline-flex h-9 items-center justify-center gap-2 rounded border border-[#ccd2cd] bg-white text-xs font-semibold text-[#49665a] hover:bg-[#edf1ee] disabled:opacity-50"
              disabled={!enabled || Boolean(action)}
              onClick={() => void generate('summary')}
              type="button"
            >
              {action === 'summary' ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <FileText className="h-4 w-4" />}
              {contextType === 'group' ? '群聊共识' : '私人摘要'}
            </button>
            <button
              className={cn(
                'inline-flex h-9 items-center justify-center gap-2 rounded border border-[#ccd2cd] bg-white text-xs font-semibold text-[#49665a] hover:bg-[#edf1ee] disabled:opacity-50',
                contextType === 'group' && 'col-span-2',
              )}
              disabled={!enabled || Boolean(action)}
              onClick={() => void generate('actions')}
              type="button"
            >
              {action === 'actions' ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <ListTodo className="h-4 w-4" />}
              行动线索
            </button>
          </div>
          {notice ? <p className="mt-3 rounded bg-[#ecefec] px-3 py-2 text-xs text-[#646b65]">{notice}</p> : null}
        </div>

        {loading ? (
          <div className="flex min-h-40 items-center justify-center text-xs text-[#7c827c]">
            <LoaderCircle className="mr-2 h-4 w-4 animate-spin" />正在加载
          </div>
        ) : null}

        {!loading && enabled ? (
          <div className="divide-y divide-[#dfe1de]">
            {artifacts.map((artifact) => (
              <section className="px-4 py-4" key={artifact.id}>
                <div className="flex items-center gap-2">
                  <h3 className="text-xs font-semibold text-[#3b413d]">
                    {artifact.artifactType === 'starters'
                      ? '开场建议'
                      : artifact.artifactType === 'actions'
                        ? '行动线索'
                        : artifact.artifactType === 'group_summary'
                          ? '群聊共识'
                          : '私人摘要'}
                  </h3>
                  <ArtifactMeta artifact={artifact} />
                  <button
                    aria-label="删除助手草稿"
                    className="ml-auto inline-flex h-7 w-7 items-center justify-center rounded text-[#8b7a78] hover:bg-[#f2e7e5]"
                    disabled={Boolean(action)}
                    onClick={() => void remove(artifact)}
                    title="删除"
                    type="button"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>

                {artifact.artifactType === 'starters' ? (
                  <div className="mt-3 space-y-2">
                    {(artifact.data as AssistantStartersData).starters.map((starter) => (
                      <button
                        className="w-full rounded-md border border-[#d9ddd9] bg-white px-3 py-2.5 text-left text-xs leading-5 text-[#4b524d] hover:border-[#83a092] hover:bg-[#f2f6f3]"
                        key={starter}
                        onClick={() => onUseDraft(starter)}
                        type="button"
                      >
                        {starter}
                      </button>
                    ))}
                  </div>
                ) : null}

                {artifact.artifactType === 'summary' ? (
                  <div className="mt-3">
                    <p className="text-xs leading-5 text-[#4b524d]">{(artifact.data as AssistantSummaryData).summary}</p>
                    <SummaryList label="要点" values={(artifact.data as AssistantSummaryData).keyPoints} />
                    <SummaryList label="待确认" values={(artifact.data as AssistantSummaryData).openQuestions} />
                  </div>
                ) : null}

                {artifact.artifactType === 'group_summary' ? (
                  <div className="mt-3">
                    <p className="text-xs leading-5 text-[#4b524d]">{(artifact.data as AssistantGroupSummaryData).summary}</p>
                    <SummaryList label="已形成共识" values={(artifact.data as AssistantGroupSummaryData).agreements} />
                    <SummaryList label="仍有分歧" values={(artifact.data as AssistantGroupSummaryData).openQuestions} />
                    <SummaryList label="下一步" values={(artifact.data as AssistantGroupSummaryData).nextSteps} />
                  </div>
                ) : null}

                {artifact.artifactType === 'actions' ? (
                  <div className="mt-3 space-y-2">
                    {(artifact.data as AssistantActionsData).cards.map((card) => (
                      <ActionCardEditor
                        card={card}
                        key={card.id}
                        onOpenMeetup={onOpenMeetup}
                        onUseDraft={onUseDraft}
                      />
                    ))}
                    {(artifact.data as AssistantActionsData).cards.length === 0 ? (
                      <p className="py-3 text-center text-xs text-[#858b85]">暂未识别到明确行动线索</p>
                    ) : null}
                  </div>
                ) : null}
              </section>
            ))}
            {artifacts.length === 0 ? (
              <div className="px-6 py-12 text-center text-xs leading-5 text-[#858b85]">
                <Sparkles className="mx-auto mb-2 h-5 w-5" />
                暂无助手草稿
              </div>
            ) : null}
          </div>
        ) : null}

        {!loading && !enabled ? (
          <div className="px-6 py-14 text-center text-xs text-[#7c827c]">
            <Power className="mx-auto mb-2 h-5 w-5" />
            聊天助手已关闭
          </div>
        ) : null}
      </div>

      <footer className="shrink-0 border-t border-[#dfe1de] px-4 py-3">
        <button
          className={cn(
            'inline-flex h-8 w-full items-center justify-center gap-2 rounded text-xs font-medium',
            confirmReset ? 'bg-[#8a4545] text-white' : 'text-[#747a74] hover:bg-[#e9ebe8]',
          )}
          disabled={Boolean(action)}
          onClick={() => void reset()}
          type="button"
        >
          <RotateCcw className="h-3.5 w-3.5" />
          {confirmReset ? '再次点击确认重置' : '重置我的助手记忆'}
        </button>
      </footer>
    </aside>
  )
}
