package collection

type StackNode[T any] struct {
	value T
	next  *StackNode[T]
}

type Stack[T any] struct {
	head *StackNode[T]
}

func (stack *Stack[T]) Push(element T) {
	node := &StackNode[T]{value: element, next: stack.head}
	stack.head = node
}

func (stack *Stack[T]) Pop() T {
	if stack.head != nil {
		element := stack.head.value
		stack.head = stack.head.next
		return element
	}
	panic("stack is empty")
}
