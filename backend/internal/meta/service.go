package meta

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/school"
)

type ContactTypeOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type UserCountStore interface {
	CountActiveUsersByEmailSuffixes(ctx context.Context, tx *sql.Tx, suffixes []string) (int, error)
}

type RegisteredUserCountCache interface {
	Load(ctx context.Context, schoolName school.Name) (int, error)
	Store(ctx context.Context, schoolName school.Name, count int, ttl time.Duration) error
	Delete(ctx context.Context, schoolName school.Name) error
}

var ErrRegisteredUserCountCacheNotFound = errors.New("meta: registered user count cache not found")

type Option func(*Service)

type Service struct {
	userCountStore           UserCountStore
	registeredUserCountCache RegisteredUserCountCache
	registeredUserCountTTL   time.Duration
}

var registeredUserCountOffsets = map[school.Name]int{
	school.Tsinghua:      67,
	school.Peking:        67,
	school.BeijingNormal: 0,
	school.Renmin:        0,
	school.Communication: 0,
	school.Beiwai:        0,
	school.USTB:          0,
	school.CSU:           0,
	school.Tongji:        0,
	school.ShanghaiJiao:  1,
	school.Fudan:         2,
	school.NUS:           3,
	school.NTU:           4,
	school.ZUEL:          5,
}

var departmentsBySchool = map[school.Name][]string{
	school.Tsinghua: {
		"城规系",
		"建筑系",
		"土木系",
		"水利系",
		"环境学院",
		"机械系",
		"精仪系",
		"能动系",
		"车辆学院",
		"工业工程系",
		"电机系",
		"电子系",
		"计算机系",
		"自动化系",
		"集成电路学院",
		"航院",
		"工物系",
		"化工系",
		"信息科学技术学院",
		"人工智能学院",
		"材料学院",
		"数学科学系",
		"物理系",
		"化学系",
		"生命学院",
		"地学系",
		"交叉信息研究院",
		"高研院",
		"经济管理学院",
		"公共管理学院",
		"五道口金融学院",
		"中文系",
		"外文系",
		"法学院",
		"新闻学院",
		"马克思主义学院",
		"人文学院",
		"社会科学学院",
		"美术学院",
		"统计系",
		"建管系",
		"天文系",
		"安全学院",
		"心理系",
		"卫健学院",
		"万科公共卫生与健康学院",
		"苏世民书院",
		"建筑技术科学系",
		"核研院",
		"教育学院",
		"药学院",
		"临床医学院",
		"软件学院",
		"网络研究院",
		"国际与地区研究院",
		"航发院",
		"新雅书院",
		"致理书院",
		"日新书院",
		"未央书院",
		"探微书院",
		"行健书院",
		"求真书院",
		"为先书院",
		"秀钟书院",
		"笃实书院",
		"至善书院",
		"无穹书院",
		"紫荆书院",
		"自强书院",
		"水木书院",
		"医学院",
		"生医工程学院",
		"深圳国际研究生院",
		"清华-伯克利深圳学院",
		"全球创新学院",
		"国家卓越工程师学院",
		"智能产业研究院",
		"碳中和研究院",
		"不方便透露",
	},
	school.Peking: {
		"数学科学学院",
		"物理学院",
		"化学与分子工程学院",
		"生命科学学院",
		"城市与环境学院",
		"地球与空间科学学院",
		"心理与认知科学学院",
		"建筑与景观设计学院",
		"信息科学技术学院",
		"计算机学院",
		"电子学院",
		"集成电路学院",
		"智能学院",
		"王选计算机研究所",
		"软件与微电子学院",
		"软件工程国家工程研究中心",
		"中国语言文学系",
		"历史学系",
		"考古文博学院",
		"哲学系（宗教学系）",
		"外国语学院",
		"艺术学院",
		"对外汉语教育学院",
		"歌剧研究院",
		"工学院",
		"力学与工程科学学院",
		"先进制造与机器人学院",
		"材料科学与工程学院",
		"未来技术学院",
		"环境科学与工程学院",
		"国际关系学院",
		"法学院",
		"信息管理系",
		"社会学系",
		"政府管理学院",
		"马克思主义学院",
		"教育学院",
		"新闻与传播学院",
		"体育教研部",
		"经济学院",
		"光华管理学院",
		"人口研究所",
		"国家发展研究院",
		"基础医学院",
		"药学院",
		"公共卫生学院",
		"护理学院",
		"医学人文学院",
		"医学继续教育学院",
		"第一医院",
		"人民医院",
		"第三医院",
		"口腔医院",
		"北京肿瘤医院",
		"第六医院",
		"深圳医院（共建）",
		"首钢医院（共建）",
		"国际医院（共建）",
		"滨海医院（共建）",
		"元培学院",
		"燕京学堂",
		"区域与国别研究院",
		"人工智能研究院",
		"碳中和研究院",
		"前沿交叉学科研究院",
		"北京国际数学研究中心",
		"现代农学院",
		"其它实体研究机构",
		"信息工程学院",
		"化学生物学与生物技术学院",
		"环境与能源学院",
		"城市规划与设计学院",
		"新材料学院",
		"科学智能学院",
		"汇丰商学院",
		"国际法学院",
		"人文社会科学学院",
	},
}

var grades = []string{
	"本科一年级",
	"本科二年级",
	"本科三年级",
	"本科四年级",
	"本科五年级及以上",
	"硕士一年级",
	"硕士二年级",
	"硕士三年级及以上",
	"博士一年级",
	"博士二年级",
	"博士三年级",
	"博士四年级",
	"博士五年级及以上",
	"已毕业（本科）",
	"已毕业（硕士）",
	"已毕业（博士）",
	"不方便透露",
}

var contactTypes = []ContactTypeOption{
	{Label: "微信", Value: "wechat"},
	{Label: "QQ", Value: "qq"},
	{Label: "手机号", Value: "phone"},
	{Label: "邮箱", Value: "email"},
	{Label: "其他", Value: "other"},
}

func WithUserCountStore(userCountStore UserCountStore) Option {
	return func(service *Service) {
		service.userCountStore = userCountStore
	}
}

func WithRegisteredUserCountCache(cache RegisteredUserCountCache, ttl time.Duration) Option {
	return func(service *Service) {
		service.registeredUserCountCache = cache
		if ttl > 0 {
			service.registeredUserCountTTL = ttl
		}
	}
}

func NewService(options ...Option) *Service {
	service := &Service{registeredUserCountTTL: 5 * time.Minute}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (s *Service) Departments() []string {
	return s.DepartmentsForSchool(school.Tsinghua)
}

func (s *Service) DepartmentsForSchool(schoolName school.Name) []string {
	return append([]string(nil), departmentsBySchool[schoolName]...)
}

func (s *Service) DepartmentsByEmail(email string) []string {
	schoolName, ok := school.NameByEmail(email)
	if !ok {
		return s.Departments()
	}
	return s.DepartmentsForSchool(schoolName)
}

func (s *Service) Grades() []string {
	return append([]string(nil), grades...)
}

func (s *Service) ContactTypes() []ContactTypeOption {
	return append([]ContactTypeOption(nil), contactTypes...)
}

func (s *Service) RegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.Tsinghua)
}

func (s *Service) PekingUniversityRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.Peking)
}

func (s *Service) BeijingNormalUniversityRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.BeijingNormal)
}

func (s *Service) RenminUniversityRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.Renmin)
}

func (s *Service) CommunicationUniversityRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.Communication)
}

func (s *Service) BeijingForeignStudiesUniversityRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.Beiwai)
}

func (s *Service) ShanghaiJiaoTongUniversityRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.ShanghaiJiao)
}

func (s *Service) FudanUniversityRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.Fudan)
}

func (s *Service) NationalUniversityOfSingaporeRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.NUS)
}

func (s *Service) NanyangTechnologicalUniversityRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.NTU)
}

func (s *Service) ZhongnanUniversityOfEconomicsAndLawRegisteredUserCount(ctx context.Context) (int, error) {
	return s.registeredUserCountBySchool(ctx, school.ZUEL)
}

func (s *Service) RefreshRegisteredUserCount(ctx context.Context) error {
	if s == nil || s.userCountStore == nil {
		return nil
	}
	if s.registeredUserCountCache == nil {
		return nil
	}
	for _, schoolName := range school.Names() {
		count, err := s.userCountStore.CountActiveUsersByEmailSuffixes(ctx, nil, school.EmailSuffixes(schoolName))
		if err != nil {
			return err
		}
		if err := s.registeredUserCountCache.Store(ctx, schoolName, count, s.registeredUserCountTTL); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) registeredUserCountBySchool(ctx context.Context, schoolName school.Name) (int, error) {
	if s != nil && s.registeredUserCountCache != nil {
		count, err := s.registeredUserCountCache.Load(ctx, schoolName)
		if err == nil {
			return count + registeredUserCountOffsetForSchool(schoolName), nil
		}
		if err != nil && !errors.Is(err, ErrRegisteredUserCountCacheNotFound) {
			// Fall back to the database when the cache is temporarily unavailable.
		}
	}
	if s == nil || s.userCountStore == nil {
		return 0, httpapi.FeatureNotReady("registered user count")
	}
	count, err := s.userCountStore.CountActiveUsersByEmailSuffixes(ctx, nil, school.EmailSuffixes(schoolName))
	if err != nil {
		return 0, httpapi.InternalError(err)
	}
	if s.registeredUserCountCache != nil {
		_ = s.registeredUserCountCache.Store(ctx, schoolName, count, s.registeredUserCountTTL)
	}
	return count + registeredUserCountOffsetForSchool(schoolName), nil
}

func registeredUserCountOffsetForSchool(schoolName school.Name) int {
	return registeredUserCountOffsets[schoolName]
}

func IsValidDepartment(value string) bool {
	for _, schoolName := range school.Names() {
		if IsValidDepartmentForSchool(schoolName, value) {
			return true
		}
	}
	return false
}

func IsValidDepartmentForSchool(schoolName school.Name, value string) bool {
	for _, item := range departmentsBySchool[schoolName] {
		if item == value {
			return true
		}
	}
	return false
}

func IsValidDepartmentForEmail(email, value string) bool {
	schoolName, ok := school.NameByEmail(email)
	if !ok {
		return false
	}
	return IsValidDepartmentForSchool(schoolName, value)
}

func IsValidGrade(value string) bool {
	for _, item := range grades {
		if item == value {
			return true
		}
	}
	return false
}

func IsValidContactType(value string) bool {
	for _, item := range contactTypes {
		if item.Value == value {
			return true
		}
	}
	return false
}

func ContactTypeLabel(value string) string {
	for _, item := range contactTypes {
		if item.Value == value {
			return item.Label
		}
	}
	return value
}
