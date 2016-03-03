package gomdies


type mnode struct {
	parent *mnode
	children []*mnode
	obj interface{}
}


func newMNode(obj interface{}) *mnode {
	return &mnode{
		children:make([]*mnode, 0),
		obj:obj,
	}
}

func (node *mnode) addChild(child *mnode) {
	if child == nil {
		return
	}

	node.children = append(node.children, child)
	child.parent = node
}

