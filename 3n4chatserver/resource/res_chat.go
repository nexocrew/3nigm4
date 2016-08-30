package resource

type ChatResource struct{}
type ChatCollection struct{}

func (c *ChatResource) Get() (int, error) {

	return 200, nil
}

func (c *ChatCollection) Get() (int, error) {
	return 200, nil
}
