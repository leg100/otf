// Package mapper maintains mappings between various resource identifiers, which
// are used by upstream layers to make decisions and efficiently lookup
// resources.
//
// For instance, the authorization layer needs to decide whether to permit
// access and cannot do so based on a single identifier (e.g. a run id) but
// needs to know which organization and workspace id it relates to.
//
// Whereas the persistence layer, with access to mappings, need only lookup
// resources based on the most appropriate identifier for which it maintains an
// index, rather having to support lookups using a multitude of identifiers.
package mapper
