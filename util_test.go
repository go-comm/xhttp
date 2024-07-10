package xhttp

import "testing"

func Test_Join(t *testing.T) {
	testers := []struct {
		NArg int
		Arg1 string
		Arg2 string
		Arg3 string
		Want string
	}{
		{
			NArg: 3,
			Arg1: "http://a.com/",
			Arg2: "b",
			Arg3: "c",
			Want: "http://a.com/b/c",
		},
		{
			NArg: 2,
			Arg1: "http://a.com/",
			Arg2: "/b/",
			Want: "http://a.com/b/",
		},
		{
			NArg: 2,
			Arg1: "http://a.com",
			Arg2: "b/c",
			Want: "http://a.com/b/c",
		},
	}

	for _, tester := range testers {
		var got string
		switch tester.NArg {
		case 1:
			got = Join(tester.Arg1)
		case 2:
			got = Join(tester.Arg1, tester.Arg2)
		case 3:
			got = Join(tester.Arg1, tester.Arg2, tester.Arg3)
		}
		if got != tester.Want {
			t.Fatalf("arg1 %v,arg2 %v,arg3 %v  want %v, got %v", tester.Arg1, tester.Arg2, tester.Arg3, tester.Want, got)
		}
	}
}
