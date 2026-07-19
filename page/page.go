package page

type Meta struct {
	Title        string
	Description  string
	Keywords     []string
	CanonicalURL string
	OGImage      string
	NoIndex      bool
}

type Page interface {
	GetTemplate() string
	GetLayout() string
	GetMeta() Meta
	GetData() any
}

type TypedPage[T any] struct {
	Template string
	Layout   string
	Meta     Meta
	Data     T
}

func New[T any](template string, data T) *TypedPage[T] {
	return &TypedPage[T]{
		Template: template,
		Data:     data,
	}
}

func (p *TypedPage[T]) WithLayout(layout string) *TypedPage[T] {
	p.Layout = layout
	return p
}

func (p *TypedPage[T]) WithMeta(meta Meta) *TypedPage[T] {
	p.Meta = meta
	return p
}

func (p *TypedPage[T]) GetTemplate() string { return p.Template }
func (p *TypedPage[T]) GetLayout() string   { return p.Layout }
func (p *TypedPage[T]) GetMeta() Meta       { return p.Meta }
func (p *TypedPage[T]) GetData() any        { return p.Data }
