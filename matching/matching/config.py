IMPORTANCE_MULTIPLIER = 1.8
MIN_ACCEPTABLE_SCORE = 0.40

DIMENSION_WEIGHTS = {
    "D1": 0.08,
    "D2": 0.07,
    "D3": 0.15,
    "D4": 0.20,
    "D5": 0.10,
    "D6": 0.20,
    "D7": 0.20,
}

# Q3 purpose compatibility matrix
PURPOSE_MATRIX = {
    "寻找认真恋爱对象":    {"寻找认真恋爱对象": 1.0, "先认真认识，再决定是否发展关系": 0.7, "寻找高质量新朋友": 0.2},
    "先认真认识，再决定是否发展关系": {"寻找认真恋爱对象": 0.7, "先认真认识，再决定是否发展关系": 1.0, "寻找高质量新朋友": 0.6},
    "寻找高质量新朋友":    {"寻找认真恋爱对象": 0.2, "先认真认识，再决定是否发展关系": 0.6, "寻找高质量新朋友": 1.0},
}

# Q6 university grouping for proximity scoring
# Universities in the same group have higher proximity scores
UNIVERSITY_GROUPS = {
    "清华大学": "beijing_top",
    "北京大学": "beijing_top",
    "北大大学": "beijing_top",  # v0.5 legacy name
    "中国人民大学": "beijing_985",
    "北京师范大学": "beijing_985",
    "中国传媒大学": "beijing_other",
    "北京外国语大学": "beijing_other",
    "上海交通大学": "shanghai",
    "复旦大学": "shanghai",
    "新加坡国立大学": "singapore",
    "南洋理工大学": "singapore",
}

# Email domain -> university name (for hard filter on Q6)
EMAIL_DOMAIN_TO_UNIVERSITY = {
    "mails.tsinghua.edu.cn": "清华大学",
    "mail.tsinghua.edu.cn": "清华大学",
    "tsinghua.edu.cn": "清华大学",
    "tsinghua.org.cn": "清华大学",
    "stu.pku.edu.cn": "北京大学",
    "pku.edu.cn": "北京大学",
    "ruc.edu.cn": "中国人民大学",
    "mails.cuc.edu.cn": "中国传媒大学",
    "cuc.edu.cn": "中国传媒大学",
    "mail.bnu.edu.cn": "北京师范大学",
    "bfsu.edu.cn": "北京外国语大学",
    "student.pumc.edu.cn": "北京协和医学院",
}

# Aliases: different names for the same university (v0.5 used "北大大学", v0.6 uses "北京大学")
UNIVERSITY_ALIASES = {
    "北大大学": "北京大学",
}

# Q5 stage ordering
STAGE_ORDER = {
    "本科低年级（大一/大二）": 1,
    "本科高年级（大三/大四）": 2,
    "硕士阶段": 3,
    "博士阶段": 4,
    "已毕业/工作": 5,
}

# Q11 future location ordering
FUTURE_LOC_ORDER = {
    "大概率继续留在北京": 1,
    "可能去其他一线城市": 2,
    "可能去二三线城市": 3,
    "可能去乡村地区": 4,
    "可能出国/出境": 5,
    "暂时不确定": None,
}

# Q24 meeting frequency ordering
FREQ_ORDER = {
    "几乎每天": 4,
    "一周 2-3 次": 3,
    "一周 1 次": 2,
    "两周及以上 1 次": 1,
}

# Q29 smoking intensity mapping
SMOKING_INTENSITY = {
    "从不": 0,
    "偶尔": 3,
    "经常": 6,
}

# Q31 drinking ordering
DRINK_ORDER = {
    "基本不喝": 1,
    "社交场合偶尔喝": 2,
    "比较常喝": 3,
}

# Q50 relationship model matrix
REL_MODEL_MATRIX = {
    "两个独立的人并肩同行":    {"两个独立的人并肩同行": 1.0, "可以互相依赖、深度绑定": 0.3, "视具体阶段动态调整": 0.7},
    "可以互相依赖、深度绑定":  {"两个独立的人并肩同行": 0.3, "可以互相依赖、深度绑定": 1.0, "视具体阶段动态调整": 0.7},
    "视具体阶段动态调整":     {"两个独立的人并肩同行": 0.7, "可以互相依赖、深度绑定": 0.7, "视具体阶段动态调整": 1.0},
}

# Q51 decision style matrix
DECISION_MATRIX = {
    "尊重彼此独立决定":    {"尊重彼此独立决定": 1.0, "一起讨论后共同决定": 0.6, "关键时刻以关系整体优先": 0.3},
    "一起讨论后共同决定":  {"尊重彼此独立决定": 0.6, "一起讨论后共同决定": 1.0, "关键时刻以关系整体优先": 0.7},
    "关键时刻以关系整体优先": {"尊重彼此独立决定": 0.3, "一起讨论后共同决定": 0.7, "关键时刻以关系整体优先": 1.0},
}

# Q42 ↔ Q43 semantic mapping matrix
# Q42 options (rows): what makes me feel secure
# Q43 options (cols): how I express love
SECURITY_EXPRESSION_MAP = {
    "稳定回复和出现":       {"主动找话题聊天": 0.8, "记住细节并回应": 0.5, "安排见面": 0.3, "直接表达欣赏": 0.2, "默默付出和帮忙": 0.4, "分享我的真实生活": 0.3},
    "明确表达喜欢和在意":    {"主动找话题聊天": 0.3, "记住细节并回应": 0.3, "安排见面": 0.2, "直接表达欣赏": 0.9, "默默付出和帮忙": 0.2, "分享我的真实生活": 0.4},
    "主动规划见面":        {"主动找话题聊天": 0.2, "记住细节并回应": 0.3, "安排见面": 1.0, "直接表达欣赏": 0.2, "默默付出和帮忙": 0.3, "分享我的真实生活": 0.2},
    "冲突时不失联":        {"主动找话题聊天": 0.5, "记住细节并回应": 0.2, "安排见面": 0.2, "直接表达欣赏": 0.3, "默默付出和帮忙": 0.3, "分享我的真实生活": 0.3},
    "愿意介绍自己的朋友和生活": {"主动找话题聊天": 0.2, "记住细节并回应": 0.2, "安排见面": 0.2, "直接表达欣赏": 0.2, "默默付出和帮忙": 0.3, "分享我的真实生活": 0.9},
    "尊重我的边界":        {"主动找话题聊天": 0.2, "记住细节并回应": 0.6, "安排见面": 0.2, "直接表达欣赏": 0.2, "默默付出和帮忙": 0.5, "分享我的真实生活": 0.3},
}

# Likert scale questions
LIKERT_QUESTIONS = {12, 13, 14, 15, 16, 19, 21, 22, 25, 26, 27, 30, 33, 34, 35, 36, 37, 38, 39, 40, 41, 44, 45, 46, 47, 48, 49}

# Multi-select questions
MULTI_SELECT_QUESTIONS = {6, 17, 18, 20, 23, 28, 32, 42, 43, 52}

# Numeric questions
NUMERIC_QUESTIONS = {7, 9}

# Range (JSON) questions
RANGE_QUESTIONS = {8, 10}
