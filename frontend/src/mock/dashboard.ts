import type { DashboardContent } from '@/types/dashboard'

import {
  getNextMatchRevealStart,
  MATCH_REVEAL_WINDOW_TEXT,
} from '@/lib/schedule'

export const dashboardContent: DashboardContent = {
  user: {
    name: 'Soulink User',
    avatarEmoji: '🌷',
  },
  countdown: {
    label: '距离下次匹配结果揭晓',
    description: MATCH_REVEAL_WINDOW_TEXT,
    targetAt: getNextMatchRevealStart(),
  },
  statuses: [
    {
      id: 'survey',
      title: '问卷状态',
      value: '已完成',
      tone: 'success',
    },
    {
      id: 'weekly-match',
      title: '本周匹配',
      value: '等待揭晓',
      tone: 'active',
      actionLabel: '查看结果',
    },
    {
      id: 'soul-portrait',
      title: '灵魂画像',
      value: '',
      tone: 'neutral',
      description: '基于问卷回答生成你的个性画像与关系风格，帮助你更快理解自己的连接偏好。',
    },
  ],
  features: [],
  weeklyMockResult: {
    cycleLabel: '第2期匹配',
    avatarEmoji: '🐱',
    name: '转运大西瓜',
    department: '医学院',
    grade: '博一',
    score: 81,
    matchReasons: [
      '她来自重庆，你来自贵州，两人对辣味的热爱是天然的共同话题，一起寻找周边地道的辣味餐厅会是不错的开始。',
      '你们都把编程和球类运动列为核心兴趣，这让约会选择可以从安静的代码交流延伸到充满活力的运动场。',
      '你期待伴侣聪明、富有创造力，而她的自我期待与你的择偶标准高度重合，这为深入交流提供了基础。',
    ],
    suggestedTopics: ['闵行周边的辣味餐厅推荐', '一起约球或看球赛', '各自的编程项目经历'],
    cautions: [
      '你更偏向顺从配合，她更愿意主动推进，前期可以把节奏和预期说清楚，避免一方推进过快。',
      '共同兴趣很多，但第一次聊天不必一次铺太满，先围绕一个具体场景展开，会更自然。',
    ],
    contactLabel: '邮箱',
    contactValue: 'xigua-match@pku.edu.cn',
  },
  history: {
    title: '还没有历史匹配记录',
    description: '完成问卷并参与每周匹配后，过往结果会沉淀在这里，方便你随时回看。',
    icon: '🕰️',
  },
}
