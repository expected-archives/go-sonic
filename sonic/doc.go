// Package sonic implements all methods to communicate with it and execute commands.
//
// Syntax terminology (from https://github.com/valeriansaliou/sonic/blob/master/PROTOCOL.md):
//
// - collection: index collection (ie. what you search in, eg. messages, products, etc.);
//
// - bucket: index bucket name (ie. user-specific search classifier in the collection if you have any eg. user-1,
//user-2, .., otherwise use a common bucket name eg. generic, default, common, ..);
//
// - terms: text for search terms (between quotes);
//
// - count: a positive integer number; set within allowed maximum & minimum limits;
//
// - object: object identifier that refers to an entity in an external database, where the searched object is stored
// (eg. you use Sonic to index CRM contacts by name; full CRM contact data is stored in a MySQL database;
// in this case the object identifier in Sonic will be the MySQL primary key for the CRM contact);
//
// action: action to be triggered (available actions: consolidate);
//
//
// Notice: the bucket terminology may confuse some Sonic users. As we are well-aware Sonic may be used in an environment
// where end-users may each hold their own search index in a given collection, we made it possible to manage per-end-user
// search indexes with bucket. If you only have a single index per collection (most Sonic users will), we advise you use
// a static generic name for your bucket, for instance: default.
package sonic
