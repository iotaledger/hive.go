# hive.go

hive.go is a Go library containing: data structures, various utils and 
abstractions which are used by both `GoShimmer` and `Hornet`.

#### Deadlock activation

Compile your program using the `deadlock` build flag in order to swap out
mutexes from the `syncutils` package with [https://github.com/sasha-s/go-deadlock](https://github.com/sasha-s/go-deadlock).