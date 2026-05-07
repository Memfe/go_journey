package main

func main() {
	// myStack := Stack{}
	// fmt.Println(myStack)
	// myStack.Push(100)
	// myStack.Push(200)
	// myStack.Push(300)
	// myStack.Push(3)
	// myStack.Push(25)
	// fmt.Println(myStack)
	// myStack.QueuePop()
	// fmt.Println(myStack)
	// myStack.QueuePop()
	// fmt.Println(myStack)
	// fmt.Println(myStack.Len())
	// fmt.Println(myStack.IsEmpty())

	// root := &Node{Data: 25}
	// root.Insert(20)
	// root.Insert(2)
	// root.Insert(6)
	// root.Insert(14)
	// root.Insert(107)
	// root.Insert(11)
	// root.Insert(27)
	// root.Insert(33)
	// leftNode := &Node{Data: 20}
	// rightNode := &Node{Data: 16}
	// leftLeft := &Node{Data: 12}
	// root.LeftChild = leftNode
	// root.RightChild = rightNode
	// root.LeftChild.LeftChild = leftLeft
	// fmt.Println("Data:", root.Data)
	// fmt.Println("Left child:", root.LeftChild.Data)
	// fmt.Println("Right child:", root.RightChild.Data)
	// fmt.Println("Left left child:", root.LeftChild.LeftChild.Data)
	// root.PreOder()
	// fmt.Println()
	// root.InOrder()
	// fmt.Println()
	// root.PostOrder()
	// fmt.Println()
	newTag := TagItems{}
	newTag.AddTag("go", "backend")
	newTag.AddTag("go", "desktop")
	newTag.AddTag("rust", "backend")
	newTag.AddTag("rust", "desktop")
	newTag.AddTag("rust", "systems")
	newTag.AddTag("go", "cli")
	newTag.AddTag("c", "systems")
	newTag.AddTag("cpp", "systems")
	newTag.AddTag("c", "cli")
	newTag.AddTag("cpp", "games")
	newTag.View()
}
