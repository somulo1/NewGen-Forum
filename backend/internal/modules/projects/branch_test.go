package projects

import "testing"

func TestBranchNameIsShortCleanAndKept(t *testing.T) {
	cases := map[[2]string]string{
		{"HUB-12", "Fix the login timeout on mobile!"}:                             "HUB-12_fix_login_timeout_mobile",
		{"PRJ-3", "Café: Add user-group access to the branch dashboard and links"}: "PRJ-3_cafe_add_user_group_access_branch",
		{"PRJ-4", "The"}: "PRJ-4_the",
		{"PRJ-5", "Supercalifragilisticexpialidocious onomatopoeia extraordinarily"}: "PRJ-5_supercalifragilisticexpialidocious",
	}
	for in, want := range cases {
		if got := BranchName(in[0], in[1]); got != want {
			t.Errorf("BranchName(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
	// made once; renaming the task keeps it; an edited one is checked
	d := map[string]any{"title": "Fix login"}
	if err := setBranch("HUB", 7, d, nil); err != nil || d["branch"] != "HUB-7_fix_login" {
		t.Fatalf("new task: %v %v", d["branch"], err)
	}
	renamed := map[string]any{"title": "Something else entirely"}
	if err := setBranch("HUB", 7, renamed, d); err != nil || renamed["branch"] != "HUB-7_fix_login" {
		t.Fatalf("renamed task: %v %v", renamed["branch"], err)
	}
	for _, bad := range []string{"-x", "a..b", "feature/", "x.lock", "has space"} {
		if err := setBranch("HUB", 7, map[string]any{"title": "x", "branch": bad}, nil); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
	ok := map[string]any{"title": "x", "branch": "feature/HUB-7_fix"}
	if err := setBranch("HUB", 7, ok, nil); err != nil || ok["branch"] != "feature/HUB-7_fix" {
		t.Fatalf("own branch name: %v %v", ok["branch"], err)
	}
}
