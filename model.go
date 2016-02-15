package gomdies

type Model interface {
	GetModelId() string
	SetModelId(id string)
}