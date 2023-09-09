package context

type RunContext struct {
	keys map[string]any
}

func (t *RunContext) Set(key string, value any) {
	t.keys[key] = value
}

func (t *RunContext) Get(key string) (value any, exists bool) {
	value, exists = t.keys[key]
	return
}

func (t *RunContext) Clear() {
	t.keys = make(map[string]any)
}

func NewRunContext() *RunContext {
	return &RunContext{
		keys: make(map[string]any),
	}
}
