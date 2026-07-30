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
      imageAlt="清华校园"
      imageSources={loginHeroImage}
      imageSrc={loginHeroImage.webp}
      subtitle="In Tsinghua, meet with a reason."
      title="找到下一位值得认识的人"
    >
      <LoginForm />
    </AuthSplitLayout>
  )
}
