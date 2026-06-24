package domain

// Product は商品エンティティです。
type Product struct {
	id    ProductID
	Name  ProductName
	Price ProductPrice
}

// NewProduct は検証済みの VO から商品を生成します。
func NewProduct(id ProductID, name ProductName, price ProductPrice) (*Product, error) {
	return &Product{
		id:    id,
		Name:  name,
		Price: price,
	}, nil
}

// ID は商品IDを返します。生成後は変えません。
func (p *Product) ID() ProductID {
	return p.id
}
