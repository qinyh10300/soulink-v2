package meta

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"soulink-backend/internal/school"
)

type stubUserCountStore struct {
	counts map[school.Name]int
	err    error
	calls  []school.Name
}

func (s *stubUserCountStore) CountActiveUsersByEmailSuffixes(_ context.Context, _ *sql.Tx, suffixes []string) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	for _, schoolName := range school.Names() {
		schoolSuffixes := school.EmailSuffixes(schoolName)
		if len(suffixes) == 0 || len(schoolSuffixes) == 0 {
			continue
		}
		if suffixes[0] == schoolSuffixes[0] {
			s.calls = append(s.calls, schoolName)
			return s.counts[schoolName], nil
		}
	}
	return 0, nil
}

type stubRegisteredUserCountCache struct {
	counts     map[school.Name]int
	hasValue   map[school.Name]bool
	loadErr    error
	storeErr   error
	loadCalls  []school.Name
	storeCalls []school.Name
}

func (c *stubRegisteredUserCountCache) Load(_ context.Context, schoolName school.Name) (int, error) {
	c.loadCalls = append(c.loadCalls, schoolName)
	if c.loadErr != nil {
		return 0, c.loadErr
	}
	if !c.hasValue[schoolName] {
		return 0, ErrRegisteredUserCountCacheNotFound
	}
	return c.counts[schoolName], nil
}

func (c *stubRegisteredUserCountCache) Store(_ context.Context, schoolName school.Name, count int, _ time.Duration) error {
	c.storeCalls = append(c.storeCalls, schoolName)
	if c.storeErr != nil {
		return c.storeErr
	}
	if c.counts == nil {
		c.counts = make(map[school.Name]int)
	}
	if c.hasValue == nil {
		c.hasValue = make(map[school.Name]bool)
	}
	c.counts[schoolName] = count
	c.hasValue[schoolName] = true
	return nil
}

func (c *stubRegisteredUserCountCache) Delete(_ context.Context, schoolName school.Name) error {
	if c.hasValue != nil {
		delete(c.hasValue, schoolName)
	}
	if c.counts != nil {
		delete(c.counts, schoolName)
	}
	return nil
}

func TestServiceReturnsCopies(t *testing.T) {
	t.Parallel()

	service := NewService()
	deps := service.Departments()
	deps[0] = "modified"
	if service.Departments()[0] == "modified" {
		t.Fatal("Departments() should return a copy")
	}

	contacts := service.ContactTypes()
	contacts[0].Label = "modified"
	if service.ContactTypes()[0].Label == "modified" {
		t.Fatal("ContactTypes() should return a copy")
	}

	pkuDeps := service.DepartmentsForSchool(school.Peking)
	pkuDeps[0] = "modified"
	if service.DepartmentsForSchool(school.Peking)[0] == "modified" {
		t.Fatal("DepartmentsForSchool() should return a copy")
	}
}

func TestValidators(t *testing.T) {
	t.Parallel()

	if !IsValidDepartment("计算机系") {
		t.Fatal("IsValidDepartment() should accept known department")
	}
	if IsValidDepartment("未知学院") {
		t.Fatal("IsValidDepartment() should reject unknown department")
	}
	if !IsValidDepartmentForEmail("test@pku.edu.cn", "计算机学院") {
		t.Fatal("IsValidDepartmentForEmail() should accept PKU department for PKU email")
	}
	if IsValidDepartmentForEmail("test@pku.edu.cn", "计算机系") {
		t.Fatal("IsValidDepartmentForEmail() should reject THU department for PKU email")
	}
	if !IsValidGrade("本科三年级") {
		t.Fatal("IsValidGrade() should accept known grade")
	}
	if !IsValidContactType("wechat") {
		t.Fatal("IsValidContactType() should accept known contact type")
	}
	if IsValidContactType("telegram") {
		t.Fatal("IsValidContactType() should reject unknown contact type")
	}
}

func TestContactTypeLabel(t *testing.T) {
	t.Parallel()

	if got := ContactTypeLabel("wechat"); got != "微信" {
		t.Fatalf("ContactTypeLabel() = %q, want %q", got, "微信")
	}
	if got := ContactTypeLabel("telegram"); got != "telegram" {
		t.Fatalf("ContactTypeLabel() = %q, want %q", got, "telegram")
	}
}

func TestRegisteredUserCountUsesCache(t *testing.T) {
	t.Parallel()

	cache := &stubRegisteredUserCountCache{
		counts:   map[school.Name]int{school.Tsinghua: 12},
		hasValue: map[school.Name]bool{school.Tsinghua: true},
	}
	store := &stubUserCountStore{counts: map[school.Name]int{school.Tsinghua: 99}}
	service := NewService(WithUserCountStore(store), WithRegisteredUserCountCache(cache, time.Minute))

	count, err := service.RegisteredUserCount(context.Background())
	if err != nil {
		t.Fatalf("RegisteredUserCount() error: %v", err)
	}
	if count != 79 {
		t.Fatalf("RegisteredUserCount() = %d, want %d", count, 79)
	}
	if len(store.calls) != 0 {
		t.Fatalf("store.calls = %d, want %d", len(store.calls), 0)
	}
}

func TestRegisteredUserCountFallsBackToStoreAndWarmsCache(t *testing.T) {
	t.Parallel()

	cache := &stubRegisteredUserCountCache{}
	store := &stubUserCountStore{counts: map[school.Name]int{school.Tsinghua: 7}}
	service := NewService(WithUserCountStore(store), WithRegisteredUserCountCache(cache, time.Minute))

	count, err := service.RegisteredUserCount(context.Background())
	if err != nil {
		t.Fatalf("RegisteredUserCount() error: %v", err)
	}
	if count != 74 {
		t.Fatalf("RegisteredUserCount() = %d, want %d", count, 74)
	}
	if len(store.calls) != 1 {
		t.Fatalf("store.calls = %d, want %d", len(store.calls), 1)
	}
	if len(cache.storeCalls) != 1 {
		t.Fatalf("cache.storeCalls = %d, want %d", len(cache.storeCalls), 1)
	}
	if !cache.hasValue[school.Tsinghua] || cache.counts[school.Tsinghua] != 7 {
		t.Fatalf("cache not warmed correctly: hasValue=%v count=%d", cache.hasValue[school.Tsinghua], cache.counts[school.Tsinghua])
	}
}

func TestRefreshRegisteredUserCountStoresLatestValue(t *testing.T) {
	t.Parallel()

	cache := &stubRegisteredUserCountCache{}
	store := &stubUserCountStore{
		counts: map[school.Name]int{
			school.Tsinghua:      25,
			school.Peking:        9,
			school.BeijingNormal: 3,
			school.Renmin:        4,
			school.Communication: 5,
			school.Beiwai:        6,
			school.ShanghaiJiao:  7,
			school.Fudan:         8,
			school.NUS:           9,
			school.NTU:           10,
			school.ZUEL:          11,
		},
	}
	service := NewService(WithUserCountStore(store), WithRegisteredUserCountCache(cache, time.Minute))

	if err := service.RefreshRegisteredUserCount(context.Background()); err != nil {
		t.Fatalf("RefreshRegisteredUserCount() error: %v", err)
	}
	if cache.counts[school.Tsinghua] != 25 {
		t.Fatalf("cache.counts[tsinghua] = %d, want %d", cache.counts[school.Tsinghua], 25)
	}
	if cache.counts[school.Peking] != 9 {
		t.Fatalf("cache.counts[peking] = %d, want %d", cache.counts[school.Peking], 9)
	}
	if cache.counts[school.BeijingNormal] != 3 {
		t.Fatalf("cache.counts[beijing_normal] = %d, want %d", cache.counts[school.BeijingNormal], 3)
	}
	if cache.counts[school.Renmin] != 4 {
		t.Fatalf("cache.counts[renmin] = %d, want %d", cache.counts[school.Renmin], 4)
	}
	if cache.counts[school.Communication] != 5 {
		t.Fatalf("cache.counts[communication] = %d, want %d", cache.counts[school.Communication], 5)
	}
	if cache.counts[school.Beiwai] != 6 {
		t.Fatalf("cache.counts[beiwai] = %d, want %d", cache.counts[school.Beiwai], 6)
	}
	if cache.counts[school.ShanghaiJiao] != 7 {
		t.Fatalf("cache.counts[shanghai_jiao_tong] = %d, want %d", cache.counts[school.ShanghaiJiao], 7)
	}
	if cache.counts[school.Fudan] != 8 {
		t.Fatalf("cache.counts[fudan] = %d, want %d", cache.counts[school.Fudan], 8)
	}
	if cache.counts[school.NUS] != 9 {
		t.Fatalf("cache.counts[nus] = %d, want %d", cache.counts[school.NUS], 9)
	}
	if cache.counts[school.NTU] != 10 {
		t.Fatalf("cache.counts[ntu] = %d, want %d", cache.counts[school.NTU], 10)
	}
	if cache.counts[school.ZUEL] != 11 {
		t.Fatalf("cache.counts[zuel] = %d, want %d", cache.counts[school.ZUEL], 11)
	}
}

func TestRegisteredUserCountIgnoresCacheReadErrorAndFallsBackToStore(t *testing.T) {
	t.Parallel()

	cache := &stubRegisteredUserCountCache{loadErr: errors.New("redis unavailable")}
	store := &stubUserCountStore{counts: map[school.Name]int{school.Tsinghua: 31}}
	service := NewService(WithUserCountStore(store), WithRegisteredUserCountCache(cache, time.Minute))

	count, err := service.RegisteredUserCount(context.Background())
	if err != nil {
		t.Fatalf("RegisteredUserCount() error: %v", err)
	}
	if count != 98 {
		t.Fatalf("RegisteredUserCount() = %d, want %d", count, 98)
	}
}

func TestPekingUniversityRegisteredUserCount(t *testing.T) {
	t.Parallel()

	cache := &stubRegisteredUserCountCache{
		counts:   map[school.Name]int{school.Peking: 5},
		hasValue: map[school.Name]bool{school.Peking: true},
	}
	store := &stubUserCountStore{counts: map[school.Name]int{school.Peking: 12}}
	service := NewService(WithUserCountStore(store), WithRegisteredUserCountCache(cache, time.Minute))

	count, err := service.PekingUniversityRegisteredUserCount(context.Background())
	if err != nil {
		t.Fatalf("PekingUniversityRegisteredUserCount() error: %v", err)
	}
	if count != 72 {
		t.Fatalf("PekingUniversityRegisteredUserCount() = %d, want %d", count, 72)
	}
}

func TestAdditionalSchoolRegisteredUserCountOffsets(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		schoolName school.Name
		method     func(*Service, context.Context) (int, error)
		baseCount  int
		want       int
	}{
		{
			name:       "beijing normal",
			schoolName: school.BeijingNormal,
			method:     (*Service).BeijingNormalUniversityRegisteredUserCount,
			baseCount:  11,
			want:       11,
		},
		{
			name:       "renmin",
			schoolName: school.Renmin,
			method:     (*Service).RenminUniversityRegisteredUserCount,
			baseCount:  8,
			want:       8,
		},
		{
			name:       "communication",
			schoolName: school.Communication,
			method:     (*Service).CommunicationUniversityRegisteredUserCount,
			baseCount:  10,
			want:       10,
		},
		{
			name:       "beiwai",
			schoolName: school.Beiwai,
			method:     (*Service).BeijingForeignStudiesUniversityRegisteredUserCount,
			baseCount:  9,
			want:       9,
		},
		{
			name:       "shanghai jiao tong",
			schoolName: school.ShanghaiJiao,
			method:     (*Service).ShanghaiJiaoTongUniversityRegisteredUserCount,
			baseCount:  9,
			want:       10,
		},
		{
			name:       "fudan",
			schoolName: school.Fudan,
			method:     (*Service).FudanUniversityRegisteredUserCount,
			baseCount:  9,
			want:       11,
		},
		{
			name:       "nus",
			schoolName: school.NUS,
			method:     (*Service).NationalUniversityOfSingaporeRegisteredUserCount,
			baseCount:  9,
			want:       12,
		},
		{
			name:       "ntu",
			schoolName: school.NTU,
			method:     (*Service).NanyangTechnologicalUniversityRegisteredUserCount,
			baseCount:  9,
			want:       13,
		},
		{
			name:       "zuel",
			schoolName: school.ZUEL,
			method:     (*Service).ZhongnanUniversityOfEconomicsAndLawRegisteredUserCount,
			baseCount:  9,
			want:       14,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cache := &stubRegisteredUserCountCache{
				counts:   map[school.Name]int{tc.schoolName: tc.baseCount},
				hasValue: map[school.Name]bool{tc.schoolName: true},
			}
			service := NewService(
				WithUserCountStore(&stubUserCountStore{counts: map[school.Name]int{tc.schoolName: 100}}),
				WithRegisteredUserCountCache(cache, time.Minute),
			)

			count, err := tc.method(service, context.Background())
			if err != nil {
				t.Fatalf("%s RegisteredUserCount() error: %v", tc.name, err)
			}
			if count != tc.want {
				t.Fatalf("%s RegisteredUserCount() = %d, want %d", tc.name, count, tc.want)
			}
		})
	}
}
