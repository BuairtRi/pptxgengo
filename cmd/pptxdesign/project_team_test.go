package main

import "testing"

func TestTeamCLIFlagContracts(t *testing.T) {
	for _, args := range [][]string{
		{}, {"unknown"}, {"inspect"}, {"patch", "--slide", "team"},
		{"inspect", "--slide", "team", "--apply=false"},
		{"inspect", "--slide", "team", "--patch", "patch.yaml"},
		{"inspect", "--slide", "team", "unexpected"},
	} {
		if e := runProjectTeam(args); e == nil {
			t.Fatalf("accepted invalid team command %v", args)
		}
	}
}
