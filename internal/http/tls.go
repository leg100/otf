package http

// SkipTLSVerification skips verification of certificates presented by a peer
// on a TLS encrypted connection.
//
// NOTE: this is a global variable because:
// (a) it's either something that turned on in every part of the running program
// (i.e. tests), or not at all.
// (b) it's used by many disparate parts of the code base and it would be a pain in the
// bum passing a local variable around everywhere.
var SkipTLSVerification bool
