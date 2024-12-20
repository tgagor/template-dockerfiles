package builder

type Stage int

const (
	Build Stage = iota
	Tag
	Push
	Remove
)

func (s Stage) String() string {
	return [...]string{"build", "tag", "push", "remove"}[s]
}
