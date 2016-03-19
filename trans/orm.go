package trans


type RMNode interface {

	Parent() RMNode

	SetParent(parent RMNode)

	Child(index int) RMNode

	AddChildren(children ...RMNode)

	Size() int

	Value() interface{}

}