import type {
  SurveyMultiSelectQuestion,
  SurveyNumberQuestion,
  SurveyOption,
  SurveyRangeQuestion,
  SurveyScaleQuestion,
  SurveySection,
  SurveySingleChoiceQuestion,
} from '@/types/survey'

function option(label: string): SurveyOption {
  return { label, value: label }
}

function singleQuestion(
  id: string,
  number: number,
  title: string,
  labels: string[],
  required = true,
): SurveySingleChoiceQuestion {
  return {
    id,
    kind: 'single',
    number,
    title,
    required,
    importantEligible: true,
    options: labels.map(option),
  }
}

function numberQuestion(
  id: string,
  number: number,
  title: string,
  config: {
    min: number
    max: number
    unit?: string
    placeholder: string
    required?: boolean
  },
): SurveyNumberQuestion {
  return {
    id,
    kind: 'number',
    number,
    title,
    required: config.required ?? true,
    importantEligible: true,
    min: config.min,
    max: config.max,
    unit: config.unit,
    placeholder: config.placeholder,
  }
}

function rangeQuestion(
  id: string,
  number: number,
  title: string,
  config: {
    min: number
    max: number
    step: number
    unit?: string
    lowerLabel: string
    upperLabel: string
    defaultMin: number
    defaultMax: number
    allowUnlimited?: boolean
    required?: boolean
    note?: string
  },
): SurveyRangeQuestion {
  return {
    id,
    kind: 'range',
    number,
    title,
    required: config.required ?? true,
    importantEligible: true,
    note: config.note,
    min: config.min,
    max: config.max,
    step: config.step,
    unit: config.unit,
    lowerLabel: config.lowerLabel,
    upperLabel: config.upperLabel,
    allowUnlimited: config.allowUnlimited,
    defaultValue: {
      min: config.defaultMin,
      max: config.defaultMax,
    },
  }
}

function scaleQuestion(
  id: string,
  number: number,
  title: string,
  leftLabel: string,
  rightLabel: string,
): SurveyScaleQuestion {
  return {
    id,
    kind: 'scale',
    number,
    title,
    required: true,
    importantEligible: true,
    min: 1,
    max: 7,
    leftLabel,
    rightLabel,
  }
}

function multiQuestion(
  id: string,
  number: number,
  title: string,
  options: string[],
  minSelections: number,
  maxSelections: number,
  required = true,
): SurveyMultiSelectQuestion {
  return {
    id,
    kind: 'multi',
    number,
    title,
    required,
    importantEligible: true,
    minSelections,
    maxSelections,
    groups: [
      {
        label: '选项',
        options: options.map(option),
      },
    ],
  }
}

const genderOptions = ['男', '女', '非二元性别']

const identityOptions = ['本科生', '硕士生', '博士生', '已毕业校友', '教职工/研究人员']

const stageOptions = [
  '本科低年级（大一/大二）',
  '本科高年级（大三/大四）',
  '硕士阶段',
  '博士阶段',
  '已毕业/工作',
]

const campusOptions = ['校本部', '五道口金融学院', '深圳国际研究生院', '其他']

const attractionTraitOptions = [
  '真诚',
  '聪明',
  '幽默',
  '温柔',
  '自律',
  '稳定',
  '有野心',
  '有审美',
  '有担当',
  '好奇心强',
  '善于倾听',
  '有自己的热爱',
]

const heartbeatMomentOptions = [
  '被认真倾听',
  '被坚定选择',
  '被温柔照顾',
  '被欣赏和夸奖',
  '一起做成一件事',
  '一起安静待着也不尴尬',
]

const dietaryRestrictionOptions = ['无', '清真', '素食', '不吃辣', '对酒精敏感', '其他']

const stableInterestOptions = [
  '阅读',
  '电影 / 剧集',
  '音乐 / 演出',
  '健身 / 跑步',
  '球类运动',
  '户外 / 徒步',
  '游戏',
  '摄影 / 视频',
  '展览 / 博物馆',
  '做饭 / 探店',
  '编程 / 科技',
  '学术讨论',
  '创作 / 写作',
  '旅行',
]

const safetyExpressionOptions = [
  '稳定回复和出现',
  '明确表达喜欢和在意',
  '主动规划见面',
  '冲突时不失联',
  '愿意介绍自己的朋友和生活',
  '尊重我的边界',
]

const likeExpressionOptions = [
  '主动找话题聊天',
  '记住细节并回应',
  '安排见面',
  '直接表达欣赏',
  '默默付出和帮忙',
  '分享我的真实生活',
]

const bottomLineOptions = [
  '不真诚',
  '边界感差',
  '情绪不稳定且拒绝沟通',
  '不忠诚',
  '对未来完全没有规划',
  '价值观冲突过大',
  '不尊重他人',
]

export const surveySections: SurveySection[] = [
  {
    id: 'outline',
    title: '基本信息',
    subtitle: '知人者智，自知者明。\n——《道德经》',
    description: '这一部分先筛选基础条件，再决定哪些连接值得继续靠近。',
    tip: '每一道题都可以标记为“对我更重要”，系统会在匹配解释里更认真参考这些题目。',
    questions: [
      singleQuestion('q1', 1, '我的性别是', genderOptions),
      singleQuestion('q2', 2, '我希望匹配到的对象性别是', genderOptions),
      singleQuestion('q3', 3, '我来这里最主要是想', [
        '寻找认真恋爱对象',
        '先认真认识，再决定是否发展关系',
        '寻找高质量新朋友',
      ]),
      singleQuestion('q4', 4, '我现在的身份是', identityOptions),
      singleQuestion('q5', 5, '我现在所处的阶段是', stageOptions),
      singleQuestion('q6', 6, '我平时主要在', campusOptions),
      numberQuestion('q7', 7, '我的年龄是（范围：18-40岁）', {
        min: 18,
        max: 40,
        placeholder: '请输入年龄',
      }),
      rangeQuestion('q8', 8, '我能接受的对方年龄范围', {
        min: 18,
        max: 40,
        step: 1,
        lowerLabel: '最低年龄',
        upperLabel: '最高年龄',
        defaultMin: 20,
        defaultMax: 28,
        allowUnlimited: true,
        note: '可切换为“不限”',
      }),
      numberQuestion('q9', 9, '我的身高是（范围：140-210cm）', {
        min: 140,
        max: 210,
        placeholder: '请输入身高（cm）',
      }),
      rangeQuestion('q10', 10, '我能接受的对方身高范围', {
        min: 140,
        max: 210,
        step: 1,
        unit: 'cm',
        lowerLabel: '最低身高',
        upperLabel: '最高身高',
        defaultMin: 160,
        defaultMax: 185,
      }),
      singleQuestion('q11', 11, '未来一到三年，我更可能', [
        '大概率继续留在北京',
        '可能去其他一线城市',
        '可能出国/出境',
        '暂时不确定',
      ]),
      scaleQuestion('q12', 12, '异地关系对我来说', '基本不能接受', '可以认真尝试'),
    ],
  },
  {
    id: 'attraction',
    title: '吸引力与预期',
    subtitle: '有美一人，清扬婉兮。\n——《诗经·野有蔓草》',
    description: '这里聚焦的是你会被什么吸引，以及你希望自己被如何感受到。',
    tip: '如果某道题对你的判断特别关键，记得标记为“对我更重要”。',
    questions: [
      scaleQuestion('q13', 13, '在社交场合里，我更像', '很外向，容易主动打开局面', '偏内向，更喜欢安静和小范围交流'),
      scaleQuestion('q14', 14, '哪种气质更容易吸引我', '温和亲近、让人放松', '清冷克制、边界感强'),
      scaleQuestion('q15', 15, '我会因为对方“认真做事”而心动', '基本不会', '非常会'),
      scaleQuestion('q16', 16, '我更欣赏有表达欲的人', '更喜欢安静克制', '更喜欢鲜明外放'),
      multiQuestion('q17', 17, '最吸引我的特质是', attractionTraitOptions, 2, 5),
      multiQuestion('q18', 18, '我希望别人先感受到我的', attractionTraitOptions, 2, 5),
      scaleQuestion(
        'q19',
        20,
        '对我来说，一个人在日常状态里的审美和打理程度，会明显影响我对这段关系的愉悦感和心动感',
        '极度不认同',
        '极度认同',
      ),
      multiQuestion('q20', 20, '一段关系里，最让我心动的瞬间通常是', heartbeatMomentOptions, 1, 3),
    ],
  },
  {
    id: 'getting-along',
    title: '生活方式',
    subtitle: '琴瑟在御，莫不静好。\n——《诗经·女曰鸡鸣》',
    description: '这一部分更关注你的生活节奏、习惯和生活感。',
    tip: '这里没有标准答案，重点在于找到更适合彼此日常相处的人。',
    questions: [
      scaleQuestion('q21', 21, '我的作息更偏向', '夜猫子', '早起型'),
      scaleQuestion('q22', 22, '现在的我，生活节奏更偏向', '松弛随性', '计划明确、效率优先'),
      singleQuestion('q23', 23, '如果周末约会，我更偏好', [
        '校园散步 / 自习 / 吃饭',
        '咖啡馆 / 展览 / City Walk',
        '运动 / 户外活动',
        '去市区认真约会',
      ]),
      singleQuestion('q24', 24, '我希望恋爱中的见面频率大约是', [
        '几乎每天',
        '一周 2-3 次',
        '一周 1 次',
        '两周及以上 1 次',
      ]),
      scaleQuestion('q25', 25, '对我来说，关系里“各自有空间”这件事', '非常重要', '更想高频陪伴'),
      scaleQuestion('q26', 26, '我对整洁和秩序的要求', '比较随意', '比较在意'),
      scaleQuestion('q27', 27, '我的消费习惯更偏向', '理性节制，能省则省', '愿意为体验和品质买单'),
      multiQuestion('q28', 28, '饮食上我有明确限制', dietaryRestrictionOptions, 1, 6),
      singleQuestion('q29', 29, '关于吸烟/电子烟，我的情况是', ['从不', '偶尔', '经常']),
      scaleQuestion('q30', 30, '??????/?????????', '??????', '????'),
      singleQuestion('q31', 31, '关于饮酒，我更接近', ['基本不喝', '社交场合偶尔喝', '比较常喝']),
      multiQuestion('q32', 32, '选出 3-6 个你最稳定、最真实的兴趣', stableInterestOptions, 3, 6),
      scaleQuestion('q33', 33, '我希望和对方的兴趣重合度', '互补更好', '高度一致更好'),
    ],
  },
  {
    id: 'closeness',
    title: '情感风格',
    subtitle: '言念君子，温其如玉。\n——《诗经·小戎》',
    description: '这一部分会帮助我们理解你的依恋节奏、沟通方式和安全感来源。',
    tip: '如果你对某种相处方式特别敏感，这一部分会非常重要。',
    questions: [
      scaleQuestion('q34', 34, '如果对方几个小时没回消息，我通常', '完全不受影响', '会明显在意'),
      scaleQuestion('q35', 35, '当我感到不安时，我更容易', '先自己消化', '希望马上和对方确认'),
      scaleQuestion('q36', 36, '在亲密关系里展露脆弱，对我来说', '很难', '比较自然'),
      scaleQuestion('q37', 37, '如果发生矛盾，我的第一反应通常是', '先冷静一下', '希望立刻说清楚'),
      scaleQuestion('q38', 38, '和好这件事上，我通常', '更希望对方先来找我', '我更愿意先迈出一步'),
      scaleQuestion('q39', 39, '通常倾向于在哪个阶段与伴侣发生性关系？', '第一次约会', '确认长期承诺关系后'),
      scaleQuestion('q40', 40, '在关系里，我更习惯', '更常照顾别人', '更希望被照顾'),
      scaleQuestion('q41', 41, '如果对方有关系很好的异性朋友，我会', '比较介意', '比较信任'),
      multiQuestion('q42', 42, '让我更有安全感的表达方式是', safetyExpressionOptions, 1, 3),
      multiQuestion('q43', 43, '当我喜欢一个人时，我更可能这样表达', likeExpressionOptions, 1, 3),
    ],
  },
  {
    id: 'future',
    title: '未来规划',
    subtitle: '执子之手，与子偕老。\n——《诗经·击鼓》',
    description: '最后这部分更关注长期取向、人生规划和关键价值底线。',
    tip: '这里不一定追求一致，但分歧的大小会直接影响长期关系的稳定性。',
    questions: [
      scaleQuestion('q44', 44, '现在的我，更偏向', '更重视生活平衡', '更追求极致投入和结果'),
      scaleQuestion('q45', 45, '我希望对方在学业/事业上', '松弛稳定就好', '越上进越好'),
      scaleQuestion('q46', 46, '毕业或工作后，我更想要的生活是', '节奏慢、生活稳', '留在高密度城市继续拼'),
      scaleQuestion('q47', 47, '关于婚姻，我目前的态度', '没有明确打算', '是重要人生目标'),
      scaleQuestion('q48', 48, '关于未来要不要孩子，我现在更偏向', '倾向不要', '倾向要'),
      scaleQuestion('q49', 49, '当“现实利益”和“做人原则”冲突时，我更看重', '现实结果', '原则和底线'),
      singleQuestion('q50', 50, '我希望亲密关系更像', [
        '两个独立的人并肩同行',
        '可以互相依赖、深度绑定',
        '视具体阶段动态调整',
      ]),
      singleQuestion('q51', 51, '在重大人生选择上，我更希望伴侣', [
        '尊重彼此独立决定',
        '一起讨论后共同决定',
        '关键时刻以关系整体优先',
      ]),
      multiQuestion('q52', 52, '我最不能妥协的长期关系底线是', bottomLineOptions, 1, 3),
    ],
  },
]


