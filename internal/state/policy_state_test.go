package state

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

func Test_GetSchedule(t *testing.T) {
	ps, err := NewPolicyState(
		&apis.StandSchedulePolicy{
			ObjectMeta: meta.ObjectMeta{
				Name: "test",
			},
			Spec: apis.StandSchedulePolicySpec{
				TargetNamespaceFilter: "namespace1",
				Schedule: apis.ScheduleSpec{
					Startup:  "2 * * * *",
					Shutdown: "1 * * * *",
				},
				Resources: apis.ResourcesSpec{},
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	sh1, err := NewSchedule("1 * * * *", "")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sh1, ps.GetSchedule(apis.StatusShutdown))

	sh2, err := NewSchedule("2 * * * *", "")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sh2, ps.GetSchedule(apis.StatusStartup))
}

func Test_UpdateStatus(t *testing.T) {
	ts := time.Now().Round(time.Minute)
	ps, err := NewPolicyState(
		&apis.StandSchedulePolicy{
			ObjectMeta: meta.ObjectMeta{
				Name: "test",
			},
			Spec: apis.StandSchedulePolicySpec{
				TargetNamespaceFilter: "namespace1",
				Schedule: apis.ScheduleSpec{
					Startup:  "* * * * *",
					Shutdown: "* * * * *",
				},
				Resources: apis.ResourcesSpec{},
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	ps.GetSchedule(apis.StatusStartup).SetFiredAfter(ts)
	ps.GetSchedule(apis.StatusShutdown).SetFiredAfter(ts)

	ps.UpdateStatus(ts.Add(time.Minute*1).Add(time.Second*20), nil, apis.StatusStartup)
	assert.Equal(t, ts.Add(time.Minute*1).Add(time.Second*20), ps.GetSchedule(apis.StatusStartup).completedAt)
	assert.Equal(t, time.Time{}, ps.GetSchedule(apis.StatusStartup).failedAt)

	ps.UpdateStatus(ts.Add(time.Minute*1).Add(time.Second*20), errors.New("some"), apis.StatusShutdown)
	assert.Equal(t, time.Time{}, ps.GetSchedule(apis.StatusShutdown).completedAt)
	assert.Equal(t, ts.Add(time.Minute*1).Add(time.Second*20), ps.GetSchedule(apis.StatusShutdown).failedAt)
}
