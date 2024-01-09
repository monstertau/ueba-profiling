package profile

type BloomFilterModel struct {
}

func (m *BloomFilterModel) Contains(key string) bool {
	return false
}

func (m *BloomFilterModel) Update(data string) {

}

func (m *BloomFilterModel) Rebuild(data []string) {

}
