package enum

type StrEnum interface {
	ItemsAmount() int
	IsValid() bool
	String() string
	Value() int
	StringByValue(value int) string
}

type StrEnumItem struct {
	Label string
	Value int
}

func AllStrEnums(strEnum StrEnum) []StrEnumItem {
	itemsAmount := strEnum.ItemsAmount()
	resp := make([]StrEnumItem, 0, itemsAmount-1)
	for i := 1; i <= itemsAmount; i++ {
		resp = append(resp, StrEnumItem{strEnum.StringByValue(i), i})
	}
	return resp
}
