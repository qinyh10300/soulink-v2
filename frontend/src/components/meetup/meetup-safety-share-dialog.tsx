import {
  Check,
  Clipboard,
  LoaderCircle,
  Mail,
  Plus,
  ShieldCheck,
  Trash2,
  X,
} from 'lucide-react'
import { useEffect, useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import {
  createTrustedContact,
  deleteTrustedContact,
  fetchTrustedContacts,
  shareMeetupSafetySummary,
  type Meetup,
  type MeetupSafetyShare,
  type TrustedContact,
} from '@/lib/meetup-api'

export function MeetupSafetyShareDialog({ meetup, onClose }: { meetup: Meetup; onClose: () => void }) {
  const [contacts, setContacts] = useState<TrustedContact[]>([])
  const [selectedID, setSelectedID] = useState<number | null>(null)
  const [adding, setAdding] = useState(false)
  const [label, setLabel] = useState('')
  const [contactType, setContactType] = useState<TrustedContact['contactType']>('email')
  const [contactValue, setContactValue] = useState('')
  const [result, setResult] = useState<MeetupSafetyShare | null>(null)
  const [busy, setBusy] = useState(false)
  const [notice, setNotice] = useState('')

  useEffect(() => {
    let cancelled = false
    void fetchTrustedContacts()
      .then((items) => {
        if (cancelled) return
        setContacts(items)
        setSelectedID(items[0]?.id ?? null)
      })
      .catch((error: unknown) => {
        if (!cancelled) setNotice(getErrorMessage(error, '可信联系人加载失败'))
      })
    return () => {
      cancelled = true
    }
  }, [])

  const addContact = async () => {
    if (busy) return
    setBusy(true)
    try {
      const created = await createTrustedContact({ label, contactType, contactValue })
      setContacts((current) => [created, ...current])
      setSelectedID(created.id)
      setLabel('')
      setContactValue('')
      setAdding(false)
      setNotice('')
    } catch (error) {
      setNotice(getErrorMessage(error, '可信联系人保存失败'))
    } finally {
      setBusy(false)
    }
  }

  const removeContact = async (contact: TrustedContact) => {
    if (busy) return
    setBusy(true)
    try {
      await deleteTrustedContact(contact.id)
      setContacts((current) => current.filter((item) => item.id !== contact.id))
      if (selectedID === contact.id) setSelectedID(null)
    } catch (error) {
      setNotice(getErrorMessage(error, '可信联系人删除失败'))
    } finally {
      setBusy(false)
    }
  }

  const share = async () => {
    if (!selectedID || busy) return
    setBusy(true)
    setNotice('')
    try {
      const shareResult = await shareMeetupSafetySummary(meetup.id, selectedID)
      setResult(shareResult)
      setNotice(shareResult.deliveryStatus === 'sent' ? '安全摘要邮件已发送' : '安全摘要已生成')
    } catch (error) {
      setNotice(getErrorMessage(error, '安全摘要分享失败'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[90] flex items-end justify-center bg-black/35 sm:items-center sm:p-5">
      <section aria-modal="true" className="flex max-h-[92dvh] w-full max-w-lg flex-col bg-white shadow-2xl sm:rounded-md" role="dialog">
        <header className="flex h-14 shrink-0 items-center border-b border-[#dedfdc] px-4">
          <ShieldCheck className="h-5 w-5 text-[#376b58]" />
          <h2 className="ml-2 text-sm font-semibold text-[#303430]">分享约见安全摘要</h2>
          <button aria-label="关闭安全分享" className="ml-auto inline-flex h-9 w-9 items-center justify-center rounded text-[#70756f] hover:bg-[#eef0ed]" onClick={onClose} type="button">
            <X className="h-4 w-4" />
          </button>
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          <div className="grid grid-cols-[5rem_1fr] gap-x-3 gap-y-2 border-y border-[#e1e3df] py-3 text-xs">
            <span className="text-[#92968f]">时间</span>
            <span className="text-[#4f554f]">{new Date(meetup.startsAt).toLocaleString('zh-CN', { hour12: false })}</span>
            <span className="text-[#92968f]">地点</span>
            <span className="text-[#4f554f]">{meetup.location || '校内地点待确认'}</span>
            <span className="text-[#92968f]">参与人数</span>
            <span className="text-[#4f554f]">{meetup.participants.length} 人</span>
          </div>

          <div className="mt-4 flex items-center justify-between">
            <h3 className="text-xs font-semibold text-[#4b504c]">可信联系人</h3>
            <button className="inline-flex h-8 items-center gap-1.5 rounded px-2 text-xs font-medium text-[#376b58] hover:bg-[#edf3ef]" onClick={() => setAdding((current) => !current)} type="button">
              <Plus className="h-3.5 w-3.5" />新增
            </button>
          </div>

          {adding ? (
            <div className="mt-2 grid gap-2 border border-[#dfe1dd] bg-[#f8f9f7] p-3 sm:grid-cols-2">
              <input aria-label="可信联系人称呼" className="h-9 rounded border border-[#d4d7d3] px-3 text-xs outline-none focus:border-[#668477]" onChange={(event) => setLabel(event.target.value)} placeholder="称呼" value={label} />
              <select aria-label="可信联系人类型" className="h-9 rounded border border-[#d4d7d3] bg-white px-3 text-xs" onChange={(event) => setContactType(event.target.value as TrustedContact['contactType'])} value={contactType}>
                <option value="email">邮箱</option>
                <option value="phone">手机</option>
                <option value="wechat">微信</option>
                <option value="other">其他</option>
              </select>
              <input aria-label="可信联系人联系方式" className="h-9 rounded border border-[#d4d7d3] px-3 text-xs outline-none focus:border-[#668477] sm:col-span-2" onChange={(event) => setContactValue(event.target.value)} placeholder="联系方式" value={contactValue} />
              <button className="inline-flex h-9 items-center justify-center gap-1.5 rounded bg-[#376b58] text-xs font-semibold text-white disabled:opacity-50 sm:col-span-2" disabled={busy || !label.trim() || !contactValue.trim()} onClick={() => void addContact()} type="button">
                {busy ? <LoaderCircle className="h-3.5 w-3.5 animate-spin" /> : <Check className="h-3.5 w-3.5" />}保存联系人
              </button>
            </div>
          ) : null}

          <div className="mt-2 divide-y divide-[#e5e6e3] border-y border-[#e1e3df]">
            {contacts.map((contact) => (
              <div className="flex min-h-12 items-center gap-3" key={contact.id}>
                <input checked={selectedID === contact.id} className="h-4 w-4 accent-[#376b58]" name="trusted-contact" onChange={() => setSelectedID(contact.id)} type="radio" />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-xs font-semibold text-[#4b504c]">{contact.label}</p>
                  <p className="truncate text-[11px] text-[#898e88]">{contact.contactValue}</p>
                </div>
                <button aria-label={`删除可信联系人 ${contact.label}`} className="inline-flex h-8 w-8 items-center justify-center rounded text-[#8a7470] hover:bg-[#f5eae7]" disabled={busy} onClick={() => void removeContact(contact)} title="删除" type="button">
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
            ))}
            {contacts.length === 0 ? <p className="py-7 text-center text-xs text-[#8b908a]">还没有可信联系人</p> : null}
          </div>

          {result?.manualText ? (
            <div className="mt-4">
              <textarea aria-label="安全摘要文本" className="min-h-28 w-full resize-y rounded border border-[#d5d8d4] bg-[#fafbf9] px-3 py-2 text-xs leading-5" readOnly value={result.manualText} />
              <button className="mt-2 inline-flex h-8 items-center gap-1.5 rounded border border-[#ced4cf] px-3 text-xs font-semibold text-[#426456] hover:bg-[#edf2ef]" onClick={() => void navigator.clipboard.writeText(result.manualText ?? '')} type="button">
                <Clipboard className="h-3.5 w-3.5" />复制摘要
              </button>
            </div>
          ) : null}
          {notice ? <p className="mt-3 text-xs text-[#755b52]">{notice}</p> : null}
        </div>

        <footer className="shrink-0 border-t border-[#dedfdc] p-4">
          <button className="inline-flex h-10 w-full items-center justify-center gap-2 rounded bg-[#315f4f] text-sm font-semibold text-white hover:bg-[#274e41] disabled:opacity-45" disabled={!selectedID || busy} onClick={() => void share()} type="button">
            {busy ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <Mail className="h-4 w-4" />}
            分享安全摘要
          </button>
        </footer>
      </section>
    </div>
  )
}
