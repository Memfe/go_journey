package main

import "fmt"

type Node struct {
	Data       int
	LeftChild  *Node
	RightChild *Node
}

// type BineryTree struct {
// 	Root *Node
// }

// func PreOder(node *Node) {
// 	if node == nil {
// 		return
// 	}
// 	fmt.Print(node.Data, " ")
// 	PreOder(node.LeftChild)
// 	PreOder(node.RightChild)
// }

// func InOrder(node *Node) {
// 	if node == nil {
// 		return
// 	}
// 	InOrder(node.LeftChild)
// 	fmt.Print(node.Data, " ")
// 	InOrder(node.RightChild)
// }

func (n *Node) PreOder() {
	if n == nil {
		return
	}
	fmt.Print(n.Data, " ")
	n.LeftChild.PreOder()
	n.RightChild.PreOder()
}

func (n *Node) InOrder() {
	if n == nil {
		return
	}
	n.LeftChild.InOrder()
	fmt.Print(n.Data, " ")
	n.RightChild.InOrder()
}

func (n *Node) PostOrder() {
	if n == nil {
		return
	}
	n.LeftChild.PostOrder()
	n.RightChild.PostOrder()
	fmt.Print(n.Data, " ")
}

func (n *Node) Insert(value int) {
	queue := []*Node{n}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.LeftChild == nil {
			current.LeftChild = &Node{Data: value}
			return
		} else {
			queue = append(queue, current.LeftChild)
		}
		if current.RightChild == nil {
			current.RightChild = &Node{Data: value}
			return
		} else {
			queue = append(queue, current.RightChild)
		}
	}
}
