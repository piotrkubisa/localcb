package cmd

type FlagPair struct {
	Long, Short string
}

func NewFlagPair(l, s string) FlagPair {
	return FlagPair{l, s}
}

func (f FlagPair) Join() string {
	if len(f.Short) == 0 {
		return f.Long
	}

	return f.Long + ", " + f.Short
}
