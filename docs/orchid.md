# Orchid

*Orchid* aims to be a superset of the Kubernetes API server using PostgreSQL as data storage, 
offering a RESTFul API to manage user defined resources, automatic change event propagation, in 
addition to an API allowing transactional changes including multiple objects (disclaimer: the need 
for such API exists and is still to be evaluated).

There are two major components in this system: the API handler and the data storage. The API handler 
is responsible for ingesting each HTTP message and decoding details of the change (for example the 
action, such as create, update or delete), validate the change, and then proceed to perform the 
change in the data storage.

The data storage (perhaps a gross oversimplification at this point, perhaps split into narrower 
scopes later on) offers an API to interact with single records, which will be translated to 
PostgreSQL semantics, and executed in the database.

Queries, either retrieving a single or a subset of a particular resource type, are also part of the 
data storage contract.
