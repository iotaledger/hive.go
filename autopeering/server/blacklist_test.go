package server

import (
	"math/rand"
	"reflect"
	"sync"
	"testing"
)

func Test_newBlacklist(t *testing.T) {
	tests := []struct {
		name string
		want *blacklist
	}{
		{
			name: "test_blacklist_new_1",
			want: &blacklist{
				list: map[string]bool{},
				RWMutex: sync.RWMutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newBlacklist(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newBlacklist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_blacklist_Add(t *testing.T) {
	type fields struct {
		list    map[string]bool
		RWMutex sync.RWMutex
	}
	type args struct {
		peer string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "test_blacklist_add_1",
			fields: fields{
				list: map[string]bool{},
				RWMutex: sync.RWMutex{},
			},
			args:   args{
				peer: "",
			},
			want: false,
		},
		{
			name:   "test_blacklist_add_2",
			fields: fields{
				list: map[string]bool{},
				RWMutex: sync.RWMutex{},
			},
			args:   args{
				peer: "192.168.1.21",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &blacklist{
				list:    tt.fields.list,
				RWMutex: tt.fields.RWMutex,
			}
			if got := b.Add(tt.args.peer); got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_blacklist_Load(t *testing.T) {
	type fields struct {
		list    map[string]bool
		RWMutex sync.RWMutex
	}
	type args struct {
		peer string
	}

	list := map[string]bool{}
	list["192.168.1.1"] = true
	list["192.168.1.2"] = false

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test_blacklist_load_1",
			fields: fields{
				list:    list,
				RWMutex: sync.RWMutex{},
			},
			args: args{
				peer: "192.168.1.3",
			},
			want: false,
		},
		{
			name: "test_blacklist_load_2",
			fields: fields{
				list:    list,
				RWMutex: sync.RWMutex{},
			},
			args: args{
				peer: "192.168.1.1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &blacklist{
				list:    tt.fields.list,
				RWMutex: tt.fields.RWMutex,
			}
			if got := b.PeerExist(tt.args.peer); got != tt.want {
				t.Errorf("PeerExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func randBool() bool {
	return rand.Float32() < 0.5
}

func Test_blacklist_PeerExist(t *testing.T) {
	type fields struct {
		list    map[string]bool
		RWMutex sync.RWMutex
	}
	type args struct {
		peer string
	}

	list := map[string]bool{}
	peerLists := []string{"192.168.1.213", "192.168.1.31","192.168.1.41"}
	peerTests := []string{"192.168.1.213", "192.168.1.32"}

	for _, item := range peerLists {
		list[item] = randBool()
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "test_blacklist_peerexist_1",
			fields: fields{
				list:    list,
				RWMutex: sync.RWMutex{},
			},
			args:   args{
				peer: peerTests[0],
			},
			want:   true,
		},
		{
			name:   "test_blacklist_peerexist_2",
			fields: fields{
				list:    list,
				RWMutex: sync.RWMutex{},
			},
			args:   args{
				peer: peerTests[1],
			},
			want:   false,
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &blacklist{
				list:    tt.fields.list,
				RWMutex: tt.fields.RWMutex,
			}
			if got := b.PeerExist(tt.args.peer); got != tt.want {
				t.Errorf("PeerExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkBlacklist_PeerExist(b *testing.B) {
	s := &Server{}
	s.blacklist = newBlacklist()

	for n := 0; n < b.N; n++ {
		s.blacklist.PeerExist("192.168.1.12:1500")
	}
}