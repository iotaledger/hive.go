package server

import (
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
			if got := b.Load(tt.args.peer); got != tt.want {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}


