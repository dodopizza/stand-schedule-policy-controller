package state

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

func Test_GetSchedule(t *testing.T) {
	ps, err := NewPolicyState(
		&apis.SchedulesSpec{
			Startup: apis.CronSchedule{
				Cron:     "2 * * * *",
				Override: "",
			},
			Shutdown: apis.CronSchedule{
				Cron:     "1 * * * *",
				Override: "",
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	sh1, err := NewSchedule(apis.CronSchedule{Cron: "1 * * * *"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sh1, ps.GetSchedule(apis.StatusShutdown))

	sh2, err := NewSchedule(apis.CronSchedule{Cron: "2 * * * *"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sh2, ps.GetSchedule(apis.StatusStartup))
}

func Test_UpdateStatus(t *testing.T) {
	ts := time.Now().Round(time.Minute)
	ps, err := NewPolicyState(
		&apis.SchedulesSpec{
			Startup: apis.CronSchedule{
				Cron: "* * * * *",
			},
			Shutdown: apis.CronSchedule{
				Cron: "* * * * *",
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	ps.GetSchedule(apis.StatusStartup).SetFiredAfter(ts)
	ps.GetSchedule(apis.StatusShutdown).SetFiredAfter(ts)

	ps.UpdateStatus(apis.StatusStartup, ts.Add(time.Minute*1).Add(time.Second*20), nil)
	assert.Equal(t, ts.Add(time.Minute*1).Add(time.Second*20), ps.GetSchedule(apis.StatusStartup).completedAt)
	assert.Equal(t, time.Time{}, ps.GetSchedule(apis.StatusStartup).failedAt)

	ps.UpdateStatus(apis.StatusShutdown, ts.Add(time.Minute*1).Add(time.Second*20), errors.New("some"))
	assert.Equal(t, time.Time{}, ps.GetSchedule(apis.StatusShutdown).completedAt)
	assert.Equal(t, ts.Add(time.Minute*1).Add(time.Second*20), ps.GetSchedule(apis.StatusShutdown).failedAt)
}
