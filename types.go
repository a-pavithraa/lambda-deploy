package main

//go:generate  stringer -type=ActionType
type ActionType int

const (
	Insert ActionType = iota
	Delete
	Update
)
