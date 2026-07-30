import { apiRequest } from '@/lib/api'

export interface VerificationCodeSendResult {
  expireSeconds: number
  cooldownSeconds?: number
}

export interface AuthUser {
  id: number
  email: string
  nickname: string
  schoolName: string
  department: string
  grade: string
  hasCompletedQuestionnaire: boolean
}

export interface AuthPayload {
  token: string
  user: AuthUser
}

export interface CurrentAuthUser extends AuthUser {
  contactType: string
  contactValue: string
}

export function login(email: string, password: string) {
  return apiRequest<AuthPayload>('/auth/login', {
    method: 'POST',
    body: {
      email,
      password,
    },
  })
}

export function sendRegisterCode(email: string) {
  return apiRequest<VerificationCodeSendResult>('/auth/register/send-code', {
    method: 'POST',
    body: { email },
  })
}

export function verifyRegisterCode(email: string, code: string) {
  return apiRequest<{ verified: boolean }>('/auth/register/verify-code', {
    method: 'POST',
    body: {
      email,
      code,
    },
  })
}

export function registerAccount(input: {
  email: string
  code: string
  password: string
  nickname: string
  department: string
  grade: string
}) {
  return apiRequest<AuthPayload>('/auth/register', {
    method: 'POST',
    body: input,
  })
}

export function sendForgotPasswordCode(email: string) {
  return apiRequest<VerificationCodeSendResult>('/auth/password/forgot/send-code', {
    method: 'POST',
    body: { email },
  })
}

export function resetForgotPassword(input: {
  email: string
  code: string
  newPassword: string
}) {
  return apiRequest<null>('/auth/password/forgot/reset', {
    method: 'POST',
    body: input,
  })
}

export function logout() {
  return apiRequest<null>('/auth/logout', {
    method: 'POST',
    auth: true,
  })
}

export function fetchCurrentUser() {
  return apiRequest<CurrentAuthUser>('/auth/me', {
    auth: true,
  })
}

export function sendChangePasswordCode() {
  return apiRequest<{ email: string } & VerificationCodeSendResult>('/users/me/password/send-code', {
    method: 'POST',
    auth: true,
  })
}

export function changePassword(input: {
  code: string
  newPassword: string
}) {
  return apiRequest<null>('/users/me/password', {
    method: 'POST',
    auth: true,
    body: input,
  })
}

export function sendDeleteAccountCode() {
  return apiRequest<{ email: string } & VerificationCodeSendResult>('/users/me/delete/send-code', {
    method: 'POST',
    auth: true,
  })
}

export function deleteMyAccount(code: string) {
  return apiRequest<null>('/users/me', {
    method: 'DELETE',
    auth: true,
    body: { code },
  })
}
