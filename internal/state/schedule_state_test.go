package state

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

func Test_GetExecutedTime(t *testing.T) {
	ts := time.Now()
	schedule, err := NewSchedule(apis.CronSchedule{Cron: "* * * * *", Override: ts.Format(time.RFC3339)})
	if err != nil {
		t.Fatal(err)
	}

	schedule.SetCompleted(ts.Add(time.Hour * 2))
	assert.Equal(t, ts.Add(time.Hour*2), schedule.GetExecutedTime())

	schedule.SetFailed(ts.Add(time.Hour * 3))
	assert.Equal(t, ts.Add(time.Hour*3), schedule.GetExecutedTime())
}

func Test_GetNextExecutionTime(t *testing.T) {
	ts := time.Now().Round(time.Minute)
	schedule, err := NewSchedule(apis.CronSchedule{Cron: "* * * * *"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, ts.Add(time.Minute), schedule.GetNextExecutionTime(ts))

	scheduleOverride, err := NewSchedule(apis.CronSchedule{Cron: "@yearly", Override: ts.Add(time.Minute).Format(time.RFC3339)})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, ts.Add(time.Minute), scheduleOverride.GetNextExecutionTime(ts))
}

func Test_SetFiredSince(t *testing.T) {
	ts := time.Now().Round(time.Minute)
	schedule, err := NewSchedule(apis.CronSchedule{Cron: "* * * * *"})
	if err != nil {
		t.Fatal(err)
	}

	schedule.SetFiredAfter(ts)
	assert.Equal(t, ts.Add(time.Minute), schedule.fireAt)
	assert.Equal(t, time.Time{}, schedule.completedAt)
	assert.Equal(t, time.Time{}, schedule.failedAt)
}

func Test_SetFailed(t *testing.T) {
	ts := time.Now().Round(time.Minute)
	schedule, err := NewSchedule(apis.CronSchedule{Cron: "* * * * *"})
	if err != nil {
		t.Fatal(err)
	}

	schedule.SetFailed(ts)
	assert.Equal(t, time.Time{}, schedule.completedAt)
	assert.Equal(t, ts, schedule.failedAt)
}

func Test_SetCompleted(t *testing.T) {
	ts := time.Now().Round(time.Minute)
	schedule, err := NewSchedule(apis.CronSchedule{Cron: "* * * * *"})
	if err != nil {
		t.Fatal(err)
	}

	schedule.SetCompleted(ts)
	assert.Equal(t, ts, schedule.completedAt)
	assert.Equal(t, time.Time{}, schedule.failedAt)
}

func Test_GetConditions(t *testing.T) {
	ts := time.Now().Round(time.Minute)
	schedule, err := NewSchedule(apis.CronSchedule{Cron: "* * * * *"})
	if err != nil {
		t.Fatal(err)
	}

	schedule.SetFiredAfter(ts)
	assert.Equal(t, []apis.StatusCondition{
		{
			Type:               apis.ConditionScheduled,
			Status:             apis.StatusShutdown,
			LastTransitionTime: meta.NewTime(ts.Add(time.Minute)),
		},
	}, schedule.GetConditions(apis.StatusShutdown))

	schedule.SetFailed(ts.Add(time.Minute * 3))
	assert.Equal(t, []apis.StatusCondition{
		{
			Type:               apis.ConditionScheduled,
			Status:             apis.StatusShutdown,
			LastTransitionTime: meta.NewTime(ts.Add(time.Minute)),
		},
		{
			Type:               apis.ConditionFailed,
			Status:             apis.StatusShutdown,
			LastTransitionTime: meta.NewTime(ts.Add(time.Minute * 3)),
		},
	}, schedule.GetConditions(apis.StatusShutdown))

	schedule.SetCompleted(ts.Add(time.Minute * 10))
	assert.Equal(t, []apis.StatusCondition{
		{
			Type:               apis.ConditionScheduled,
			Status:             apis.StatusShutdown,
			LastTransitionTime: meta.NewTime(ts.Add(time.Minute)),
		},
		{
			Type:               apis.ConditionCompleted,
			Status:             apis.StatusShutdown,
			LastTransitionTime: meta.NewTime(ts.Add(time.Minute * 10)),
		},
	}, schedule.GetConditions(apis.StatusShutdown))
}

func Test_ScheduleRequiredCron(t *testing.T) {
	ts := time.Now().Round(time.Minute)
	schedule, err := NewSchedule(apis.CronSchedule{Cron: "* * * * *"})
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, schedule.ScheduleRequired(ts))

	schedule.SetFiredAfter(ts)
	assert.False(t, schedule.ScheduleRequired(ts))

	// next fire time will be after 2 mins since ts
	fire := ts.Add(time.Minute)

	// complete after 10 secs since fire
	completed := fire.Add(time.Second * 10)

	schedule.SetCompleted(completed)

	// no schedule required after fire + 10 + 10 secs
	assert.False(t, schedule.ScheduleRequired(completed.Add(time.Second*10)))

	// schedule required after fire + 10 + 30 secs (40 > 1 min / 2)
	assert.True(t, schedule.ScheduleRequired(completed.Add(time.Second*30)))
}

func Test_ScheduleRequiredOverride(t *testing.T) {
	ts := time.Now().Round(time.Minute)
	schedule, err := NewSchedule(
		apis.CronSchedule{
			Override: ts.Add(time.Minute * 2).Format(time.RFC3339),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, schedule.ScheduleRequired(ts))

	schedule.SetFiredAfter(ts)
	assert.False(t, schedule.ScheduleRequired(ts))

	// next fire time will be after 2 mins since ts
	fire := ts.Add(time.Minute * 2)

	// complete after 10 secs since fire
	completed := fire.Add(time.Second * 10)

	schedule.SetCompleted(completed)

	// no schedule required after fire
	assert.False(t, schedule.ScheduleRequired(completed.Add(time.Minute*1)))
}
