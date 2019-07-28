package timer

import (
	"testing"
)

var (
	quit int64
)

func TestAddCrontab(t *testing.T) {
	count := 0
	AddCrontab("* * * * *", "test", func(d *int) {
		t.Logf("crontab every minute")
		*d += 1
		if *d == 2 {
			quit = 1
		}
	}, &count)
	check()
}

func TestCancel(t *testing.T) {
	var h Handle
	AddCrontab("* * * * *", "test", func() {
		t.Logf("crontab every minute 1")
	})

	h = AddCrontab("* * * * *", "test", func() {
		t.Logf("crontab every minute 2")
		h.Cancel()
	})
	check()
}
