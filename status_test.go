package userconfig

import "testing"

func TestAggregateStatus(t *testing.T) {
	table := []struct {
		Status1        Status
		Status2        Status
		ExpectedStatus Status
	}{
		{STATUS_DOWN, STATUS_DOWN, STATUS_DOWN},
		{STATUS_UP, STATUS_UP, STATUS_UP},
		{STATUS_UP, STATUS_STARTING, STATUS_STARTING},
		{STATUS_UP, STATUS_STOPPING, STATUS_STOPPING},
		{STATUS_DOWN, STATUS_STOPPING, STATUS_DOWN},
		{STATUS_FAILED, STATUS_STARTING, STATUS_FAILED},
		{STATUS_FAILED, STATUS_DOWN, STATUS_FAILED},
		{STATUS_DOWN, STATUS_UP, STATUS_DOWN},
	}

	for index, testEntry := range table {
		agg := AggregateStatus(testEntry.Status1, testEntry.Status2)

		if agg != testEntry.ExpectedStatus {
			t.Errorf("%02d: AggregateStatus(%v, %v) = %v, expected %v\n",
				index,
				testEntry.Status1,
				testEntry.Status2,
				agg,
				testEntry.ExpectedStatus,
			)
		}
	}
}
