package gomdies


type MNode struct {
	parent *MNode
	children []*MNode
	obj interface{}
}


func newMNode(obj interface{}) *MNode {
	return &MNode{
		children:make([]*MNode, 0),
		obj:obj,
	}
}

func (node *MNode) addChild(child *MNode) {
	if child == nil {
		return
	}

	node.children = append(node.children, child)
	child.parent = node
}

