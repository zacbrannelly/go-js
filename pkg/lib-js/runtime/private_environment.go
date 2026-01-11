package runtime

type PrivateName struct {
	Description string
}

type PrivateEnvironment struct {
	OuterPrivateEnvironment *PrivateEnvironment
	Names                   []PrivateName
}

func NewPrivateEnvironment(outerPrivateEnvironment *PrivateEnvironment) *PrivateEnvironment {
	return &PrivateEnvironment{
		OuterPrivateEnvironment: outerPrivateEnvironment,
		Names:                   make([]PrivateName, 0),
	}
}
