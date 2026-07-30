import {
  BadgeCheck,
  CalendarSync,
  MessagesSquare,
  ScanSearch,
} from 'lucide-react'

import type {
  FaqItem,
  FeatureItem,
  FlowStep,
  HeroStat,
  NavItem,
} from '@/types/landing'

export const navigationItems: NavItem[] = [
  { label: '关于', href: '/#features', kind: 'route' },
  { label: 'FAQ', href: '/#faq', kind: 'route' },
]

export const heroRegistrationCount = 10000

export const heroStats: HeroStat[] = []

export const featureItems: FeatureItem[] = [
  {
    title: '深度问卷',
    description:
      '从生活节奏、沟通方式到关系期待，问题设计覆盖长期相处中真正重要的差异与共识。',
    icon: ScanSearch,
  },
  {
    title: '校园认证',
    description:
      '以学校邮箱和身份信息为基础建立可信环境，让每一次匹配都建立在真实校园场景中。',
    icon: BadgeCheck,
  },
  {
    title: '定时匹配',
    description:
      '平台会在每周四晚统一进入揭晓窗口，20:00 至 21:00 之间陆续同步结果，减少无尽刷新和即时焦虑。',
    icon: CalendarSync,
  },
  {
    title: '匹配解释',
    description:
      '不仅给出结果，还会说明相似点、互补项与适合的破冰方向，帮助你更自然地迈出第一步。',
    icon: MessagesSquare,
  },
]

export const flowSteps: FlowStep[] = [
  {
    title: '创建账号',
    description: '使用校园身份进入平台，完成基础资料与展示偏好设置。',
  },
  {
    title: '填写问卷',
    description: '分步骤完成结构化问卷，并在每周四 14:00 前提交或修改你的答案。',
  },
  {
    title: '等待匹配',
    description: '每周四晚 20:00 至 21:00 进入结果揭晓窗口，系统会陆续同步本轮结果。',
  },
  {
    title: '查看结果',
    description: '获得一位高相关对象，并看到清晰的匹配理由与建议切入点。',
  },
]

export const faqItems: FaqItem[] = [
  {
    question: '谁可以加入 Soulink？',
    answer:
    // '目前仅开放给完成清华和北大校内邮箱验证的校内学生、教职工、校友。我们希望先认真守护社区的边界，让真诚的灵魂，更容易彼此靠近。',
    '目前，SouLink 高校联盟已支持所有已入驻 SouLink 的高校邮箱。若您的高校希望加入 SouLink 高校联盟，欢迎联系我们：thusoulink@163.com，并附上您所在高校的邮箱后缀信息。'
  },
  {
    question: '信息安全吗？',
    answer:
      '我们认真对待每一份信任。你的信息将被妥善保护，仅用于必要的匹配，不会被用于无关用途。我们也会持续完善系统防护与安全设计，尽力守护你的信息安全。',
  },
  {
    question: '匹配结果会在什么时候揭晓？',
    answer:
      '我们会在每周四晚 20:00-21:00 统一揭晓本轮结果。加入本轮的用户，将在这一时段内收到属于自己的匹配通知。',
  },
  {
    question: '我必须每周都参加吗？',
    answer:
      '不需要。你可以按自己的节奏，选择哪一轮加入。好的相遇，不必仓促，只需在恰好的时刻发生。比起频繁相遇，我们更在意你愿意认真奔赴。',
  },
]
