package profile

type BloomFilterModel struct {
}

func (m *BloomFilterModel) Contains(entity string, attribute string) bool {
	return false
}

func (m *BloomFilterModel) Update(entity string, attributes map[string]int) {

}
