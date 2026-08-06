import {
  Camera,
  Check,
  Eye,
  LoaderCircle,
  RefreshCw,
  RotateCcw,
  X,
} from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import {
  fetchPhotoVerificationEvidence,
  fetchPhotoVerifications,
  reviewPhotoVerification,
  type PhotoVerification,
  type PhotoVerificationStatus,
} from '@/lib/safety-api'

const statusLabels: Record<PhotoVerificationStatus, string> = {
  pending: '待审核',
  approved: '已通过',
  rejected: '未通过',
}

export function PhotoVerificationAdminPanel() {
  const [status, setStatus] = useState<PhotoVerificationStatus | ''>('pending')
  const [requests, setRequests] = useState<PhotoVerification[]>([])
  const [evidenceUrls, setEvidenceUrls] = useState<Record<number, string>>({})
  const evidenceUrlsRef = useRef<Record<number, string>>({})
  const [notes, setNotes] = useState<Record<number, string>>({})
  const [busyID, setBusyID] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setNotice('')
    try {
      const result = await fetchPhotoVerifications(status)
      setRequests(result)
      setNotes(Object.fromEntries(result.map((item) => [item.id, item.reviewNotes])))
    } catch (error) {
      setNotice(getErrorMessage(error, '照片审核队列加载失败'))
    } finally {
      setLoading(false)
    }
  }, [status])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    evidenceUrlsRef.current = evidenceUrls
  }, [evidenceUrls])

  useEffect(() => () => {
    Object.values(evidenceUrlsRef.current).forEach((url) => URL.revokeObjectURL(url))
  }, [])

  const revealEvidence = async (requestID: number) => {
    if (evidenceUrls[requestID] || busyID) return
    setBusyID(requestID)
    try {
      const blob = await fetchPhotoVerificationEvidence(requestID)
      const url = URL.createObjectURL(blob)
      setEvidenceUrls((current) => ({ ...current, [requestID]: url }))
    } catch (error) {
      setNotice(getErrorMessage(error, '审核照片读取失败'))
    } finally {
      setBusyID(null)
    }
  }

  const review = async (request: PhotoVerification, nextStatus: 'approved' | 'rejected') => {
    if (busyID) return
    setBusyID(request.id)
    setNotice('')
    try {
      const updated = await reviewPhotoVerification(request.id, {
        status: nextStatus,
        notes: notes[request.id] ?? '',
      })
      setRequests((current) =>
        status && updated.status !== status
          ? current.filter((item) => item.id !== updated.id)
          : current.map((item) => (item.id === updated.id ? updated : item)),
      )
      setNotice(nextStatus === 'approved' ? '照片审核已通过' : '审核结果已保存')
    } catch (error) {
      setNotice(getErrorMessage(error, '照片审核更新失败'))
    } finally {
      setBusyID(null)
    }
  }

  return (
    <section className="py-6">
      <div className="flex flex-col gap-3 border-b border-[#dfdcd6] pb-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-base font-semibold text-[#34312d]">照片人工审核</h2>
          <p className="mt-1 text-xs text-[#807b74]">原图仅在当前受限审核页面按需读取</p>
        </div>
        <div className="flex items-center gap-2">
          <select
            aria-label="筛选照片审核状态"
            className="h-9 rounded border border-[#d8d4cd] bg-white px-3 text-xs"
            onChange={(event) => setStatus(event.target.value as PhotoVerificationStatus | '')}
            value={status}
          >
            <option value="">全部状态</option>
            <option value="pending">待审核</option>
            <option value="approved">已通过</option>
            <option value="rejected">未通过</option>
          </select>
          <button
            aria-label="刷新照片审核"
            className="inline-flex h-9 w-9 items-center justify-center rounded border border-[#d8d4cd] bg-white text-[#69655f] hover:bg-[#f1efeb]"
            onClick={() => void load()}
            title="刷新"
            type="button"
          >
            <RefreshCw className="h-4 w-4" />
          </button>
        </div>
      </div>

      {notice ? <div className="mt-4 rounded border border-[#e5ddd4] bg-[#fffaf6] px-3 py-2 text-xs text-[#765e50]">{notice}</div> : null}
      {loading ? (
        <div className="flex min-h-64 items-center justify-center text-sm text-[#77736c]">
          <LoaderCircle className="mr-2 h-4 w-4 animate-spin" />正在读取审核队列
        </div>
      ) : requests.length === 0 ? (
        <div className="flex min-h-64 flex-col items-center justify-center border-b border-[#dfdcd6] text-sm text-[#77736c]">
          <Camera className="mb-3 h-6 w-6 text-[#668477]" />当前筛选下没有照片审核
        </div>
      ) : (
        <div className="divide-y divide-[#e3e0db] border-b border-[#dfdcd6]">
          {requests.map((request) => (
            <article className="grid gap-5 py-5 lg:grid-cols-[18rem_minmax(0,1fr)_18rem]" key={request.id}>
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <strong className="truncate text-sm text-[#37332f]">{request.userNickname}</strong>
                  <span className="rounded bg-[#efede9] px-2 py-1 text-[10px] text-[#6f6a64]">{statusLabels[request.status]}</span>
                </div>
                <p className="mt-1 truncate text-xs text-[#8a857e]">{request.userEmail}</p>
                {request.appealReason ? (
                  <div className="mt-3 rounded border border-[#ddd4e3] bg-[#f8f4fa] px-3 py-2 text-xs leading-5 text-[#684d78]">
                    <p className="mb-1 flex items-center gap-1 font-semibold"><RotateCcw className="h-3.5 w-3.5" />用户申诉</p>
                    {request.appealReason}
                  </div>
                ) : null}
              </div>

              <div className="min-h-44 overflow-hidden border border-[#ddd9d3] bg-[#f1efeb]">
                {evidenceUrls[request.id] ? (
                  <img alt={`${request.userNickname} 的人工审核照片`} className="h-full max-h-80 w-full object-contain" src={evidenceUrls[request.id]} />
                ) : (
                  <button
                    className="flex h-full min-h-44 w-full flex-col items-center justify-center text-xs font-medium text-[#5f6e66] hover:bg-[#e8ece9]"
                    disabled={busyID === request.id}
                    onClick={() => void revealEvidence(request.id)}
                    type="button"
                  >
                    {busyID === request.id ? <LoaderCircle className="mb-2 h-5 w-5 animate-spin" /> : <Eye className="mb-2 h-5 w-5" />}
                    读取受限照片
                  </button>
                )}
              </div>

              <div>
                <textarea
                  aria-label={`照片审核 ${request.id} 的说明`}
                  className="min-h-24 w-full resize-y rounded border border-[#d8d4cd] bg-white px-3 py-2 text-xs leading-5 outline-none focus:border-[#668477]"
                  maxLength={1000}
                  onChange={(event) => setNotes((current) => ({ ...current, [request.id]: event.target.value }))}
                  placeholder="拒绝时必须说明原因"
                  value={notes[request.id] ?? ''}
                />
                {request.status === 'pending' ? (
                  <div className="mt-3 grid grid-cols-2 gap-2">
                    <button
                      className="inline-flex h-9 items-center justify-center gap-1.5 rounded border border-[#d7c5c0] text-xs font-semibold text-[#8a4e43] hover:bg-[#fff3f0] disabled:opacity-45"
                      disabled={busyID === request.id || (notes[request.id] ?? '').trim().length < 4}
                      onClick={() => void review(request, 'rejected')}
                      type="button"
                    >
                      <X className="h-4 w-4" />不通过
                    </button>
                    <button
                      className="inline-flex h-9 items-center justify-center gap-1.5 rounded bg-[#315f4f] text-xs font-semibold text-white hover:bg-[#274e41] disabled:opacity-45"
                      disabled={busyID === request.id || !evidenceUrls[request.id]}
                      onClick={() => void review(request, 'approved')}
                      type="button"
                    >
                      <Check className="h-4 w-4" />通过
                    </button>
                  </div>
                ) : null}
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  )
}
