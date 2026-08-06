import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { AuthSplitLayout } from '@/components/auth/auth-split-layout'
import { RegisterForm } from '@/components/auth/register-form'
import { RegisterPasswordStep } from '@/components/auth/register-password-step'
import { RegisterProfileStep } from '@/components/auth/register-profile-step'
import { FloatingToast } from '@/components/shared/floating-toast'
import { getErrorMessage } from '@/lib/api'
import {
  registerAccount,
  sendRegisterCode,
  verifyRegisterCode,
} from '@/lib/auth-api'
import type {
  RegisterEmailStepValues,
  RegisterPasswordStepValues,
  RegisterProfileStepValues,
} from '@/lib/auth-schema'
import { fetchGrades } from '@/lib/meta-api'
import { resetDashboardAnnouncementSeen } from '@/lib/survey-storage'
import { connectionPaths } from '@/lib/connection-routes'
import { saveAuthenticatedUser } from '@/lib/user-profile-storage'
import { VERIFICATION_CODE_COUNTDOWN_SECONDS } from '@/lib/verification-code'

type RegisterStep = 1 | 2 | 3
type StatusTone = 'default' | 'error'

const initialPasswordValues: RegisterPasswordStepValues = {
  verificationCode: '',
  password: '',
  confirmPassword: '',
}

const initialProfileValues: RegisterProfileStepValues = {
  nickname: '',
  department: '',
  grade: '',
}

export function RegisterPage() {
  const navigate = useNavigate()
  const [currentStep, setCurrentStep] = useState<RegisterStep>(1)
  const [countdown, setCountdown] = useState(0)
  const [registerEmail, setRegisterEmail] = useState('')
  const [passwordValues, setPasswordValues] =
    useState<RegisterPasswordStepValues>(initialPasswordValues)
  const [profileValues, setProfileValues] =
    useState<RegisterProfileStepValues>(initialProfileValues)
  const [grades, setGrades] = useState<string[]>([])
  const [isMetaLoading, setIsMetaLoading] = useState(true)
  const [statusMessage, setStatusMessage] = useState('')
  const [statusTone, setStatusTone] = useState<StatusTone>('default')
  const [toastNotice, setToastNotice] = useState('')

  useEffect(() => {
    if (countdown <= 0) {
      return
    }

    const timer = window.setTimeout(() => {
      setCountdown((value) => value - 1)
    }, 1000)

    return () => window.clearTimeout(timer)
  }, [countdown])

  useEffect(() => {
    let cancelled = false

    const loadMeta = async () => {
      setIsMetaLoading(true)

      try {
        const gradeList = await fetchGrades()

        if (cancelled) {
          return
        }

        setGrades(gradeList)
      } catch (error) {
        if (cancelled) {
          return
        }

        setStatusTone('error')
        setStatusMessage(getErrorMessage(error, '注册选项加载失败，请刷新后重试'))
      } finally {
        if (!cancelled) {
          setIsMetaLoading(false)
        }
      }
    }

    void loadMeta()

    return () => {
      cancelled = true
    }
  }, [registerEmail])

  const startCountdown = (seconds?: number) => {
    setCountdown(seconds && seconds > 0 ? seconds : VERIFICATION_CODE_COUNTDOWN_SECONDS)
  }

  const setErrorStatus = (error: unknown, fallback: string) => {
    const message = getErrorMessage(error, fallback)
    setStatusTone('error')
    setStatusMessage('')
    setToastNotice(message)
  }

  const setDefaultStatus = (message: string) => {
    setStatusTone('default')
    setStatusMessage(message)
  }

  const handleEmailContinue = async (values: RegisterEmailStepValues) => {
    try {
      const response = await sendRegisterCode(values.email)
      setRegisterEmail(values.email)
      setPasswordValues(initialPasswordValues)
      setProfileValues(initialProfileValues)
      setCurrentStep(2)
      setDefaultStatus('验证码已发送，请查收邮箱。')
      startCountdown(response.cooldownSeconds)
    } catch (error) {
      setErrorStatus(error, '验证码发送失败，请稍后重试')
    }
  }

  const handlePasswordContinue = async (values: RegisterPasswordStepValues) => {
    try {
      await verifyRegisterCode(registerEmail, values.verificationCode)
      setPasswordValues(values)
      setCurrentStep(3)
      setDefaultStatus('验证码校验成功，请继续完善资料。')
    } catch (error) {
      setErrorStatus(error, '验证码校验失败，请重试')
    }
  }

  const handleProfileComplete = async (values: RegisterProfileStepValues) => {
    setProfileValues(values)

    try {
      const result = await registerAccount({
        email: registerEmail,
        code: passwordValues.verificationCode,
        password: passwordValues.password,
        nickname: values.nickname,
        department: values.department,
        grade: values.grade,
      })

      saveAuthenticatedUser(result.token, result.user)
      resetDashboardAnnouncementSeen()
      navigate(connectionPaths.root)
    } catch (error) {
      setErrorStatus(error, '注册失败，请稍后重试')
    }
  }

  const handleResendCode = async () => {
    try {
      const response = await sendRegisterCode(registerEmail)
      setDefaultStatus('验证码已重新发送，请查收邮箱。')
      startCountdown(response.cooldownSeconds)
    } catch (error) {
      setErrorStatus(error, '验证码发送失败，请稍后重试')
    }
  }

  useEffect(() => {
    if (!toastNotice) {
      return
    }

    const timer = window.setTimeout(() => {
      setToastNotice('')
    }, 3000)

    return () => window.clearTimeout(timer)
  }, [toastNotice])

  const noticeClass =
    statusTone === 'error' ? 'text-sm text-[#b42318]' : 'text-sm text-[#71717a]'

  return (
    <>
      <FloatingToast message={toastNotice} />

      <AuthSplitLayout
        subtitle="验证清华身份，保留你的选择权，然后用一句话开始寻找值得认识的人。"
        title="加入 SouLink，让 AI 帮你更好地认识人"
      >
        {currentStep === 1 ? (
          <div className="space-y-6">
            <RegisterForm
              defaultEmail={registerEmail}
              onContinue={handleEmailContinue}
            />
            {statusMessage ? <p className={noticeClass}>{statusMessage}</p> : null}
          </div>
        ) : null}

        {currentStep === 2 ? (
          <div className="space-y-6">
            <RegisterPasswordStep
              countdown={countdown}
              defaultValues={passwordValues}
              email={registerEmail}
              onBack={() => {
                setCurrentStep(1)
                setStatusMessage('')
              }}
              onContinue={handlePasswordContinue}
              onResend={handleResendCode}
            />
            {statusMessage ? <p className={noticeClass}>{statusMessage}</p> : null}
          </div>
        ) : null}

        {currentStep === 3 ? (
          <div className="space-y-6">
            <RegisterProfileStep
              defaultValues={profileValues}
              grades={grades}
              isOptionsLoading={isMetaLoading}
              onBack={() => {
                setCurrentStep(2)
                setStatusMessage('')
              }}
              onComplete={handleProfileComplete}
            />
            {statusMessage ? <p className={noticeClass}>{statusMessage}</p> : null}
          </div>
        ) : null}
      </AuthSplitLayout>
    </>
  )
}
