package school

import "testing"

func TestNameByEmail(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		email string
		want  Name
	}{
		{email: "user@mails.tsinghua.edu.cn", want: Tsinghua},
		{email: "user@mail.tsinghua.edu.cn", want: Tsinghua},
		{email: "user@tsinghua.org.cn", want: Tsinghua},
		{email: "user@mail.tsinghua.org.cn", want: Tsinghua},
		{email: "user@tsinghua.edu.cn", want: Tsinghua},
		{email: "user@student.pumc.edu.cn", want: Tsinghua},
		{email: "user@pku.edu.cn", want: Peking},
		{email: "user@stu.pku.edu.cn", want: Peking},
		{email: "user@bimba.pku.edu.cn", want: Peking},
		{email: "user@mail.bnu.edu.cn", want: BeijingNormal},
		{email: "user@cuc.edu.cn", want: Communication},
		{email: "user@ruc.edu.cn", want: Renmin},
		{email: "user@mail.bfsu.edu.cn", want: Beiwai},
		{email: "user@bfsu.edu.cn", want: Beiwai},
		{email: "user@stu.zuel.edu.cn", want: ZUEL},
		{email: "user@u.nus.edu", want: NUS},
		{email: "user@e.ntu.edu.sg", want: NTU},
		{email: "user@m.fudan.edu.cn", want: Fudan},
		{email: "user@sjtu.edu.cn", want: ShanghaiJiao},
	}

	for _, tc := range testCases {
		got, ok := NameByEmail(tc.email)
		if !ok {
			t.Fatalf("NameByEmail(%q) reported unsupported email", tc.email)
		}
		if got != tc.want {
			t.Fatalf("NameByEmail(%q) = %q, want %q", tc.email, got, tc.want)
		}
	}
}

func TestDisplayNameByEmail(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		email string
		want  Name
	}{
		{email: "user@mails.tsinghua.edu.cn", want: Tsinghua},
		{email: "user@pku.edu.cn", want: Peking},
		{email: "user@mail.bnu.edu.cn", want: BeijingNormal},
		{email: "user@cuc.edu.cn", want: Communication},
		{email: "user@ruc.edu.cn", want: Renmin},
		{email: "user@mail.bfsu.edu.cn", want: Beiwai},
		{email: "user@bfsu.edu.cn", want: Beiwai},
		{email: "user@stu.zuel.edu.cn", want: ZUEL},
		{email: "user@u.nus.edu", want: NUS},
		{email: "user@e.ntu.edu.sg", want: NTU},
		{email: "user@m.fudan.edu.cn", want: Fudan},
		{email: "user@sjtu.edu.cn", want: ShanghaiJiao},
	}

	for _, tc := range testCases {
		got, ok := DisplayNameByEmail(tc.email)
		if !ok {
			t.Fatalf("DisplayNameByEmail(%q) reported unsupported email", tc.email)
		}
		if got != DisplayName(tc.want) {
			t.Fatalf("DisplayNameByEmail(%q) = %q, want %q", tc.email, got, DisplayName(tc.want))
		}
	}
}

func TestKeyByEmail(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		email string
		want  string
	}{
		{email: "user@mails.tsinghua.edu.cn", want: "thu"},
		{email: "user@pku.edu.cn", want: "pku"},
		{email: "user@mail.bnu.edu.cn", want: "bnu"},
		{email: "user@cuc.edu.cn", want: "cuc"},
		{email: "user@ruc.edu.cn", want: "ruc"},
		{email: "user@mail.bfsu.edu.cn", want: "bfsu"},
		{email: "user@bfsu.edu.cn", want: "bfsu"},
		{email: "user@stu.zuel.edu.cn", want: "zuel"},
		{email: "user@u.nus.edu", want: "nus"},
		{email: "user@e.ntu.edu.sg", want: "ntu"},
		{email: "user@m.fudan.edu.cn", want: "fdu"},
		{email: "user@sjtu.edu.cn", want: "sjtu"},
	}

	for _, tc := range testCases {
		got, ok := KeyByEmail(tc.email)
		if !ok {
			t.Fatalf("KeyByEmail(%q) reported unsupported email", tc.email)
		}
		if got != tc.want {
			t.Fatalf("KeyByEmail(%q) = %q, want %q", tc.email, got, tc.want)
		}
	}
}
