package example

/*
import "github.bus.zalan.do/pathfinder/shop/registry"

const response = `
{
	id: "experiment:foo1",
	name: "foo",
	products: [{
		id: "product:foo:1",
		name: "product foo one",
		price: 42,
		stock: {
			id: "stock:foo:1",
			quantity: 3
		}
	}]
}
`

const modelMML = `
let collection define(
	fieldMapped("name", mapName)
	containsBy("id", "products", define(
		field("name")
		field("price")
		containsOneBy("id", "stock", define(
			field("quantity")
			queryOne(getProductStock)
		))
		query(getProductsByCollection)
	))
	queryOne("id", getCollectionByID)
)

// contains: automatically passes the object
// containsBy: allows parallelization by specifying the field in advance
// queryOneBy: needs to contain the spec for the id to know that it's a field available for parallelization
// fields: why defining them, if requiring mapping anyway?
// id included by default
`

Define(

	StringMapped("name", mapName),

	ContainsBy("id", "products", Define(

		String("name"),

		Int("price"),

		ContainsOneBy("stockId", "stock", Define(

			Int("quantity"),

			QueryOne(StockConnector.getProductStock),
		)),

		Query(ProductConnector.getProductsByCollection),
	)),

	QueryOne(CollectionConnector.getCollectionByID),
)
*/
