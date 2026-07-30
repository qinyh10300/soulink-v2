import bfsuHeroImage from '@/assets/bfsu.jpg'
import bnuHeroImage from '@/assets/bnu.png'
import cucHeroImage from '@/assets/cuc.jpg'
import pkuHeroImage from '@/assets/pku.jpg'
import schoolHeroImage from '@/assets/school.jpg'
import thuHeroImage from '@/assets/thu.png'
import bnuStory1 from '@/assets/story/BNU-1.png'
import bnuStory2 from '@/assets/story/BNU-2.png'
import bnuStory3 from '@/assets/story/BNU-3.png'
import bnuStory4 from '@/assets/story/BNU-4.png'
import bfsuStory1 from '@/assets/story/BFSU-1.png'
import bfsuStory2 from '@/assets/story/BFSU-2.png'
import bfsuStory3 from '@/assets/story/BFSU-3.png'
import bfsuStory4 from '@/assets/story/BFSU-4.png'
import cucStory1 from '@/assets/story/CUC-1.jpg'
import cucStory2 from '@/assets/story/CUC-2.jpg'
import cucStory3 from '@/assets/story/CUC-3.jpg'
import cucStory4 from '@/assets/story/CUC-4.jpg'
import pkuStory1 from '@/assets/story/PKU-1.png'
import pkuStory2 from '@/assets/story/PKU-2.png'
import pkuStory3 from '@/assets/story/PKU-3.png'
import pkuStory4 from '@/assets/story/PKU-4.png'
import thuStory1 from '@/assets/story/THU-1.png'
import thuStory2 from '@/assets/story/THU-2.png'
import thuStory3 from '@/assets/story/THU-3.png'
import thuStory4 from '@/assets/story/THU-4.png'

export type SchoolKey = 'thu' | 'pku' | 'bnu' | 'ruc' | 'cuc' | 'bfsu'

export interface SchoolConfig {
  key: SchoolKey
  shortName: string
  name: string
  englishName: string
  locationLabel: string
  title: string
  tagline: string
  summary: string
  accentClassName: string
  glowClassName: string
  heroImage: string
  storyImages: string[]
  poemLines: string[]
  stats: Array<{ label: string; value: string }>
}

export const schoolConfigs: Record<SchoolKey, SchoolConfig> = {
  thu: {
    key: 'thu',
    shortName: '清华',
    name: '清华大学',
    englishName: 'Tsinghua University',
    locationLabel: '清华路·情人坡',
    title: 'THUers',
    tagline: '水木之间，灵魂相逢',
    summary:
      '从图书馆到操场，从深夜灯火到晨雾微光，Soulink 在清华想做的，是让认真表达和认真靠近重新变得浪漫。',
    accentClassName: 'from-[#6d8bb4] via-[#b9c8e4] to-[#efe6da]',
    glowClassName: 'from-[#9db5d8]/50 via-[#e7d9c4]/20 to-transparent',
    heroImage: thuHeroImage,
    storyImages: [thuStory1, thuStory2, thuStory3, thuStory4],
    poemLines: [
      '东操有风，西荷塘有月，',
      '北馆灯亮，南园声未歇。',
      '光影流转，人群往复，',
      '心意明灭，却总想遇见更懂的人。',
    ],
    stats: [
      { label: '关键词', value: '慢热 / 真诚 / 长线陪伴' },
      { label: '场景感', value: '图书馆、操场、情人坡' },
    ],
  },
  pku: {
    key: 'pku',
    shortName: '北大',
    name: '北京大学',
    englishName: 'Peking University',
    locationLabel: '未名湖畔·燕园夜色',
    title: 'PKUers',
    tagline: '未名湖畔，灵魂相逢',
    summary:
      '从燕园树影到湖边夜色，Soulink 想把北大的相遇写成一种更安静也更有分量的靠近方式。',
    accentClassName: 'from-[#6f6b8f] via-[#d4bfd1] to-[#f5ece2]',
    glowClassName: 'from-[#a79dc2]/45 via-[#e8d8e3]/18 to-transparent',
    heroImage: pkuHeroImage,
    storyImages: [pkuStory1, pkuStory2, pkuStory3, pkuStory4],
    poemLines: [
      '塔影临湖，书灯映夜，',
      '燕园有风，行人不止。',
      '四时流转，草木自深，',
      '人海之中，也想更认真地遇见彼此。',
    ],
    stats: [
      { label: '关键词', value: '克制 / 诗意 / 彼此理解' },
      { label: '场景感', value: '未名湖、博雅塔、燕园夜路' },
    ],
  },
  bnu: {
    key: 'bnu',
    shortName: '北师大',
    name: '北京师范大学',
    englishName: 'Beijing Normal University',
    locationLabel: '京师路·木铎金声',
    title: 'BNUers',
    tagline: '京师学脉，理想相续',
    summary:
      '从林荫小径到教学楼灯光，Soulink 想把北师大的相遇写成更温和、更稳定，也更愿意倾听的靠近方式。',
    accentClassName: 'from-[#7e9a86] via-[#d8e4d3] to-[#f3ece2]',
    glowClassName: 'from-[#93b39e]/45 via-[#e1e9dc]/18 to-transparent',
    heroImage: bnuHeroImage,
    storyImages: [bnuStory1, bnuStory2, bnuStory3, bnuStory4],
    poemLines: [
      '京师有风，木铎有声，',
      '银杏渐黄，灯火渐明。',
      '四季更替，人海往复，',
      '也愿在此，慢慢遇见同频的人。',
    ],
    stats: [
      { label: '关键词', value: '温和 / 稳定 / 倾听感' },
      { label: '场景感', value: '林荫道、教学楼、秋日校园' },
    ],
  },
  ruc: {
    key: 'ruc',
    shortName: '人大',
    name: '中国人民大学',
    englishName: 'Renmin University of China',
    locationLabel: '人大·人文广场',
    title: 'RUCers',
    tagline: '人文热望，真诚相遇',
    summary:
      '在人文与公共讨论的气息里，Soulink 想让表达清晰、价值感明确的相遇，更自然地发生。',
    accentClassName: 'from-[#a3625e] via-[#ead0c9] to-[#f7eee7]',
    glowClassName: 'from-[#d2a79e]/38 via-[#f0ddd7]/18 to-transparent',
    heroImage: schoolHeroImage,
    storyImages: [thuStory1, thuStory2, thuStory3, thuStory4],
    poemLines: [
      '书声与人声交汇，',
      '理想和日常同行。',
      '在观点碰撞之外，',
      '也想遇见愿意认真回应的人。',
    ],
    stats: [
      { label: '关键词', value: '清晰 / 热忱 / 价值共鸣' },
      { label: '场景感', value: '人文楼群、林荫小径、校园广场' },
    ],
  },
  cuc: {
    key: 'cuc',
    shortName: '中传',
    name: '中国传媒大学',
    englishName: 'Communication University of China',
    locationLabel: '中传·文化广场',
    title: 'CUCers',
    tagline: '镜头之外，认真相遇',
    summary:
      '在表达、创作与叙事交织的校园里，Soulink 想把灵感与真诚都留给更值得的相遇。',
    accentClassName: 'from-[#8b6178] via-[#ead5dd] to-[#f8efe9]',
    glowClassName: 'from-[#d8b3c3]/38 via-[#f1e0e6]/18 to-transparent',
    heroImage: cucHeroImage,
    storyImages: [cucStory1, cucStory2, cucStory3, cucStory4],
    poemLines: [
      '镜头记录光影，',
      '声音延伸想象。',
      '故事尚未落幕，人海之中，',
      '仍想遇见一人共讲。',
    ],
    stats: [
      { label: '关键词', value: '表达 / 灵感 / 共鸣感' },
      { label: '场景感', value: '演播楼、梧桐道、晚风操场' },
    ],
  },
  bfsu: {
    key: 'bfsu',
    shortName: '北外',
    name: '北京外国语大学',
    englishName: 'Beijing Foreign Studies University',
    locationLabel: '北外·西院林荫',
    title: 'BFSUers',
    tagline: '语通世界，心向彼此',
    summary:
      '在语言、文化与世界想象相互靠近的地方，Soulink 想让多元背景里的真诚连接更容易发生。',
    accentClassName: 'from-[#5c7f86] via-[#d7e5e6] to-[#f4efea]',
    glowClassName: 'from-[#a6c7cc]/40 via-[#e5eff0]/18 to-transparent',
    heroImage: bfsuHeroImage,
    storyImages: [bfsuStory1, bfsuStory2, bfsuStory3, bfsuStory4],
    poemLines: [
      '风过校园，书页轻翻，',
      '灯火深处，夜色未阑。',
      '世间语言何其千万，',
      '却仍想与你，轻声细谈。',
    ],
    stats: [
      { label: '关键词', value: '多元 / 理解 / 世界感' },
      { label: '场景感', value: '林荫道、教学楼、傍晚校园' },
    ],
  },
}

export const schoolConfigList = [
  schoolConfigs.thu,
  schoolConfigs.pku,
  schoolConfigs.bnu,
  schoolConfigs.ruc,
  schoolConfigs.cuc,
  schoolConfigs.bfsu,
]

export function isSchoolKey(value: string | undefined): value is SchoolKey {
  return value === 'thu' || value === 'pku' || value === 'bnu' || value === 'ruc' || value === 'cuc' || value === 'bfsu'
}
