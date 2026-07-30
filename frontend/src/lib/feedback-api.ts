import { apiRequest } from '@/lib/api'

export interface FeedbackSubmitResult {
  ticketId: string
  receivedAt: string
}

export function submitFeedback(input: {
  category: 'match' | 'bug' | 'feature' | 'other'
  rating: number
  details: string
  contactEmail?: string
}) {
  return apiRequest<FeedbackSubmitResult>('/feedback', {
    method: 'POST',
    auth: true,
    body: input,
  })
}
