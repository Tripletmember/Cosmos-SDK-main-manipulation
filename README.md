# cosmos-sdk-crud

It uses a CRUD model to interact with objects inside KVStores. Using indexes it allows to get objects not only by their primary keys but also secondary fields, it allows hence to group set of objects and update them. Multi key searches are supported. 

Your type needs to implement the `types.Object` to interact with the crud store.


