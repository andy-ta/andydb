# AndyDB

Born out of necessity during a hackathon, AndyDB is a quick-and-easy REST-based NoSQL database inspired by MongoDB.

AndyDB is still in very early alpha and if you are not in my immediate friend circle then it is not ACID-compliant.

## Usage

You can either build the binary yourself with `go build` or [download](https://github.com/mockoon/mockoon/releases)
the release.

After starting `andydb.exe`, you can make simple RESTful CRUD requests to the server at `http://localhost:42069/api`.

For example, `curl -d '{"email": "andy@andy.db"}' http://localhost:42069/api/contacts` will create the contacts resource
type (since it does not exist yet) and will save the provided body as an entry of that resource. 
It will return the created object in JSON format with a new field `_id` (now a UUID v7) that can be used for future operations.
Subsequent POST requests to the contacts resource append each entry to that collection.

With the `_id` you may now perform a GET / PUT / DELETE requests in the format of:

- GET `http://localhost:42069/api/contacts/{_id}`
  - `curl http://localhost:42069/api/contacts/{_id}`
- PUT `http://localhost:42069/api/contacts/{_id}` 
  - `curl -X PUT -d '{"email": "db@andy.db"}' http://localhost:42069/api/contacts/{_id}`
- DELETE `http://localhost:42069/api/contacts/{_id}`
- `curl -X DELETE http://localhost:42069/api/contacts/{_id}`

If you don't provide the id for a GET request, it will return all entries of the resource.

## Development hints

- Run `ANDYDB_ENV=dev go run .` (or omit the env var when using `go run .`, it defaults to dev) so the server logs every request and emits the startup banner with the current mode.
- The server already enforces a 1 MB JSON body limit, rejects malformed bodies, and wraps every handler with recovery middleware, so panicked handlers just return 500s.
- The in-memory datastore now uses mutex guards plus UUID v7 keys and has race-tested coverage in `app/database/concurrency_test.go`.
- Graceful shutdown is wired through `signal.NotifyContext`, so `Ctrl+C` closes the HTTP listener cleanly.

## Upcoming Features

- Locking
- Save to file
- Security
- Containerization
- Type checking
- Cleanups by coding in Go better
- Etc.