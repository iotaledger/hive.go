package discover

import (
	"sync"
	"testing"
	"time"
)

func Test_pingFilter_blacklist(t *testing.T) {
	type fields struct {
		lastPing map[string]history
		RWMutex  sync.RWMutex
	}
	type args struct {
		peer string
	}

	lPing := map[string]history{}
	lPing["192.168.1.1"] = history{
		t:       time.Now(),
		counter: blacklistThreshold + 1,
	}
	lPing["192.168.1.2"] = history{
		t:       time.Now(),
		counter: blacklistThreshold - 3,
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test_pingfilter_blacklist_1",
			fields: fields{
				lastPing: lPing,
				RWMutex:  sync.RWMutex{},
			},
			args: args{"192.168.1.1"},
			want: true,
		},
		{
			name: "test_pingfilter_blacklist_2",
			fields: fields{
				lastPing: lPing,
				RWMutex:  sync.RWMutex{},
			},
			args: args{"192.168.1.2"},
			want: false,
		},
		{
			name: "test_pingfilter_blacklist_3",
			fields: fields{
				lastPing: lPing,
				RWMutex:  sync.RWMutex{},
			},
			args: args{"192.168.1.2"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &pingFilter{
				lastPing: tt.fields.lastPing,
				RWMutex:  tt.fields.RWMutex,
			}
			if got := p.blacklist(tt.args.peer); got != tt.want {
				t.Errorf("blacklist() = %v, want %v", got, tt.want)
			}
		})
	}
}
