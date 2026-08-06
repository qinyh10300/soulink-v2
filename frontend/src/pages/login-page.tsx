import { AuthSplitLayout } from '@/components/auth/auth-split-layout'
import { LoginForm } from '@/components/auth/login-form'

export function LoginPage() {
  return (
    <AuthSplitLayout
      subtitle="从饭搭子、学习伙伴到项目合伙人，告诉 AI 你此刻为什么想认识一个人。"
      title="从一个真实意图，走向一次真实连接"
    >
      <LoginForm />
    </AuthSplitLayout>
  )
}
