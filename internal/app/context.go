package app

type ContextID int

const (
	ContextMain ContextID = iota
	ContextChannelDetail
	ContextChannelEdit
	ContextChannelAdd
	ContextHelp
	ContextMore
)

type Context struct {
	ID   ContextID
	Data interface{}
}

type ContextStack struct {
	stack []Context
}

func NewContextStack(initial Context) *ContextStack {
	return &ContextStack{stack: []Context{initial}}
}

func (cs *ContextStack) Push(ctx Context) {
	cs.stack = append(cs.stack, ctx)
}

func (cs *ContextStack) Pop() Context {
	if len(cs.stack) <= 1 {
		return cs.stack[0]
	}
	ctx := cs.stack[len(cs.stack)-1]
	cs.stack = cs.stack[:len(cs.stack)-1]
	return ctx
}

func (cs *ContextStack) Current() Context {
	return cs.stack[len(cs.stack)-1]
}

func (cs *ContextStack) Len() int {
	return len(cs.stack)
}
