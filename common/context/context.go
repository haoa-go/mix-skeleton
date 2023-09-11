package context

type LogContextInterface interface {
	GetLogArgs() []any
}

type RunContext struct {
	keys      map[string]any
	logFields []string
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
	t.logFields = make([]string, 0)
}

func NewRunContext() *RunContext {
	return &RunContext{
		keys: make(map[string]any),
	}
}

func (t *RunContext) AddLogField(field string) {
	t.logFields = append(t.logFields, field)
}

func (t *RunContext) GetLogFields() []string {
	return t.logFields
}

func (t *RunContext) GetLogArgs() (args []any) {
	logFields := t.GetLogFields()
	if len(logFields) > 0 {
		for _, field := range logFields {
			if value, exists := t.Get(field); exists {
				args = append(args, field, value)
			}
		}
	}
	return
}
