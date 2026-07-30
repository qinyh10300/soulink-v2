package school

import "strings"

type Name string

const (
	Tsinghua      Name = "tsinghua"
	Peking        Name = "peking"
	Renmin        Name = "renmin"
	Communication Name = "communication"
	Beiwai        Name = "beiwai"
	BeijingNormal Name = "beijing_normal"
	USTB          Name = "ustb"
	CSU           Name = "csu"
	Tongji        Name = "tongji"
	ZUEL          Name = "zuel"
	NUS           Name = "nus"
	NTU           Name = "ntu"
	ShanghaiJiao  Name = "shanghai_jiao_tong"
	Fudan         Name = "fudan"
)

var emailSuffixes = map[Name][]string{
	Tsinghua: {
		"@mails.tsinghua.edu.cn",
		"@mail.tsinghua.edu.cn",
		"@tsinghua.org.cn",
		"@mail.tsinghua.org.cn",
		"@tsinghua.edu.cn",
		"@student.pumc.edu.cn",
	},
	Peking: {
		"@pku.edu.cn",
		"@stu.pku.edu.cn",
		"@bimba.pku.edu.cn",
	},
	Renmin: {
		"@ruc.edu.cn",
	},
	Communication: {
		"@cuc.edu.cn",
		"@mails.cuc.edu.cn",
	},
	Beiwai: {
		"@mail.bfsu.edu.cn",
		"@bfsu.edu.cn",
	},
	BeijingNormal: {
		"@mail.bnu.edu.cn",
	},
	USTB: {
		"@mail.xs.ustb.edu.cn",
	},
	CSU: {
		"@csu.edu.cn",
	},
	Tongji: {
		"@tongji.edu.cn",
	},
	ZUEL: {
		"@stu.zuel.edu.cn",
	},
	NUS: {
		"@u.nus.edu",
	},
	NTU: {
		"@e.ntu.edu.sg",
	},
	Fudan: {
		"@m.fudan.edu.cn",
	},
	ShanghaiJiao: {
		"@sjtu.edu.cn",
	},
}

var displayNames = map[Name]string{
	Tsinghua:      "清华大学",
	Peking:        "北京大学",
	Renmin:        "中国人民大学",
	Communication: "中国传媒大学",
	Beiwai:        "北京外国语大学",
	BeijingNormal: "北京师范大学",
	USTB:          "北京科技大学",
	CSU:           "中南大学",
	Tongji:        "同济大学",
	ZUEL:          "中南财经政法大学",
	NUS:           "新加坡国立大学",
	NTU:           "南洋理工大学",
	ShanghaiJiao:  "上海交通大学",
	Fudan:         "复旦大学",
}

var keys = map[Name]string{
	Tsinghua:      "thu",
	Peking:        "pku",
	Renmin:        "ruc",
	Communication: "cuc",
	Beiwai:        "bfsu",
	BeijingNormal: "bnu",
	USTB:          "ustb",
	CSU:           "csu",
	Tongji:        "tju",
	ZUEL:          "zuel",
	NUS:           "nus",
	NTU:           "ntu",
	ShanghaiJiao:  "sjtu",
	Fudan:         "fdu",
}

func Names() []Name {
	return []Name{
		Tsinghua,
		Peking,
		Renmin,
		Communication,
		Beiwai,
		BeijingNormal,
		USTB,
		CSU,
		Tongji,
		ZUEL,
		NUS,
		NTU,
		ShanghaiJiao,
		Fudan,
	}
}

func EmailSuffixes(name Name) []string {
	return append([]string(nil), emailSuffixes[name]...)
}

func SupportedEmailSuffixes() []string {
	total := 0
	for _, name := range Names() {
		total += len(emailSuffixes[name])
	}
	suffixes := make([]string, 0, total)
	for _, name := range Names() {
		suffixes = append(suffixes, emailSuffixes[name]...)
	}
	return suffixes
}

func SupportsEmail(email string) bool {
	_, ok := NameByEmail(email)
	return ok
}

func NameByEmail(email string) (Name, bool) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	for _, name := range Names() {
		for _, suffix := range emailSuffixes[name] {
			if strings.HasSuffix(normalized, suffix) {
				return name, true
			}
		}
	}
	return "", false
}

func ParseName(value string) (Name, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(Tsinghua), "thu":
		return Tsinghua, true
	case string(Peking), "pku", "peking-university":
		return Peking, true
	case string(Renmin), "ruc", "renmin-university":
		return Renmin, true
	case string(Communication), "cuc", "communication-university":
		return Communication, true
	case string(Beiwai), "bfsu", "beijing-foreign-studies":
		return Beiwai, true
	case string(BeijingNormal), "bnu", "beijing-normal-university":
		return BeijingNormal, true
	case string(USTB), "university-of-science-and-technology-beijing":
		return USTB, true
	case string(CSU), "central-south-university":
		return CSU, true
	case string(Tongji), "tju", "tongji-university":
		return Tongji, true
	case string(ZUEL), "zhongnan-university-of-economics-and-law":
		return ZUEL, true
	case string(NUS), "national-university-of-singapore":
		return NUS, true
	case string(NTU), "nanyang-technological-university":
		return NTU, true
	case string(ShanghaiJiao), "sjtu", "sjdu", "shanghai-jiao-tong-university":
		return ShanghaiJiao, true
	case string(Fudan), "fdu", "fudan-university":
		return Fudan, true
	default:
		return "", false
	}
}

func DisplayName(name Name) string {
	return displayNames[name]
}

func DisplayNameByEmail(email string) (string, bool) {
	name, ok := NameByEmail(email)
	if !ok {
		return "", false
	}
	displayName := DisplayName(name)
	if displayName == "" {
		return "", false
	}
	return displayName, true
}

func Key(name Name) string {
	return keys[name]
}

func KeyByEmail(email string) (string, bool) {
	name, ok := NameByEmail(email)
	if !ok {
		return "", false
	}
	key := Key(name)
	if key == "" {
		return "", false
	}
	return key, true
}
