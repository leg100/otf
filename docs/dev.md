# Development

Many tests, including unit tests, require access to a postgres database. A dedicated database should assigned for this purpose. Tests expect the environment variable `OTF_TEST_DATABASE_URL` to contain a valid postgres connection string.
