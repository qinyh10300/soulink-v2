import thuHeroAvif from '@/assets/optimized/heroes/thu.avif'
import thuHeroWebp from '@/assets/optimized/heroes/thu.webp'
import { AuthSplitLayout } from '@/components/auth/auth-split-layout'
import { LoginForm } from '@/components/auth/login-form'

const loginHeroImage = {
  avif: thuHeroAvif,
  webp: thuHeroWebp,
}

export function LoginPage() {
  return (
    <AuthSplitLayout
      imageAlt=" Soulink login hero"
      imageSources={loginHeroImage}
      imageSrc={loginHeroImage.webp}
      subtitle="With All Soulinkers"
      title="探索新的AI社交范式"
    >
      <LoginForm />
    </AuthSplitLayout>
  )
}
